<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

import GameModeBadge from '@/components/game/GameModeBadge.vue'
import HoldQueue from '@/components/game/HoldQueue.vue'
import NextQueue from '@/components/game/NextQueue.vue'
import type { GameOverReason, MatchWinReason } from '@/game/types'
import { useGameSessionStore } from '@/stores/gameSession'
import { useMatchStore } from '@/stores/match'
import { SPRINT_LINE_GOAL } from '@/types/game'

defineProps<{
  /** `top`: stats row above the matrix. `bottom`: game-over + default slot (controls). */
  band: 'top' | 'bottom'
}>()

const store = useGameSessionStore()
const matchStore = useMatchStore()
const { t } = useI18n()

function formatMs(ms: number): string {
  const totalSec = Math.floor(ms / 1000)
  const m = Math.floor(totalSec / 60)
  const s = totalSec % 60
  return `${m}:${s.toString().padStart(2, '0')}`
}

const timerLabel = computed(() => {
  if (store.variant === 'ultra' && store.timeRemainingMs !== null) {
    return formatMs(store.timeRemainingMs)
  }
  return formatMs(store.elapsedMs)
})

const linesDisplay = computed(() => {
  if (store.variant === 'sprint') {
    return `${store.lines}/${SPRINT_LINE_GOAL}`
  }
  return `${store.lines}/${store.goal}`
})

const showLevel = computed(
  () => store.variant === 'marathon' || store.variant === 'ultra' || store.variant === 'multiplayer',
)

const showPhase = computed(() => store.variant === 'marathon' || store.variant === 'multiplayer')

const showLobbyName = computed(
  () => store.multiplayerGameId != null && matchStore.lobbyName !== '',
)

const showTimer = computed(() => store.variant === 'sprint' || store.variant === 'ultra')

const failureLabel = computed(() => {
  const r: GameOverReason | undefined = store.gameOverReason
  if (r === 'blockOut') return t('game.hud.blockOut')
  if (r === 'lockOut') return t('game.hud.lockOut')
  if (r === 'topOut') return t('game.hud.topOut')
  return ''
})

const winTitle = computed(() => {
  const r = store.matchEndReason as MatchWinReason | undefined
  if (r === 'sprintComplete') return t('game.hud.sprintComplete')
  if (r === 'ultraComplete') return t('game.hud.timesUp')
  return t('game.hud.complete')
})

const matchEnded = computed(() => store.matchEndKind !== 'playing')

const showMultiplayerWaiting = computed(
  () =>
    store.multiplayerGameId != null &&
    matchEnded.value &&
    !matchStore.matchEnded,
)

const showReconnecting = computed(
  () => store.multiplayerGameId != null && matchStore.reconnecting && !matchStore.matchEnded,
)

const showSoloGameOver = computed(
  () => matchEnded.value && !store.multiplayerGameId && !matchStore.matchEnded,
)
</script>

<template>
  <header v-if="band === 'top'" class="game-hud game-hud--bar" aria-live="polite">
    <HoldQueue :hold-piece="store.holdPiece" :can-hold="store.canHold" />
    <GameModeBadge :variation="store.variant" size="sm" />
    <div v-if="showLobbyName" class="game-hud__stat game-hud__stat--phase">
      <span class="game-hud__label">{{ t('game.hud.lobby') }}</span>
      <span class="game-hud__value game-hud__value--sm">{{ matchStore.lobbyName }}</span>
    </div>
    <div class="game-hud__stat game-hud__stat--score">
      <span class="game-hud__label">{{ t('game.hud.score') }}</span>
      <span class="game-hud__value">{{ store.score.toLocaleString() }}</span>
      <span v-if="store.backToBackActive" class="game-hud__b2b">{{ t('game.hud.b2b') }}</span>
    </div>
    <div v-if="showLevel" class="game-hud__stat">
      <span class="game-hud__label">{{ t('game.hud.level') }}</span>
      <span class="game-hud__value">{{ store.level }}</span>
    </div>
    <div class="game-hud__stat">
      <span class="game-hud__label">{{ t('game.hud.lines') }}</span>
      <span class="game-hud__value">{{ linesDisplay }}</span>
    </div>
    <div v-if="showTimer" class="game-hud__stat">
      <span class="game-hud__label">{{
        store.variant === 'ultra' ? t('game.hud.time') : t('game.hud.timer')
      }}</span>
      <span class="game-hud__value">{{ timerLabel }}</span>
    </div>
    <div v-if="showPhase" class="game-hud__stat game-hud__stat--phase">
      <span class="game-hud__label">{{ t('game.hud.phase') }}</span>
      <span class="game-hud__value game-hud__value--sm">{{ store.phaseLabel }}</span>
    </div>
    <NextQueue :pieces="store.nextPieces" />
    <p v-if="showReconnecting" class="game-hud__reconnect" role="status">
      {{ t('game.reconnecting') }}
    </p>
  </header>

  <footer v-else class="game-hud game-hud--footer" aria-live="polite">
    <div v-if="showMultiplayerWaiting" class="game-hud__game-over" role="status">
      <p class="game-hud__go-title">{{ t('game.matchResults.youOut') }}</p>
      <p v-if="failureLabel" class="game-hud__go-reason">{{ failureLabel }}</p>
      <p class="game-hud__go-hint">{{ t('game.matchResults.waiting') }}</p>
    </div>
    <div v-else-if="showSoloGameOver" class="game-hud__game-over" role="status">
      <template v-if="store.matchEndKind === 'won'">
        <p class="game-hud__go-title game-hud__go-title--win">{{ winTitle }}</p>
        <p class="game-hud__go-score">{{ t('game.hud.score') }}: {{ store.score.toLocaleString() }}</p>
      </template>
      <template v-else>
        <p class="game-hud__go-title">{{ t('game.hud.gameOver') }}</p>
        <p v-if="failureLabel" class="game-hud__go-reason">{{ failureLabel }}</p>
        <p class="game-hud__go-score">{{ t('game.hud.score') }}: {{ store.score.toLocaleString() }}</p>
      </template>
      <p class="game-hud__go-hint">{{ t('game.hud.menuHint') }}</p>
    </div>
    <div class="game-hud__slot">
      <slot />
    </div>
  </footer>
</template>

<style scoped>
.game-hud {
  font-family: var(--font-display);
  color: var(--color-text);
  background: var(--color-panel);
  border: 1px solid var(--color-border);
  box-sizing: border-box;
}

.game-hud--bar {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  justify-content: center;
  gap: var(--sp-3) var(--sp-4);
  padding: var(--sp-2) var(--sp-3);
  flex-shrink: 0;
  width: 100%;
  max-width: 100%;
  overflow: visible;
}

.game-hud__stat {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.125rem;
  min-width: 0;
}

.game-hud__stat--phase {
  max-width: 8rem;
}

.game-hud__stat--score {
  min-width: 3.5rem;
}

.game-hud__b2b {
  font-size: 0.45rem;
  color: var(--accent-selected);
  letter-spacing: 0.08em;
}

.game-hud__go-score {
  margin: 0 0 var(--sp-1);
  font-size: var(--fs-xs);
  color: var(--color-text);
}

.game-hud__label {
  font-size: 0.5rem;
  color: var(--color-text-dim);
  text-transform: uppercase;
  letter-spacing: 0.06em;
}

.game-hud__value {
  font-size: var(--fs-xs);
}

.game-hud__value--sm {
  font-size: 0.5rem;
  word-break: break-word;
  text-align: center;
}

.game-hud--footer {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--sp-2);
  padding: var(--sp-2) var(--sp-3);
  flex-shrink: 0;
  width: 100%;
}

.game-hud__slot {
  width: 100%;
  max-width: 36rem;
}

.game-hud__game-over {
  text-align: center;
}

.game-hud__go-title {
  margin: 0 0 var(--sp-1);
  font-size: var(--fs-xs);
  color: var(--t-red);
}

.game-hud__go-title--win {
  color: var(--accent-selected);
}

.game-hud__go-reason {
  margin: 0 0 var(--sp-1);
  font-size: 0.5rem;
  color: var(--color-text-dim);
}

.game-hud__go-hint {
  margin: 0;
  font-size: 0.5rem;
  color: var(--color-text-dim);
}

.game-hud__reconnect {
  margin: 0;
  width: 100%;
  text-align: center;
  font-size: 0.5rem;
  color: var(--t-amber, var(--color-text-dim));
  letter-spacing: 0.04em;
}
</style>
