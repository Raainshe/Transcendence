<script setup lang="ts">
import { computed } from 'vue'

import type { GameOverReason } from '@/game/types'
import { useGameSessionStore } from '@/stores/gameSession'

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
  <aside class="game-hud" aria-live="polite">
    <div class="game-hud__block">
      <span class="game-hud__label">Level</span>
      <span class="game-hud__value">{{ store.level }}</span>
    </div>
    <div class="game-hud__block">
      <span class="game-hud__label">Lines</span>
      <span class="game-hud__value">{{ store.lines }} / {{ store.goal }}</span>
    </div>
    <div class="game-hud__block">
      <span class="game-hud__label">Phase</span>
      <span class="game-hud__value game-hud__value--sm">{{ store.phaseLabel }}</span>
    </div>
    <div class="game-hud__block game-hud__block--wide">
      <span class="game-hud__label">Next</span>
      <span class="game-hud__next">{{ store.nextQueueLabels.join(' ') }}</span>
    </div>
    <div v-if="store.gameOver" class="game-hud__game-over" role="status">
      <p class="game-hud__go-title">Game over</p>
      <p v-if="reasonLabel" class="game-hud__go-reason">{{ reasonLabel }}</p>
      <p class="game-hud__go-hint">Press Esc to return to the menu.</p>
    </div>
  </aside>
</template>

<style scoped>
.game-hud {
  font-family: var(--font-display);
  font-size: var(--fs-xs);
  color: var(--color-text);
  background: var(--color-panel);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  padding: var(--sp-4);
  min-width: 12rem;
  display: flex;
  flex-direction: column;
  gap: var(--sp-3);
}

.game-hud__block {
  display: flex;
  flex-direction: column;
  gap: var(--sp-1);
}

.game-hud__block--wide {
  max-width: 100%;
}

.game-hud__label {
  font-size: var(--fs-xs);
  color: var(--color-text-dim);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.game-hud__value {
  font-size: var(--fs-sm);
}

.game-hud__value--sm {
  font-size: var(--fs-xs);
  word-break: break-word;
}

.game-hud__next {
  font-family: var(--font-mono);
  font-size: var(--fs-sm);
  letter-spacing: 0.15em;
  color: var(--accent-selected);
}

.game-hud__game-over {
  margin-top: var(--sp-2);
  padding-top: var(--sp-3);
  border-top: 1px solid var(--color-border);
}

.game-hud__go-title {
  margin: 0 0 var(--sp-2);
  font-size: var(--fs-sm);
  color: var(--t-red);
}

.game-hud__go-reason {
  margin: 0 0 var(--sp-2);
  font-size: var(--fs-xs);
  color: var(--color-text-dim);
}

.game-hud__go-hint {
  margin: 0;
  font-size: var(--fs-xs);
  color: var(--color-text-dim);
  line-height: 1.5;
}
</style>
