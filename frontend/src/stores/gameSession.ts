import { defineStore } from 'pinia'
import { ref, shallowRef } from 'vue'

import { Engine } from '@/game/engine/Engine'
import { InputController } from '@/game/input/InputController'
import { KeyboardAdapter } from '@/game/input/KeyboardAdapter'
import {
  EnginePhase,
  type GameOverReason,
  MinoType,
  type PieceType,
} from '@/game/types'

/** Max delta per frame to avoid huge jumps after tab backgrounding. */
const MAX_FRAME_DT_MS = 100

function pieceTypeToLetter(t: PieceType): string {
  switch (t) {
    case MinoType.I:
      return 'I'
    case MinoType.O:
      return 'O'
    case MinoType.T:
      return 'T'
    case MinoType.S:
      return 'S'
    case MinoType.Z:
      return 'Z'
    case MinoType.J:
      return 'J'
    case MinoType.L:
      return 'L'
    default:
      return '?'
  }
}

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

export const useGameSessionStore = defineStore('gameSession', () => {
  const engine = shallowRef<Engine | null>(null)
  let input: InputController | null = null
  let keyboard: KeyboardAdapter | null = null

  const active = ref(false)

  const level = ref(1)
  const lines = ref(0)
  const goal = ref(10)
  const phaseLabelRef = ref('GENERATION')
  const gameOver = ref(false)
  const gameOverReason = ref<GameOverReason | undefined>(undefined)
  const nextQueueLabels = ref<string[]>([])

  function syncHudFromEngine(): void {
    const e = engine.value
    if (!e) return
    const s = e.state
    level.value = s.level
    lines.value = s.lines
    goal.value = s.goal
    phaseLabelRef.value = phaseLabel(s.phase)
    gameOver.value = s.gameOver
    gameOverReason.value = s.gameOverReason
    nextQueueLabels.value = s.nextQueue.map(pieceTypeToLetter)
  }

  function beginSession(seed?: number): void {
    endSession()
    const eng = new Engine(seed !== undefined ? { seed } : {})
    const ctrl = new InputController(eng)
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
    syncHudFromEngine()
  }

  function endSession(): void {
    keyboard?.detach()
    keyboard = null
    input = null
    engine.value = null
    active.value = false
    gameOver.value = false
    gameOverReason.value = undefined
    nextQueueLabels.value = []
  }

  function stepFrame(dtMs: number): void {
    const e = engine.value
    if (!e || !input) return
    const dt = Math.min(Math.max(0, dtMs), MAX_FRAME_DT_MS)
    input.update(dt)
    e.update(dt)
    // Events reserved for slice 9 (audio / VFX); drain so the buffer stays bounded.
    e.drainEvents()
    syncHudFromEngine()
  }

  return {
    engine,
    active,
    level,
    lines,
    goal,
    phaseLabel: phaseLabelRef,
    gameOver,
    gameOverReason,
    nextQueueLabels,
    beginSession,
    endSession,
    stepFrame,
  }
})
