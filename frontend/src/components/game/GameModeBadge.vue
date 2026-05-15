<script setup lang="ts">
import { computed, useId } from 'vue'

import '@/assets/styles/game/game-mode-help.css'
import { getGameVariationInfo, type GameVariation } from '@/types/game'

const props = withDefaults(
  defineProps<{
    variation: GameVariation
    size?: 'sm' | 'md'
    showHeading?: boolean
  }>(),
  {
    size: 'sm',
    showHeading: true,
  },
)

const tooltipId = useId()

const info = computed(() => getGameVariationInfo(props.variation))
</script>

<template>
  <span
    class="game-mode-badge"
    :class="`game-mode-badge--${size}`"
    :data-tooltip="info.tooltip"
    :title="info.tooltip"
    tabindex="0"
    :aria-label="`Game mode: ${info.label}`"
    :aria-describedby="tooltipId"
  >
    <span v-if="showHeading" class="game-mode-badge__heading">Mode</span>
    <span class="game-mode-badge__label">{{ info.label }}</span>
    <span :id="tooltipId" class="visually-hidden">{{ info.tooltip }}</span>
  </span>
</template>

<style scoped>
.visually-hidden {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}
</style>
