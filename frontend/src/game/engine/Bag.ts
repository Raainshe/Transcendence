/**
 * Random Tetrimino generation using the official Guideline "7-bag" system
 * (§3.3 / §A1.2.1).
 *
 * Each bag is the set `[I, O, T, S, Z, J, L]` shuffled into a random order
 * with Fisher-Yates; pieces are drawn from the front. A new bag is generated
 * the moment the previous one is exhausted, so within every 7 consecutive
 * draws each of the 7 Tetriminos appears exactly once.
 *
 * The PRNG (Mulberry32) is seedable. With a fixed seed the sequence is
 * deterministic, which we use for replays, multiplayer lockstep, and tests.
 * Mulberry32 is 32-bit, has a period of 2^32, passes the basic statistical
 * smoke tests we need for piece distribution, and fits in ~10 lines — small
 * enough to keep inline rather than pulling in a dependency.
 */

import { MinoType, type PieceType } from '@/game/types'

/** The full set of Tetriminos one bag contains, in canonical Guideline order (§3.1). */
const SEVEN_BAG: readonly PieceType[] = [
  MinoType.I,
  MinoType.O,
  MinoType.T,
  MinoType.S,
  MinoType.Z,
  MinoType.J,
  MinoType.L,
]

const BAG_SIZE = SEVEN_BAG.length

/**
 * Mulberry32 PRNG. Returns a function that yields a uniform `[0, 1)` float on
 * each call. The state is a single 32-bit unsigned integer, advanced in place
 * via captured-let closure (no globals).
 */
function mulberry32(seed: number): () => number {
  let s = seed >>> 0
  return () => {
    s = (s + 0x6d2b79f5) >>> 0
    let t = s
    t = Math.imul(t ^ (t >>> 15), t | 1)
    t ^= t + Math.imul(t ^ (t >>> 7), t | 61)
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296
  }
}

/** Produces a non-zero default seed from the current high-resolution time. */
function defaultSeed(): number {
  // performance.now() may be a float; force into a 32-bit unsigned. Date.now()
  // is added so two Bags created in the same animation frame still differ.
  const t =
    typeof performance !== 'undefined' && typeof performance.now === 'function'
      ? performance.now()
      : 0
  return ((Math.floor(t * 1000) ^ Date.now()) >>> 0) || 1
}

/**
 * A stateful Tetrimino generator. Cheap to construct (one Fisher-Yates over 7
 * elements per bag). Pure with respect to its constructor argument — same
 * seed always produces the same infinite stream of pieces.
 */
export class Bag {
  private rand: () => number
  /** Pieces remaining in the current bag, drawn from the front. */
  private current: PieceType[]
  /**
   * Pre-built next bag, populated on demand by `peek()` when a look-ahead
   * crosses a bag boundary. Always either empty or holds a freshly shuffled
   * full bag. `next()` rolls this into `current` rather than re-shuffling.
   */
  private lookahead: PieceType[] = []

  constructor(seed: number = defaultSeed()) {
    this.rand = mulberry32(seed)
    this.current = this.makeBag()
  }

  /** Returns the next piece in the stream, advancing the generator. */
  next(): PieceType {
    if (this.current.length === 0) this.refill()
    // shift() is safe here: bags are 7 elements, allocations are cheap and
    // rare (once per 7 draws). Avoids any index bookkeeping.
    const piece = this.current.shift()
    if (piece === undefined) {
      // Should be impossible given the refill above, but TypeScript's
      // `noUncheckedIndexedAccess` makes shift's `T | undefined` return
      // explicit. Treat as an unrecoverable invariant violation.
      throw new Error('Bag.next: current bag unexpectedly empty after refill')
    }
    return piece
  }

  /**
   * Returns the next `n` pieces WITHOUT consuming them. The same `Bag`
   * instance, queried twice with the same `n`, returns identical arrays.
   *
   * Used to drive the Next Queue display (§2.4): the engine can ask for the
   * next 6 pieces every frame without affecting the generator's state.
   */
  peek(n: number): PieceType[] {
    if (n <= 0) return []
    const out: PieceType[] = []
    for (let i = 0; i < n; i++) {
      if (i < this.current.length) {
        out.push(this.current[i] as PieceType)
        continue
      }
      const lookIdx = i - this.current.length
      if (lookIdx >= this.lookahead.length) {
        // Extend lookahead one bag at a time until we have enough.
        this.lookahead.push(...this.makeBag())
      }
      out.push(this.lookahead[lookIdx] as PieceType)
    }
    return out
  }

  /**
   * Promotes the lookahead (if any) into `current`. Building a fresh bag
   * only happens when no lookahead is queued.
   */
  private refill(): void {
    if (this.lookahead.length >= BAG_SIZE) {
      this.current = this.lookahead.splice(0, BAG_SIZE)
    } else {
      this.current = this.makeBag()
    }
  }

  /** Returns a freshly shuffled `[I, O, T, S, Z, J, L]` using Fisher-Yates. */
  private makeBag(): PieceType[] {
    const bag = SEVEN_BAG.slice()
    for (let i = bag.length - 1; i > 0; i--) {
      const j = Math.floor(this.rand() * (i + 1))
      const a = bag[i] as PieceType
      const b = bag[j] as PieceType
      bag[i] = b
      bag[j] = a
    }
    return bag
  }
}
