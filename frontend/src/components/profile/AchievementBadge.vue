<script setup lang="ts">
import { tierColors, type Tier } from '@/types/badges'

const props = defineProps<{
  tier: Tier
  catImage: string
  badgeName: string
}>()

const color = tierColors[props.tier]

</script>

<template>
  <div
    class="achievement-badge"
    :style="{
      '--badge-size': `80px`,
      '--tier-color': color,
    }"
    :title="badgeName"
  >

    <svg
      class="achievement-badge__piece"
      viewBox="0 0 4 4"
      aria-hidden="true"
    >
      <rect x="0" y="2" width="1" height="1" :fill="'rgba(184, 13, 247, 0.98)'" stroke="rgba(0,0,0,0.8)" stroke-width="0.1" />
      <rect x="0.5" y="2.5" width="0.5" height="0.5" fill="rgba(0, 0, 0, 0.8)" />
      <rect x="0" y="3" width="1" height="1" :fill="'rgba(184, 13, 247, 0.98)'" stroke="rgba(0,0,0,0.8)" stroke-width="0.1" />
      <rect x="0.5" y="3.5" width="0.5" height="0.5" fill="rgba(0,0,0,0.8)" />
      <rect x="1" y="3" width="1" height="1" :fill="'rgba(184, 13, 247, 0.98)'" stroke="rgba(0,0,0,0.8)" stroke-width="0.1" />
      <rect x="1.5" y="3.5" width="0.5" height="0.5" fill="rgba(0,0,0,0.8)" />
      <rect x="2" y="3" width="1" height="1" :fill="'rgba(184, 13, 247, 0.98)'" stroke="rgba(0,0,0,0.8)" stroke-width="0.1" />
      <rect x="2.5" y="3.5" width="0.5" height="0.5" fill="rgba(0,0,0,0.8)" />
    </svg>

    <img
      class="achievement-badge__cat"
      :src="`/badges/${catImage}`"
      :alt="badgeName"
    />
  </div>
</template>

<style scoped>
.achievement-badge {
  position: relative;
  width: 70px;
  height: 70px;
  border-radius: 50%;
  background-color: var(--tier-color);
  box-shadow:
    0 0 0 2px rgba(255, 255, 255, 0.12),
    0 0 12px color-mix(in srgb, var(--tier-color) 60%, transparent);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.achievement-badge__piece {
  position: absolute;
  bottom: 15%;
  left: 20%;
  width: 70%;
  height: 70%;
  box-sizing: border-box;
  opacity: 0.6;
}

.achievement-badge__cat {
  position: absolute;
  bottom: 33%;
  left: 60%;
  transform: translateX(-50%);
  width: 60%;
  height: 60%;
  object-fit: contain;
  mix-blend-mode: multiply;
  z-index: 1;
}
</style>