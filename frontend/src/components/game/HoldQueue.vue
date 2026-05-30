<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

import { MINO_COLORS, type PieceType } from '@/game/types'

const props = defineProps<{
  holdPiece: PieceType | null
  canHold: boolean
}>()

const { t } = useI18n()
const holdPieceAria = computed(() =>
  props.holdPiece
    ? t('game.queue.holdFilledAria', { piece: props.holdPiece })
    : t('game.queue.holdEmptyAria'),
)
</script>

<template>
  <div
    class="hold-queue"
    :class="{ 'hold-queue--disabled': !canHold }"
    :aria-label="t('game.queue.holdAria')"
  >
    <span class="hold-queue__label">{{ t('game.queue.hold') }}</span>
    <div class="hold-queue__cell" role="img" :aria-label="holdPieceAria">
      <span
        v-if="holdPiece !== null"
        class="hold-queue__mino"
        :style="{ background: MINO_COLORS[holdPiece] }"
      />
    </div>
  </div>
</template>

<style scoped>
.hold-queue {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.125rem;
  min-width: 2.25rem;
}

.hold-queue--disabled .hold-queue__cell {
  opacity: 0.45;
}

.hold-queue__label {
  font-size: 0.5rem;
  color: var(--color-text-dim);
  text-transform: uppercase;
  letter-spacing: 0.06em;
}

.hold-queue__cell {
  width: 1.75rem;
  height: 1.75rem;
  box-sizing: border-box;
  border: 1px dashed var(--color-border);
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.2);
}

.hold-queue__mino {
  width: 1.1rem;
  height: 1.1rem;
  border-radius: 2px;
  box-shadow: inset 0 0 0 1px rgba(0, 0, 0, 0.25);
}
</style>
