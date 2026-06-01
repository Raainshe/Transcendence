<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { RouterLink } from 'vue-router'

import { resolveAssetUrl } from '@/api/client'
import * as gamesApi from '@/api/games'
import type { LeaderboardEntry } from '@/types/api'
import {
  GAME_VARIATIONS,
  GAME_VARIATION_LABELS,
  type GameVariation,
} from '@/types/game'

import '@/assets/styles/views/leaderboard-view.css'

const LEADERBOARD_LIMIT = 50

const { t, locale } = useI18n()

const entries = ref<LeaderboardEntry[]>([])
const loading = ref(true)
const loadError = ref(false)

const dateFormatter = computed(
  () =>
    new Intl.DateTimeFormat(locale.value, {
      dateStyle: 'medium',
      timeStyle: 'short',
    }),
)

onMounted(() => {
  void loadLeaderboard()
})

async function loadLeaderboard(): Promise<void> {
  loading.value = true
  loadError.value = false
  try {
    entries.value = await gamesApi.getLeaderboard(LEADERBOARD_LIMIT)
  } catch {
    entries.value = []
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

function playerInitials(username: string): string {
  return username.slice(0, 2).toUpperCase()
}

function formatDate(iso: string | null): string {
  if (!iso) return '—'
  const date = new Date(iso)
  if (Number.isNaN(date.getTime())) return '—'
  return dateFormatter.value.format(date)
}

function formatScore(score: number): string {
  return score.toLocaleString(locale.value)
}
</script>

<template>
  <section class="leaderboard-view">
    <header class="leaderboard-view__header">
      <h1 class="leaderboard-view__title">{{ t('leaderboard.title') }}</h1>
      <p class="leaderboard-view__subtitle">{{ t('leaderboard.subtitle') }}</p>
    </header>

    <div class="leaderboard-view__panel">
      <p v-if="loading" class="leaderboard-view__status">{{ t('leaderboard.loading') }}</p>
      <p
        v-else-if="loadError"
        class="leaderboard-view__status leaderboard-view__status--error"
      >
        {{ t('leaderboard.loadError') }}
      </p>
      <p v-else-if="entries.length === 0" class="leaderboard-view__status">
        {{ t('leaderboard.empty') }}
      </p>

      <div v-else class="leaderboard-view__table-wrap">
        <table class="leaderboard-view__table">
          <thead>
            <tr>
              <th scope="col">{{ t('leaderboard.colRank') }}</th>
              <th scope="col">{{ t('leaderboard.colPlayer') }}</th>
              <th scope="col">{{ t('leaderboard.colScore') }}</th>
              <th scope="col">{{ t('leaderboard.colLines') }}</th>
              <th scope="col">{{ t('leaderboard.colLevel') }}</th>
              <th scope="col">{{ t('leaderboard.colMode') }}</th>
              <th scope="col">{{ t('leaderboard.colDate') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in entries" :key="`${row.user_id}-${row.rank}-${row.finished_at}`">
              <td class="leaderboard-view__rank">{{ row.rank }}</td>
              <td>
                <div class="leaderboard-view__player">
                  <img
                    v-if="resolveAssetUrl(row.avatar_url)"
                    :src="resolveAssetUrl(row.avatar_url)!"
                    :alt="t('leaderboard.avatarAlt', { username: row.username })"
                    class="leaderboard-view__avatar"
                    width="32"
                    height="32"
                    loading="lazy"
                  />
                  <span
                    v-else
                    class="leaderboard-view__avatar leaderboard-view__avatar--placeholder"
                    aria-hidden="true"
                  >
                    {{ playerInitials(row.username) }}
                  </span>
                  <RouterLink
                    :to="{ name: 'userProfile', params: { id: row.user_id } }"
                    class="leaderboard-view__username leaderboard-view__username--link"
                  >
                    {{ row.username }}
                  </RouterLink>
                </div>
              </td>
              <td class="leaderboard-view__score leaderboard-view__num">
                {{ formatScore(row.score) }}
              </td>
              <td class="leaderboard-view__num">{{ row.lines_cleared }}</td>
              <td class="leaderboard-view__num">{{ row.level_reached }}</td>
              <td class="leaderboard-view__mode">{{ modeLabel(row.mode) }}</td>
              <td class="leaderboard-view__date">{{ formatDate(row.finished_at) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </section>
</template>
