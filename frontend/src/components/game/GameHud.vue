<script setup lang="ts">
import { computed } from 'vue'

import HoldQueue from '@/components/game/HoldQueue.vue'
import NextQueue from '@/components/game/NextQueue.vue'
import type { GameOverReason } from '@/game/types'
import { useGameSessionStore } from '@/stores/gameSession'

defineProps<{
  /** `top`: stats row above the matrix. `bottom`: game-over + default slot (controls). */
  band: 'top' | 'bottom'
}>()

const store = useGameSessionStore()

const reasonLabel = computed(() => {
  const r: GameOverReason | undefined = store.gameOverReason
  if (r === 'blockOut') return 'BLOCK OUT'
  if (r === 'lockOut') return 'LOCK OUT'
  if (r === 'topOut') return 'TOP OUT'
  return ''
})
</script>

<template>
  <header v-if="band === 'top'" class="game-hud game-hud--bar" aria-live="polite">
    <HoldQueue :hold-piece="store.holdPiece" :can-hold="store.canHold" />
    <div class="game-hud__stat">
      <span class="game-hud__label">Lvl</span>
      <span class="game-hud__value">{{ store.level }}</span>
    </div>
    <div class="game-hud__stat">
      <span class="game-hud__label">Lines</span>
      <span class="game-hud__value">{{ store.lines }}/{{ store.goal }}</span>
    </div>
    <div class="game-hud__stat game-hud__stat--phase">
      <span class="game-hud__label">Phase</span>
      <span class="game-hud__value game-hud__value--sm">{{ store.phaseLabel }}</span>
    </div>
    <NextQueue :pieces="store.nextPieces" />
  </header>

  <footer v-else class="game-hud game-hud--footer" aria-live="polite">
    <div v-if="store.gameOver" class="game-hud__game-over" role="status">
      <p class="game-hud__go-title">Game over</p>
      <p v-if="reasonLabel" class="game-hud__go-reason">{{ reasonLabel }}</p>
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
