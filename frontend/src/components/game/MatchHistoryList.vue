<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'

import * as gamesApi from '@/api/games'
import type { GameSummary } from '@/types/api'
import {
  GAME_VARIATIONS,
  GAME_VARIATION_LABELS,
  type GameVariation,
} from '@/types/game'

import '@/assets/styles/components/match-history.css'

const props = withDefaults(
  defineProps<{
    userId: string
    pageSize?: number
  }>(),
  { pageSize: 10 },
)

const emit = defineEmits<{
  select: [gameId: string]
}>()

const { t, locale } = useI18n()

const games = ref<GameSummary[]>([])
const total = ref(0)
const page = ref(0)
const loading = ref(true)
const loadError = ref(false)

const dateFormatter = computed(
  () =>
    new Intl.DateTimeFormat(locale.value, {
      dateStyle: 'medium',
      timeStyle: 'short',
    }),
)

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / props.pageSize)))
const canPrev = computed(() => page.value > 0)
const canNext = computed(() => (page.value + 1) * props.pageSize < total.value)

watch(
  () => props.userId,
  () => {
    page.value = 0
    void load()
  },
  { immediate: true },
)

watch(page, () => {
  void load()
})

async function load(): Promise<void> {
  if (!props.userId) return
  loading.value = true
  loadError.value = false
  try {
    const res = await gamesApi.listGames({
      userId: props.userId,
      limit: props.pageSize,
      offset: page.value * props.pageSize,
    })
    games.value = res.games
    total.value = res.total
  } catch {
    games.value = []
    total.value = 0
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

function formatWhen(game: GameSummary): string {
  const iso = game.finished_at ?? game.created_at
  const date = new Date(iso)
  if (Number.isNaN(date.getTime())) return '—'
  return dateFormatter.value.format(date)
}

function prevPage(): void {
  if (canPrev.value) page.value -= 1
}

function nextPage(): void {
  if (canNext.value) page.value += 1
}
</script>

<template>
  <div class="match-history">
    <p v-if="loading" class="match-history__status">{{ t('matchHistory.loading') }}</p>
    <p v-else-if="loadError" class="match-history__status">{{ t('matchHistory.loadError') }}</p>
    <p v-else-if="games.length === 0" class="match-history__status">
      {{ t('matchHistory.empty') }}
    </p>

    <template v-else>
      <ul class="match-history__list">
        <li v-for="game in games" :key="game.id">
          <button type="button" class="match-history__row" @click="emit('select', game.id)">
            <span class="match-history__mode">{{ modeLabel(game.mode) }}</span>
            <span class="match-history__date">{{ formatWhen(game) }}</span>
          </button>
        </li>
      </ul>

      <div v-if="total > pageSize" class="match-history__pagination">
        <button
          type="button"
          class="match-history__page-btn"
          :disabled="!canPrev"
          @click="prevPage"
        >
          {{ t('matchHistory.prev') }}
        </button>
        <span class="match-history__page-info">
          {{ t('matchHistory.page', { current: page + 1, total: totalPages }) }}
        </span>
        <button
          type="button"
          class="match-history__page-btn"
          :disabled="!canNext"
          @click="nextPage"
        >
          {{ t('matchHistory.next') }}
        </button>
      </div>
    </template>
  </div>
</template>
