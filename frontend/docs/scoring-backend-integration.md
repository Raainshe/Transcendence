# Scoring system — backend integration guide

This document describes how the **frontend** produces match scoring data and how the **backend** should accept, store, and validate it. The client is authoritative for gameplay math during a session; the server stores an auditable record for leaderboards, stats, and optional replay verification later.

**Source of truth in code**

| Concern | Path |
|--------|------|
| JSON types (`MatchRecordV1`, `ScoreBreakdown`, …) | [`frontend/src/game/scoring/types.ts`](../src/game/scoring/types.ts) |
| Record builder | [`frontend/src/game/scoring/export.ts`](../src/game/scoring/export.ts) |
| Scoring rules (§8) | [`frontend/src/game/engine/Scoring.ts`](../src/game/engine/Scoring.ts) |
| T-Spin detection (§9) | [`frontend/src/game/engine/TSpin.ts`](../src/game/engine/TSpin.ts) |
| Session ledger + export timing | [`frontend/src/stores/gameSession.ts`](../src/stores/gameSession.ts) |

When the TypeScript types change, treat them as the contract and update this doc if field semantics change.

---

## Architecture (three layers)

```text
┌─────────────────────────────────────────────────────────────┐
│  Engine (authoritative)                                      │
│  - Computes score on each lock                               │
│  - Emits `score-awarded` events with ScoreBreakdown          │
└───────────────────────────┬─────────────────────────────────┘
                            │ drainEvents()
┌───────────────────────────▼─────────────────────────────────┐
│  gameSession (Pinia)                                         │
│  - Appends each breakdown to scoreLedger[]                     │
│  - Syncs running totals to HUD                               │
│  - buildMatchRecord() → MatchRecordV1 on game over / quit    │
└───────────────────────────┬─────────────────────────────────┘
                            │ POST JSON (future)
┌───────────────────────────▼─────────────────────────────────┐
│  Backend API                                                 │
│  - Persist match + ledger                                    │
│  - Leaderboards / user stats                                 │
└─────────────────────────────────────────────────────────────┘
```

**Important:** Vue components and the HUD **never** calculate score. They only display values synced from the engine / store.

---

## When a match record is produced

| Event | `endedAt` set? | `lastMatchRecord` |
|-------|----------------|-------------------|
| `beginSession()` | Cleared | `null` |
| Each frame with scoring | — | Ledger grows via `score-awarded` |
| `gameOver` (block-out, etc.) | Yes (ISO 8601) | `buildMatchRecord()` |
| `endSession()` (quit to menu) | Yes if still active | `buildMatchRecord()` |

**Frontend access (today):**

```ts
const store = useGameSessionStore()

// After game over or endSession:
const record = store.lastMatchRecord // MatchRecordV1 | null

// Or build manually while session metadata still exists:
const record = store.buildMatchRecord()
```

**Planned client submit (not implemented yet):**

```ts
if (store.lastMatchRecord) {
  await fetch('/api/matches', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(store.lastMatchRecord),
  })
}
```

Use **`runId`** as the idempotency key: one persisted row per `runId`, reject duplicates with `409 Conflict`.

---

## `MatchRecordV1` — request body shape

`schemaVersion` is fixed at **`1`** for this contract. Breaking changes require `schemaVersion: 2`.

### Top-level fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `schemaVersion` | `1` | yes | Contract version. |
| `runId` | `string` (UUID) | yes | Unique match id from `crypto.randomUUID()` at `beginSession`. |
| `seed` | `number` (uint32) | yes | PRNG seed for the 7-bag; same seed ⇒ same piece sequence. |
| `variation` | `string` | yes | Game mode from menu: `"marathon"` \| `"sprint"` \| `"ultra"` \| `"multiplayer"` (see [`types/game.ts`](../src/types/game.ts)). |
| `playerCount` | `number` | yes | `1`–`4` from menu settings. |
| `startedAt` | `string` | yes | ISO 8601 UTC when the session started. |
| `endedAt` | `string` | no | ISO 8601 UTC when the match ended (game over or quit). |
| `final` | `ScoreSnapshot` | yes | Totals at end of match (see below). |
| `events` | `ScoreBreakdown[]` | yes | Append-only scoring ledger for the whole match. |

### `ScoreSnapshot` (`final`)

| Field | Type | Description |
|-------|------|-------------|
| `score` | `number` | Total match score (sum of all `events[].pointsAwarded`). |
| `level` | `number` | Level at end (1–15, Marathon-style). |
| `lines` | `number` | Total lines cleared in the match. |
| `backToBackActive` | `boolean` | Whether a B2B bonus would apply on the *next* difficult clear. |
| `backToBackCount` | `number` | **Longest** B2B chain length reached this match (stat field). |

### `ScoreBreakdown` (`events[]`)

One entry per scoring action (usually 1–3 per piece lock: line clear, soft drop, hard drop).

| Field | Type | Description |
|-------|------|-------------|
| `sequence` | `number` | Monotonic per match, starting at 1. Use for ordering / replay alignment. |
| `reason` | `ScoreReason` | See enum below. |
| `level` | `number` | Level **at time of award** (affects multipliers). |
| `linesCleared` | `number` | Lines cleared by this action (`0` for drops / T-Spin no-lines). |
| `lineClearKind` | `LineClearKind` | `none` \| `single` \| `double` \| `triple` \| `tetris`. |
| `tSpinKind` | `TSpinKind` | `none` \| `mini` \| `full` (from lock-time detection §9). |
| `basePoints` | `number` | Table lookup **before** level and B2B (see scoring tables). |
| `backToBackMultiplier` | `number` | `1` or `1.5`. |
| `pointsAwarded` | `number` | `floor(basePoints × level × backToBackMultiplier)` for line clears; drops use `basePoints × level`. |
| `backToBackChain` | `number` | B2B chain length **after** this event (`0` = inactive). |

### String enums (stable in JSON)

**`ScoreReason`**

| Value | Meaning |
|-------|---------|
| `lineClear` | Standard or T-Spin line clear. |
| `softDrop` | Soft-drop cells for the piece that just locked. |
| `hardDrop` | Hard-drop cells for the piece that just locked. |
| `tSpinNoLines` | T-Spin Mini/Full with **zero** lines cleared. |

**`LineClearKind`:** `none`, `single`, `double`, `triple`, `tetris`

**`TSpinKind`:** `none`, `mini`, `full`

---

## Example payload

```json
{
  "schemaVersion": 1,
  "runId": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "seed": 42,
  "variation": "sprint",
  "playerCount": 1,
  "startedAt": "2026-05-15T12:00:00.000Z",
  "endedAt": "2026-05-15T12:08:32.500Z",
  "final": {
    "score": 12450,
    "level": 3,
    "lines": 22,
    "backToBackActive": false,
    "backToBackCount": 2
  },
  "events": [
    {
      "sequence": 1,
      "reason": "hardDrop",
      "level": 1,
      "linesCleared": 0,
      "lineClearKind": "none",
      "tSpinKind": "none",
      "basePoints": 38,
      "backToBackMultiplier": 1,
      "pointsAwarded": 38,
      "backToBackChain": 0
    },
    {
      "sequence": 2,
      "reason": "lineClear",
      "level": 1,
      "linesCleared": 1,
      "lineClearKind": "single",
      "tSpinKind": "none",
      "basePoints": 100,
      "backToBackMultiplier": 1,
      "pointsAwarded": 100,
      "backToBackChain": 0
    }
  ]
}
```

---

## Suggested database model

Normalize for querying; keep the full ledger for audit.

### `matches`

| Column | Type | Notes |
|--------|------|-------|
| `id` | UUID PK | Can equal `runId` from client. |
| `user_id` | FK | From auth (not sent by client today). |
| `seed` | BIGINT | |
| `variation` | VARCHAR | |
| `player_count` | SMALLINT | |
| `started_at` | TIMESTAMPTZ | |
| `ended_at` | TIMESTAMPTZ | |
| `score` | INTEGER | Copy of `final.score`. |
| `level` | SMALLINT | |
| `lines` | INTEGER | |
| `back_to_back_count` | SMALLINT | `final.backToBackCount`. |
| `schema_version` | SMALLINT | Default `1`. |
| `created_at` | TIMESTAMPTZ | Server insert time. |

### `match_score_events`

| Column | Type | Notes |
|--------|------|-------|
| `id` | BIGSERIAL PK | |
| `match_id` | UUID FK | |
| `sequence` | INTEGER | UNIQUE per match with `match_id`. |
| `reason` | VARCHAR | |
| `level` | SMALLINT | |
| `lines_cleared` | SMALLINT | |
| `line_clear_kind` | VARCHAR | |
| `t_spin_kind` | VARCHAR | |
| `base_points` | INTEGER | |
| `b2b_multiplier` | NUMERIC(2,1) | |
| `points_awarded` | INTEGER | |
| `b2b_chain` | SMALLINT | |

Optional: store raw `events` JSONB on `matches` for fast replay/debug, with normalized rows for analytics.

---

## Recommended API

### `POST /api/matches`

**Request:** `MatchRecordV1` JSON body (above).

**Responses**

| Status | When |
|--------|------|
| `201 Created` | New match stored; body may echo `{ "id": "<runId>" }`. |
| `409 Conflict` | `runId` already exists. |
| `400 Bad Request` | Validation failed (see below). |
| `401 Unauthorized` | No valid session (when auth exists). |

**Headers:** `Content-Type: application/json`

Associate `user_id` from the server session/JWT — do **not** trust a client-sent user id in the body for v1.

### Optional reads

- `GET /api/matches/:runId` — full record + events (admin / profile).
- `GET /api/leaderboard?variation=sprint` — aggregate `final.score` DESC, limit N.

---

## Server-side validation checklist

The client is trusted for v1, but these checks catch bugs and tampering:

1. **`schemaVersion === 1`**
2. **`runId`** is a valid UUID; unique in DB.
3. **`events`** is non-empty for any match with `final.score > 0` (unless you allow empty edge cases).
4. **`events[].sequence`** is strictly increasing `1..N` with no gaps or duplicates.
5. **`sum(events[].pointsAwarded) === final.score`** (integer equality).
6. **`final.lines`** should match sum of `linesCleared` on line-clear events only if you want strict consistency (note: `lines` on snapshot is total lines from engine, sum of per-lock line counts should match).
7. **Enum fields** ∈ allowed string sets above.
8. **`backToBackMultiplier`** ∈ `{ 1, 1.5 }`.
9. **`pointsAwarded`** recomputation (optional strict mode):
   - Line / T-Spin: `floor(basePoints × level × multiplier)` using tables below.
   - Drops: `basePoints × level` where `basePoints = cells × (1 soft \| 2 hard)`.
10. **`endedAt >= startedAt`** when both present.
11. **`seed`**, **`variation`**, **`playerCount`** within allowed ranges.

**Future anti-cheat:** replay inputs from `seed` + input log (not implemented). `seed` + `events` are the hooks for verification.

---

## Scoring rules reference (§8)

Implemented in [`Scoring.ts`](../src/game/engine/Scoring.ts). Base points are multiplied by **`level`** at award time.

### Standard line clears (× level)

| Kind | Base |
|------|------|
| Single | 100 |
| Double | 300 |
| Triple | 500 |
| Tetris | 800 |

### T-Spin line clears (× level)

| Lines | Full | Mini |
|-------|------|------|
| 1 | 800 | 200 |
| 2 | 1200 | 400 |
| 3 | 1600 | 600 |

### T-Spin, no lines (× level)

| Kind | Base |
|------|------|
| Full | 400 |
| Mini | 100 |

### Drop points (× level)

| Type | Per cell |
|------|----------|
| Soft drop | 1 |
| Hard drop | 2 |

Awarded once per lock for the piece that just landed (cells accumulated during that piece’s life).

### Back-to-Back (×1.5)

- **Difficult clear:** Tetris (4 lines) **or** any T-Spin that clears ≥ 1 line (Mini or Full).
- Second consecutive difficult clear (and each after): multiply that line-clear award by **1.5** (`backToBackMultiplier: 1.5`).
- Resets after a non-difficult line clear (e.g. plain Single/Double/Triple without T-Spin).
- T-Spin with **zero** lines does **not** advance or maintain B2B.
- Drop points never use B2B.

T-Spin **detection** (§9) runs at lock before row clear; see [`TSpin.ts`](../src/game/engine/TSpin.ts). Kick index **4** (SRS rotation point 5) ⇒ always **full** T-Spin when other T-Spin conditions hold.

---

## Multiple events per lock

On each piece lock the engine may emit **up to three** `score-awarded` events, in order:

1. Line clear / T-Spin (`reason`: `lineClear` or `tSpinNoLines`) — if applicable  
2. Soft drop (`softDrop`) — if `softDropCells > 0`  
3. Hard drop (`hardDrop`) — if `hardDropCells > 0`  

Each becomes one `events[]` row with its own `sequence`.

---

## Versioning and evolution

| Change | Action |
|--------|--------|
| New optional fields | Prefer adding to `MatchRecordV2` with `schemaVersion: 2`. |
| New `ScoreReason` values | Extend enum; old servers ignore unknown reasons or reject with 400. |
| Sprint/Ultra scoring (slice 8) | May add `variation`-specific tables; document per variation. |
| User id / multiplayer | Add server-side fields; client may send `runId` per player in team modes later. |

---

## Quick checklist for backend developers

- [ ] Accept `POST` body matching `MatchRecordV1`.
- [ ] Persist `runId` as unique match id.
- [ ] Store `events` for audit; index `(variation, final.score)` for leaderboards.
- [ ] Validate `sum(pointsAwarded) === final.score`.
- [ ] Attach `user_id` from auth, not from client body.
- [ ] Return `409` on duplicate `runId`.
- [ ] Keep TypeScript types in sync when frontend changes the contract.

For gameplay questions (T-Spin corners, B2B edge cases), refer to the Tetris Guideline §8–§9 and the engine tests under `frontend/src/game/engine/__tests__/`.
