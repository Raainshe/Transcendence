<script setup lang="ts">
import { ref, useId, watch } from 'vue'
import { useI18n } from 'vue-i18n'

import * as gamesApi from '@/api/games'
import type { GameDetail } from '@/types/api'
import {
  GAME_VARIATIONS,
  GAME_VARIATION_LABELS,
  type GameVariation,
} from '@/types/game'

import '@/assets/styles/auth/auth-modal.css'
import '@/assets/styles/components/match-history.css'

const props = defineProps<{
  open: boolean
  gameId: string | null
  highlightUserId?: string
  /** Shown for the highlighted player row (defaults to "you" on own profile). */
  highlightLabel?: string
}>()

const emit = defineEmits<{
  close: []
}>()

const { t, locale } = useI18n()

const titleId = useId()
const detail = ref<GameDetail | null>(null)
const loading = ref(false)
const loadError = ref(false)

watch(
  () => [props.open, props.gameId] as const,
  ([open, id]) => {
    if (open && id) {
      void loadDetail(id)
    } else {
      detail.value = null
      loadError.value = false
    }
  },
)

async function loadDetail(id: string): Promise<void> {
  loading.value = true
  loadError.value = false
  detail.value = null
  try {
    const { game } = await gamesApi.getGame(id)
    detail.value = game
  } catch {
    loadError.value = true
  } finally {
    loading.value = false
  }
}

function modeLabel(mode: string): string {
  if ((GAME_VARIATIONS as readonly string[]).includes(mode)) {
    return GAME_VARIATION_LABELS[mode as GameVariation]
  }
  return mode
}

function formatWhen(iso: string | null): string {
  if (!iso) return '—'
  const date = new Date(iso)
  if (Number.isNaN(date.getTime())) return '—'
  return new Intl.DateTimeFormat(locale.value, {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(date)
}

function formatScore(score: number): string {
  return score.toLocaleString(locale.value)
}

function onBackdropKeydown(event: KeyboardEvent): void {
  if (event.key === 'Escape' && !loading.value) {
    event.preventDefault()
    event.stopPropagation()
    emit('close')
  }
}

function isHighlighted(userId: string): boolean {
  return !!props.highlightUserId && props.highlightUserId === userId
}
</script>

<template>
  <div
    v-if="open"
    class="auth-modal-overlay"
    tabindex="-1"
    @keydown="onBackdropKeydown"
    @click.self="!loading && emit('close')"
  >
    <div
      class="auth-modal-panel auth-modal-panel--wide"
      role="dialog"
      aria-modal="true"
      :aria-labelledby="titleId"
      @click.stop
    >
      <h2 :id="titleId" class="auth-modal-panel__title">{{ t('matchHistory.detailTitle') }}</h2>

      <p v-if="loading" class="game-detail-modal__meta">{{ t('matchHistory.loading') }}</p>
      <p v-else-if="loadError" class="game-detail-modal__meta">{{ t('matchHistory.detailError') }}</p>

      <template v-else-if="detail">
        <p class="game-detail-modal__meta">
          {{ modeLabel(detail.mode) }} · {{ formatWhen(detail.finished_at ?? detail.created_at) }}
        </p>
        <table class="game-detail-modal__table">
          <thead>
            <tr>
              <th scope="col">{{ t('matchHistory.colPlayer') }}</th>
              <th scope="col">{{ t('matchHistory.colScore') }}</th>
              <th scope="col">{{ t('matchHistory.colLines') }}</th>
              <th scope="col">{{ t('matchHistory.colLevel') }}</th>
              <th scope="col">{{ t('matchHistory.colResult') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="player in detail.players"
              :key="player.id"
              :class="{ 'game-detail-modal__row--highlight': isHighlighted(player.user_id) }"
            >
              <td>
                <span v-if="isHighlighted(player.user_id)">
                  {{ highlightLabel ?? t('matchHistory.you') }}
                </span>
                <span v-else>{{ t('matchHistory.playerShort', { id: player.user_id.slice(0, 8) }) }}</span>
              </td>
              <td>{{ formatScore(player.score) }}</td>
              <td>{{ player.lines_cleared }}</td>
              <td>{{ player.level_reached }}</td>
              <td>
                <span v-if="player.is_winner" class="game-detail-modal__win">
                  {{ t('matchHistory.win') }}
                </span>
                <span v-else>—</span>
              </td>
            </tr>
          </tbody>
        </table>
      </template>

      <div class="auth-modal-panel__actions">
        <button type="button" class="auth-modal-panel__close" :disabled="loading" @click="emit('close')">
          {{ t('matchHistory.close') }}
        </button>
      </div>
    </div>
  </div>
</template>
