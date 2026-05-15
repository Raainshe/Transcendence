import { defineStore } from 'pinia'
import { ref, shallowRef } from 'vue'

import { Engine } from '@/game/engine/Engine'
import { InputController } from '@/game/input/InputController'
import { KeyboardAdapter } from '@/game/input/KeyboardAdapter'
import { toMatchRecordV1 } from '@/game/scoring/export'
import type { MatchRecordV1, ScoreBreakdown } from '@/game/scoring/types'
import { EnginePhase, type GameOverReason, type PieceType } from '@/game/types'
import { useGameSettingsStore } from '@/stores/gameSettings'

/** Max delta per frame to avoid huge jumps after tab backgrounding. */
const MAX_FRAME_DT_MS = 100

function phaseLabel(phase: EnginePhase): string {
  switch (phase) {
    case EnginePhase.Generation:
      return 'GENERATION'
    case EnginePhase.Falling:
      return 'FALLING'
    case EnginePhase.Lock:
      return 'LOCK'
    case EnginePhase.GameOver:
      return 'GAME OVER'
    default:
      return String(phase)
  }
}

function defaultSessionSeed(): number {
  return (Math.floor(Math.random() * 2 ** 31) ^ Date.now()) >>> 0 || 1
}

export const useGameSessionStore = defineStore('gameSession', () => {
  const engine = shallowRef<Engine | null>(null)
  let input: InputController | null = null
  let keyboard: KeyboardAdapter | null = null

  const active = ref(false)
  const paused = ref(false)

  const level = ref(1)
  const lines = ref(0)
  const goal = ref(10)
  const score = ref(0)
  const backToBackActive = ref(false)
  const backToBackCount = ref(0)
  const phaseLabelRef = ref('GENERATION')
  const gameOver = ref(false)
  const gameOverReason = ref<GameOverReason | undefined>(undefined)
  const nextPieces = ref<PieceType[]>([])
  const holdPiece = ref<PieceType | null>(null)
  const canHold = ref(true)

  const runId = ref('')
  const sessionSeed = ref(0)
  const startedAt = ref('')
  const endedAt = ref<string | undefined>(undefined)
  const scoreLedger = ref<ScoreBreakdown[]>([])
  const lastMatchRecord = ref<MatchRecordV1 | null>(null)

  function processEngineEvents(): void {
    const e = engine.value
    if (!e) return
    for (const event of e.drainEvents()) {
      if (event.type === 'score-awarded') {
        scoreLedger.value.push({ ...event.breakdown })
      }
    }
  }

  function syncHudFromEngine(): void {
    const e = engine.value
    if (!e) return
    const s = e.state
    level.value = s.level
    lines.value = s.lines
    goal.value = s.goal
    score.value = s.score
    backToBackActive.value = s.backToBackActive
    backToBackCount.value = s.backToBackCount
    phaseLabelRef.value = phaseLabel(s.phase)
    gameOver.value = s.gameOver
    gameOverReason.value = s.gameOverReason
    nextPieces.value = [...s.nextQueue]
    holdPiece.value = s.holdPiece
    canHold.value = s.canHold

    if (s.gameOver && !endedAt.value) {
      endedAt.value = new Date().toISOString()
      lastMatchRecord.value = buildMatchRecord()
    }
  }

  function buildMatchRecord(): MatchRecordV1 {
    const settings = useGameSettingsStore()
    const e = engine.value
    return toMatchRecordV1({
      runId: runId.value,
      seed: sessionSeed.value,
      variation: settings.variation,
      playerCount: settings.playerCount,
      startedAt: startedAt.value,
      endedAt: endedAt.value,
      final: {
        score: score.value,
        level: level.value,
        lines: lines.value,
        backToBackActive: backToBackActive.value,
        backToBackCount: backToBackCount.value,
      },
      events: scoreLedger.value,
    })
  }

  function pause(): void {
    if (paused.value) return
    paused.value = true
    input?.releaseAll()
    engine.value?.softDrop(false)
  }

  function resume(): void {
    paused.value = false
  }

  function beginSession(seed?: number): void {
    endSession()
    const s = seed ?? defaultSessionSeed()
    sessionSeed.value = s
    runId.value = crypto.randomUUID()
    startedAt.value = new Date().toISOString()
    endedAt.value = undefined
    scoreLedger.value = []
    lastMatchRecord.value = null

    const eng = new Engine({ seed: s })
    const ctrl = new InputController(eng, { isInputBlocked: () => paused.value })
    const kb = new KeyboardAdapter(ctrl)

    engine.value = eng
    input = ctrl
    keyboard = kb

    try {
      kb.attach()
    } catch {
      // No window (SSR / tests) — caller can still stepFrame manually.
    }

    active.value = true
    paused.value = false
    syncHudFromEngine()
  }

  function endSession(): void {
    if (active.value && !endedAt.value) {
      endedAt.value = new Date().toISOString()
      lastMatchRecord.value = buildMatchRecord()
    }

    keyboard?.detach()
    keyboard = null
    input = null
    engine.value = null
    active.value = false
    paused.value = false
    gameOver.value = false
    gameOverReason.value = undefined
    nextPieces.value = []
    holdPiece.value = null
    canHold.value = true
    score.value = 0
    backToBackActive.value = false
    backToBackCount.value = 0
    runId.value = ''
    sessionSeed.value = 0
    startedAt.value = ''
    endedAt.value = undefined
    scoreLedger.value = []
  }

  function stepFrame(dtMs: number): void {
    const e = engine.value
    if (!e || !input) return
    if (paused.value) {
      syncHudFromEngine()
      return
    }
    const dt = Math.min(Math.max(0, dtMs), MAX_FRAME_DT_MS)
    input.update(dt)
    e.update(dt)
    processEngineEvents()
    syncHudFromEngine()
  }

  return {
    engine,
    active,
    paused,
    level,
    lines,
    goal,
    score,
    backToBackActive,
    backToBackCount,
    phaseLabel: phaseLabelRef,
    gameOver,
    gameOverReason,
    nextPieces,
    holdPiece,
    canHold,
    runId,
    sessionSeed,
    startedAt,
    endedAt,
    scoreLedger,
    lastMatchRecord,
    beginSession,
    endSession,
    stepFrame,
    pause,
    resume,
    buildMatchRecord,
  }
})
