<script setup lang="ts">
import { computed } from 'vue'

import GameModeBadge from '@/components/game/GameModeBadge.vue'
import HoldQueue from '@/components/game/HoldQueue.vue'
import NextQueue from '@/components/game/NextQueue.vue'
import type { GameOverReason, MatchWinReason } from '@/game/types'
import { useGameSessionStore } from '@/stores/gameSession'
import { SPRINT_LINE_GOAL } from '@/types/game'

defineProps<{
  /** `top`: stats row above the matrix. `bottom`: game-over + default slot (controls). */
  band: 'top' | 'bottom'
}>()

const store = useGameSessionStore()

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

const showTimer = computed(() => store.variant === 'sprint' || store.variant === 'ultra')

const failureLabel = computed(() => {
  const r: GameOverReason | undefined = store.gameOverReason
  if (r === 'blockOut') return 'BLOCK OUT'
  if (r === 'lockOut') return 'LOCK OUT'
  if (r === 'topOut') return 'TOP OUT'
  return ''
})

const winTitle = computed(() => {
  const r = store.matchEndReason as MatchWinReason | undefined
  if (r === 'sprintComplete') return 'Sprint complete!'
  if (r === 'ultraComplete') return "Time's up!"
  return 'Complete!'
})

const matchEnded = computed(() => store.matchEndKind !== 'playing')
</script>

<template>
  <header v-if="band === 'top'" class="game-hud game-hud--bar" aria-live="polite">
    <HoldQueue :hold-piece="store.holdPiece" :can-hold="store.canHold" />
    <GameModeBadge :variation="store.variant" size="sm" />
    <div class="game-hud__stat game-hud__stat--score">
      <span class="game-hud__label">Score</span>
      <span class="game-hud__value">{{ store.score.toLocaleString() }}</span>
      <span v-if="store.backToBackActive" class="game-hud__b2b">B2B</span>
    </div>
    <div v-if="showLevel" class="game-hud__stat">
      <span class="game-hud__label">Lvl</span>
      <span class="game-hud__value">{{ store.level }}</span>
    </div>
    <div class="game-hud__stat">
      <span class="game-hud__label">Lines</span>
      <span class="game-hud__value">{{ linesDisplay }}</span>
    </div>
    <div v-if="showTimer" class="game-hud__stat">
      <span class="game-hud__label">{{ store.variant === 'ultra' ? 'Time' : 'Timer' }}</span>
      <span class="game-hud__value">{{ timerLabel }}</span>
    </div>
    <div v-if="showPhase" class="game-hud__stat game-hud__stat--phase">
      <span class="game-hud__label">Phase</span>
      <span class="game-hud__value game-hud__value--sm">{{ store.phaseLabel }}</span>
    </div>
    <NextQueue :pieces="store.nextPieces" />
  </header>

  <footer v-else class="game-hud game-hud--footer" aria-live="polite">
    <div v-if="matchEnded" class="game-hud__game-over" role="status">
      <template v-if="store.matchEndKind === 'won'">
        <p class="game-hud__go-title game-hud__go-title--win">{{ winTitle }}</p>
        <p class="game-hud__go-score">Score: {{ store.score.toLocaleString() }}</p>
      </template>
      <template v-else>
        <p class="game-hud__go-title">Game over</p>
        <p v-if="failureLabel" class="game-hud__go-reason">{{ failureLabel }}</p>
        <p class="game-hud__go-score">Score: {{ store.score.toLocaleString() }}</p>
      </template>
      <p class="game-hud__go-hint">Esc: menu</p>
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
</style>
