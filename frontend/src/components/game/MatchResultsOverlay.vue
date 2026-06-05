<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'

import GameDetailModal from '@/components/game/GameDetailModal.vue'
import { useAuthStore } from '@/stores/auth'
import type { MatchEndedPayload } from '@/types/api'

import '@/assets/styles/game/match-results-overlay.css'

const props = defineProps<{
  gameId: string
  results: MatchEndedPayload
}>()

const router = useRouter()
const auth = useAuthStore()
const { t } = useI18n()

const showDetail = ref(false)

const selfId = computed(() => auth.user?.id?.trim().toLowerCase() ?? null)

const sortedPlayers = computed(() =>
  [...props.results.players].sort((a, b) => a.placement - b.placement),
)

const title = computed(() => {
  const winnerId = props.results.winner_user_id?.trim().toLowerCase() ?? null
  if (!winnerId) return t('game.matchResults.noWinner')
  if (selfId.value && winnerId === selfId.value) return t('game.matchResults.youWon')
  const winner = props.results.players.find(
    (p) => p.user_id.trim().toLowerCase() === winnerId,
  )
  if (winner) {
    return t('game.matchResults.winner', { name: winner.username })
  }
  return t('game.matchResults.matchOver')
})

const titleIsWin = computed(() => {
  const winnerId = props.results.winner_user_id?.trim().toLowerCase() ?? null
  return !!(selfId.value && winnerId && winnerId === selfId.value)
})

function isSelf(userId: string): boolean {
  return selfId.value === userId.trim().toLowerCase()
}

function goHome(): void {
  void router.push({ name: 'home' })
}
</script>

<template>
  <div class="match-results-overlay" role="dialog" aria-modal="true">
    <div class="match-results-overlay__panel">
      <h2
        class="match-results-overlay__title"
        :class="{ 'match-results-overlay__title--win': titleIsWin }"
      >
        {{ title }}
      </h2>
      <ol class="match-results-overlay__list">
        <li
          v-for="player in sortedPlayers"
          :key="player.user_id"
          class="match-results-overlay__row"
          :class="{
            'match-results-overlay__row--self': isSelf(player.user_id),
            'match-results-overlay__row--winner': player.is_winner,
          }"
        >
          <span class="match-results-overlay__place">#{{ player.placement }}</span>
          <span class="match-results-overlay__name">
            <RouterLink :to="{ name: 'userProfile', params: { id: player.user_id } }">
              {{ player.username }}
            </RouterLink>
            <span v-if="isSelf(player.user_id)"> {{ t('game.matchResults.you') }}</span>
          </span>
          <span class="match-results-overlay__stats">
            {{ t('game.opponents.score', { score: player.score, lines: player.lines, level: player.level }) }}
          </span>
        </li>
      </ol>
      <div class="match-results-overlay__actions">
        <button type="button" class="match-results-overlay__btn" @click="goHome">
          {{ t('game.matchResults.home') }}
        </button>
        <button
          type="button"
          class="match-results-overlay__btn match-results-overlay__btn--primary"
          @click="showDetail = true"
        >
          {{ t('game.matchResults.viewMatch') }}
        </button>
      </div>
    </div>
    <GameDetailModal
      :open="showDetail"
      :game-id="gameId"
      :highlight-user-id="auth.user?.id"
      @close="showDetail = false"
    />
  </div>
</template>
