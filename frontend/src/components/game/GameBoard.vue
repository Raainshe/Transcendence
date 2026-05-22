<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { storeToRefs } from 'pinia'

import { computeGhostPiece } from '@/game/engine/Ghost'
import { getOccupiedCells } from '@/game/engine/Tetrimino'
import {
  MATRIX_VISIBLE_HEIGHT,
  MATRIX_WIDTH,
  MINO_COLORS,
  MinoType,
  type PieceType,
} from '@/game/types'
import { LINE_CLEAR_FX_MS } from '@/game/fx/types'
import { useGameFxStore } from '@/stores/gameFx'
import { useGameSessionStore } from '@/stores/gameSession'

const MIN_CELL = 8
const MAX_CELL = 30

const store = useGameSessionStore()
const fxStore = useGameFxStore()
const { engine } = storeToRefs(store)

const canvasRef = ref<HTMLCanvasElement | null>(null)
const cellPx = ref(20)
let rafId = 0
let lastTs = 0
let resizeObserver: ResizeObserver | null = null

function colorForCell(value: number): string | null {
  if (value === MinoType.Empty) return null
  return MINO_COLORS[value as PieceType]
}

function rowFromTop(matrixY: number): number {
  return MATRIX_VISIBLE_HEIGHT - matrixY
}

function drawFx(ctx: CanvasRenderingContext2D, cell: number, now: number): void {
  fxStore.prune(now)
  const pad = 1
  const size = cell - pad * 2

  const trail = fxStore.hardDropTrail
  if (trail && trail.expiresAt > now) {
    const t = 1 - (trail.expiresAt - now) / 150
    const alpha = 0.55 * (1 - t)
    ctx.fillStyle = trail.color
    ctx.globalAlpha = alpha
    for (let x = trail.minX; x <= trail.maxX; x++) {
      const left = (x - 1) * cell + pad
      const topY = Math.max(trail.fromY, trail.toY)
      const bottomY = Math.min(trail.fromY, trail.toY)
      const top = rowFromTop(topY) * cell + pad
      const height = (topY - bottomY + 1) * cell - pad * 2
      if (height > 0) {
        ctx.fillRect(left, top, size, height)
      }
    }
    ctx.globalAlpha = 1
  }

  const flash = fxStore.lineClearFlash
  if (flash && flash.expiresAt > now) {
    const progress = 1 - (flash.expiresAt - now) / LINE_CLEAR_FX_MS
    const pulse = progress < 0.5 ? progress * 2 : 2 - progress * 2
    ctx.globalAlpha = 0.35 + pulse * 0.45
    for (const { x, y, color } of flash.clearedCells) {
      if (y < 1 || y > MATRIX_VISIBLE_HEIGHT) continue
      const left = (x - 1) * cell + pad
      const top = rowFromTop(y) * cell + pad
      ctx.fillStyle = color ? MINO_COLORS[color] : '#fff'
      ctx.fillRect(left, top, size, size)
      ctx.fillStyle = 'rgba(255, 255, 255, 0.9)'
      ctx.fillRect(left, top, size, size)
    }
    ctx.globalAlpha = 1
  }
}

function measureCellFromParent(): void {
  const canvas = canvasRef.value
  const parent = canvas?.parentElement
  if (!parent) return
  const r = parent.getBoundingClientRect()
  const w = Math.max(0, r.width)
  const h = Math.max(0, r.height)
  const byW = Math.floor(w / MATRIX_WIDTH)
  const byH = Math.floor(h / MATRIX_VISIBLE_HEIGHT)
  const next = Math.max(MIN_CELL, Math.min(MAX_CELL, Math.min(byW, byH)))
  if (next !== cellPx.value) cellPx.value = next
}

function resizeCanvas(): void {
  const canvas = canvasRef.value
  if (!canvas) return
  measureCellFromParent()
  const cell = cellPx.value
  const dpr = Math.min(window.devicePixelRatio ?? 1, 2)
  const w = MATRIX_WIDTH * cell
  const h = MATRIX_VISIBLE_HEIGHT * cell
  canvas.style.width = `${w}px`
  canvas.style.height = `${h}px`
  canvas.style.maxWidth = '100%'
  canvas.style.maxHeight = '100%'
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

  const cell = cellPx.value
  const w = MATRIX_WIDTH * cell
  const h = MATRIX_VISIBLE_HEIGHT * cell
  ctx.clearRect(0, 0, w, h)
  ctx.fillStyle = 'rgba(0, 0, 0, 0.45)'
  ctx.fillRect(0, 0, w, h)

  ctx.strokeStyle = 'rgba(255, 255, 255, 0.06)'
  ctx.lineWidth = 1
  for (let gx = 0; gx <= MATRIX_WIDTH; gx++) {
    ctx.beginPath()
    ctx.moveTo(gx * cell, 0)
    ctx.lineTo(gx * cell, h)
    ctx.stroke()
  }
  for (let gy = 0; gy <= MATRIX_VISIBLE_HEIGHT; gy++) {
    ctx.beginPath()
    ctx.moveTo(0, gy * cell)
    ctx.lineTo(w, gy * cell)
    ctx.stroke()
  }

  const pad = 1
  const matrix = eng.matrixRef
  for (let y = 1; y <= MATRIX_VISIBLE_HEIGHT; y++) {
    for (let x = 1; x <= MATRIX_WIDTH; x++) {
      const v = matrix.get(x, y)
      const c = colorForCell(v)
      if (!c) continue
      ctx.fillStyle = c
      ctx.fillRect((x - 1) * cell + pad, rowFromTop(y) * cell + pad, cell - pad * 2, cell - pad * 2)
    }
  }

  drawFx(ctx, cell, performance.now())

  const piece = eng.state.currentPiece
  if (piece) {
    const ghost = computeGhostPiece(matrix, piece)
    const sameCell =
      ghost.origin.x === piece.origin.x &&
      ghost.origin.y === piece.origin.y &&
      ghost.facing === piece.facing
    if (!sameCell) {
      const prevAlpha = ctx.globalAlpha
      ctx.globalAlpha = 0.3
      ctx.strokeStyle = 'rgba(255, 255, 255, 0.35)'
      for (const { x, y } of getOccupiedCells(ghost)) {
        if (y < 1 || y > MATRIX_VISIBLE_HEIGHT) continue
        const row = rowFromTop(y)
        ctx.fillStyle = MINO_COLORS[ghost.type]
        ctx.fillRect((x - 1) * cell + pad, row * cell + pad, cell - pad * 2, cell - pad * 2)
        ctx.strokeRect((x - 1) * cell + pad, row * cell + pad, cell - pad * 2, cell - pad * 2)
      }
      ctx.globalAlpha = prevAlpha
    }

    ctx.fillStyle = MINO_COLORS[piece.type]
    for (const { x, y } of getOccupiedCells(piece)) {
      if (y < 1 || y > MATRIX_VISIBLE_HEIGHT) continue
      ctx.fillRect((x - 1) * cell + pad, rowFromTop(y) * cell + pad, cell - pad * 2, cell - pad * 2)
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
  const canvas = canvasRef.value
  const parent = canvas?.parentElement
  if (parent) {
    resizeObserver = new ResizeObserver(() => {
      resizeCanvas()
    })
    resizeObserver.observe(parent)
  }
  window.addEventListener('resize', resizeCanvas)
  resizeCanvas()
  canvasRef.value?.focus()
  lastTs = 0
  rafId = requestAnimationFrame(loop)
})

onBeforeUnmount(() => {
  cancelAnimationFrame(rafId)
  resizeObserver?.disconnect()
  resizeObserver = null
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
  box-sizing: border-box;
  border: 2px solid var(--color-border);
  border-radius: var(--radius-sm);
  image-rendering: pixelated;
  max-width: 100%;
  max-height: 100%;
}
</style>
