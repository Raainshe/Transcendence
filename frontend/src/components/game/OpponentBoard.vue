<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'

import { resolveAssetUrl } from '@/api/client'
import {
  decodeLockedMatrixBase64,
  MATRIX_CELL_COUNT,
} from '@/game/engine/matrixSnapshot'
import { drawLockedMatrix } from '@/game/render/drawMatrix'
import { MATRIX_VISIBLE_HEIGHT, MATRIX_WIDTH } from '@/game/types'

import '@/assets/styles/game/opponent-board.css'

const props = defineProps<{
  username: string
  avatarUrl: string | null
  score: number
  lines: number
  level: number
  alive: boolean
  connected?: boolean
  boardB64: string
}>()

const { t } = useI18n()

const CELL_PX = 8
const canvasRef = ref<HTMLCanvasElement | null>(null)

const displayName = computed(() => props.username || t('game.opponents.unknown'))
const avatarSrc = computed(() => resolveAssetUrl(props.avatarUrl))
const initials = computed(() => displayName.value.slice(0, 2).toUpperCase())
const waitingForBoard = computed(() => !props.boardB64)
const isDisconnected = computed(() => props.alive && props.connected === false)

function paint(): void {
  const canvas = canvasRef.value
  if (!canvas) return
  const ctx = canvas.getContext('2d')
  if (!ctx) return

  const grid = decodeLockedMatrixBase64(props.boardB64) ?? new Uint8Array(MATRIX_CELL_COUNT)
  drawLockedMatrix(ctx, grid, CELL_PX, true)
}

function resizeCanvas(): void {
  const canvas = canvasRef.value
  if (!canvas) return
  const dpr = Math.min(window.devicePixelRatio ?? 1, 2)
  const w = MATRIX_WIDTH * CELL_PX
  const h = MATRIX_VISIBLE_HEIGHT * CELL_PX
  canvas.style.width = `${w}px`
  canvas.style.height = `${h}px`
  canvas.width = Math.floor(w * dpr)
  canvas.height = Math.floor(h * dpr)
  const ctx = canvas.getContext('2d')
  if (ctx) ctx.setTransform(dpr, 0, 0, dpr, 0, 0)
  paint()
}

watch(
  () => [props.boardB64, props.alive] as const,
  () => {
    paint()
  },
)

onMounted(() => {
  resizeCanvas()
})
</script>

<template>
  <article class="opponent-board">
    <header class="opponent-board__header">
      <img
        v-if="avatarSrc"
        :src="avatarSrc"
        alt=""
        class="opponent-board__avatar"
      />
      <div v-else class="opponent-board__avatar opponent-board__avatar--fallback" aria-hidden="true">
        {{ initials }}
      </div>
      <div class="opponent-board__meta">
        <p class="opponent-board__name">{{ displayName }}</p>
        <p class="opponent-board__stats">
          {{ t('game.opponents.score', { score, lines, level }) }}
        </p>
      </div>
    </header>
    <div class="opponent-board__canvas-wrap">
      <canvas
        ref="canvasRef"
        class="opponent-board__canvas"
        :aria-label="t('game.opponents.boardAria', { name: displayName })"
      />
      <div v-if="isDisconnected" class="opponent-board__wait-overlay" role="status">
        {{ t('game.opponents.disconnected') }}
      </div>
      <div v-else-if="waitingForBoard" class="opponent-board__wait-overlay" role="status">
        {{ t('game.opponents.waiting') }}
      </div>
      <div v-else-if="!alive" class="opponent-board__out-overlay" role="status">
        {{ t('game.opponents.out') }}
      </div>
    </div>
  </article>
</template>
