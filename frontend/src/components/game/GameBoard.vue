<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { storeToRefs } from 'pinia'

import { getOccupiedCells } from '@/game/engine/Tetrimino'
import {
  MATRIX_VISIBLE_HEIGHT,
  MATRIX_WIDTH,
  MINO_COLORS,
  MinoType,
  type PieceType,
} from '@/game/types'
import { useGameSessionStore } from '@/stores/gameSession'

/** CSS pixels per cell (visible 10×20 playfield). */
const CELL = 28

const store = useGameSessionStore()
const { engine } = storeToRefs(store)

const canvasRef = ref<HTMLCanvasElement | null>(null)
let rafId = 0
let lastTs = 0

function colorForCell(value: number): string | null {
  if (value === MinoType.Empty) return null
  return MINO_COLORS[value as PieceType]
}

function resizeCanvas(): void {
  const canvas = canvasRef.value
  if (!canvas) return
  const dpr = Math.min(window.devicePixelRatio ?? 1, 2)
  const w = MATRIX_WIDTH * CELL
  const h = MATRIX_VISIBLE_HEIGHT * CELL
  canvas.style.width = `${w}px`
  canvas.style.height = `${h}px`
  canvas.style.maxWidth = '100%'
  canvas.width = Math.floor(w * dpr)
  canvas.height = Math.floor(h * dpr)
  const ctx = canvas.getContext('2d')
  if (ctx) {
    ctx.setTransform(dpr, 0, 0, dpr, 0, 0)
  }
}

function draw(): void {
  const canvas = canvasRef.value
  const eng = engine.value
  if (!canvas || !eng) return
  const ctx = canvas.getContext('2d')
  if (!ctx) return

  const w = MATRIX_WIDTH * CELL
  const h = MATRIX_VISIBLE_HEIGHT * CELL
  ctx.clearRect(0, 0, w, h)
  ctx.fillStyle = 'rgba(0, 0, 0, 0.45)'
  ctx.fillRect(0, 0, w, h)

  ctx.strokeStyle = 'rgba(255, 255, 255, 0.06)'
  ctx.lineWidth = 1
  for (let gx = 0; gx <= MATRIX_WIDTH; gx++) {
    ctx.beginPath()
    ctx.moveTo(gx * CELL, 0)
    ctx.lineTo(gx * CELL, h)
    ctx.stroke()
  }
  for (let gy = 0; gy <= MATRIX_VISIBLE_HEIGHT; gy++) {
    ctx.beginPath()
    ctx.moveTo(0, gy * CELL)
    ctx.lineTo(w, gy * CELL)
    ctx.stroke()
  }

  const pad = 1
  const matrix = eng.matrixRef
  for (let y = 1; y <= MATRIX_VISIBLE_HEIGHT; y++) {
    for (let x = 1; x <= MATRIX_WIDTH; x++) {
      const v = matrix.get(x, y)
      const c = colorForCell(v)
      if (!c) continue
      const rowFromTop = MATRIX_VISIBLE_HEIGHT - y
      ctx.fillStyle = c
      ctx.fillRect((x - 1) * CELL + pad, rowFromTop * CELL + pad, CELL - pad * 2, CELL - pad * 2)
    }
  }

  const piece = eng.state.currentPiece
  if (piece) {
    ctx.fillStyle = MINO_COLORS[piece.type]
    for (const { x, y } of getOccupiedCells(piece)) {
      if (y < 1 || y > MATRIX_VISIBLE_HEIGHT) continue
      const rowFromTop = MATRIX_VISIBLE_HEIGHT - y
      ctx.fillRect((x - 1) * CELL + pad, rowFromTop * CELL + pad, CELL - pad * 2, CELL - pad * 2)
    }
  }
}

function loop(ts: number): void {
  if (!store.active) return
  if (lastTs === 0) {
    lastTs = ts
  } else {
    const dt = ts - lastTs
    lastTs = ts
    store.stepFrame(dt)
  }
  draw()
  rafId = requestAnimationFrame(loop)
}

onMounted(() => {
  store.beginSession()
  resizeCanvas()
  window.addEventListener('resize', resizeCanvas)
  canvasRef.value?.focus()
  lastTs = 0
  rafId = requestAnimationFrame(loop)
})

onBeforeUnmount(() => {
  cancelAnimationFrame(rafId)
  window.removeEventListener('resize', resizeCanvas)
  store.endSession()
})
</script>

<template>
  <canvas
    ref="canvasRef"
    class="game-board"
    tabindex="0"
    role="application"
    aria-label="Tetris playfield"
  />
</template>

<style scoped>
.game-board {
  display: block;
  border: 2px solid var(--color-border);
  border-radius: var(--radius-sm);
  image-rendering: pixelated;
}
</style>
