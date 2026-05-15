<script setup lang="ts">
import { MINO_COLORS, type PieceType } from '@/game/types'

defineProps<{
  pieces: PieceType[]
}>()
</script>

<template>
  <div class="next-queue" aria-label="Next pieces">
    <span class="next-queue__label">Next</span>
    <div class="next-queue__row" role="list">
      <span
        v-for="(p, i) in pieces"
        :key="`${i}-${p}`"
        class="next-queue__mino"
        role="listitem"
        :style="{ background: MINO_COLORS[p] }"
      />
    </div>
  </div>
</template>

<style scoped>
.next-queue {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 0.125rem;
  min-width: 0;
}

.next-queue__label {
  font-size: 0.5rem;
  color: var(--color-text-dim);
  text-transform: uppercase;
  letter-spacing: 0.06em;
}

.next-queue__row {
  display: flex;
  flex-direction: row;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 0.25rem;
  max-width: min(100vw - 6rem, 18rem);
}

.next-queue__mino {
  width: 0.65rem;
  height: 0.65rem;
  border-radius: 2px;
  flex-shrink: 0;
  box-shadow: inset 0 0 0 1px rgba(0, 0, 0, 0.25);
}
</style>
