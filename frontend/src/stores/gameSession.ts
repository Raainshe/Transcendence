import { defineStore } from 'pinia'
import { ref, shallowRef } from 'vue'

import { Engine } from '@/game/engine/Engine'
import { InputController } from '@/game/input/InputController'
import { KeyboardAdapter } from '@/game/input/KeyboardAdapter'
import { EnginePhase, type GameOverReason, type PieceType } from '@/game/types'

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

export const useGameSessionStore = defineStore('gameSession', () => {
  const engine = shallowRef<Engine | null>(null)
  let input: InputController | null = null
  let keyboard: KeyboardAdapter | null = null

  const active = ref(false)
  const paused = ref(false)

  const level = ref(1)
  const lines = ref(0)
  const goal = ref(10)
  const phaseLabelRef = ref('GENERATION')
  const gameOver = ref(false)
  const gameOverReason = ref<GameOverReason | undefined>(undefined)
  const nextPieces = ref<PieceType[]>([])
  const holdPiece = ref<PieceType | null>(null)
  const canHold = ref(true)

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
    nextPieces.value = [...s.nextQueue]
    holdPiece.value = s.holdPiece
    canHold.value = s.canHold
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
    const eng = new Engine(seed !== undefined ? { seed } : {})
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
    // Events reserved for slice 9 (audio / VFX); drain so the buffer stays bounded.
    e.drainEvents()
    syncHudFromEngine()
  }

  return {
    engine,
    active,
    paused,
    level,
    lines,
    goal,
    phaseLabel: phaseLabelRef,
    gameOver,
    gameOverReason,
    nextPieces,
    holdPiece,
    canHold,
    beginSession,
    endSession,
    stepFrame,
    pause,
    resume,
  }
})
