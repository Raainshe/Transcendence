<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  xp: number
}>()

function xpForLevel(level: number): number {
  if (level <= 1) return 0
  return 300 * (level - 1) * level / 2
}

function levelFromXP(xp: number): number {
  let level = 1
  while (xpForLevel(level + 1) <= xp) {
    level++
  }
  return level
}

const userLevel = computed(() => levelFromXP(props.xp))
const currLevelBaseXP = computed(() => xpForLevel(userLevel.value))
const nextLevelBaseXP = computed(() => xpForLevel(userLevel.value + 1))

const xpIntoLevel = computed(() => props.xp - currLevelBaseXP.value)
const xpNeeded = computed(() => nextLevelBaseXP.value - currLevelBaseXP.value)

const progress = computed(() => (xpIntoLevel.value / xpNeeded.value) * 100)

</script>

<template>
  <div class="xp-bar" :title="`Level ${level} — ${xp} XP total`">
   <p class="xp-bar__level">{{ level }}</p>
    <div
      class="xp-bar__track"
      role="progressbar"
      :aria-valuenow="progress"
      aria-valuemin="0"
      aria-valuemax="100"
    >
      <div class="xp-bar__fill" :style="{ width: `${progress}%` }">
      </div>
      <div
        v-for="(_, i) in 19"
        :key="i"
        class="xp-bar__segment"
        :style="{ left: `${((i + 1) / 20) * 100}%` }"
      />
    </div>
    <p class="xp-bar__count">{{ xpIntoLevel }} / {{ xpNeeded }} XP</p>
  </div>
</template>

<style scoped>
.xp-bar {
  display: flex;
  flex-direction: column;
  gap: var(--sp-1);
  width: 100%;
}

.xp-bar__segment {
  position: absolute;
  top: 0;
  bottom: 0;
  width: 1px;
  background: rgba(0, 0, 0, 0.55);
  transform: translateX(-50%);
}

.xp-bar__level {
  text-align: center;
  font-family: var(--font-display);
  font-size: 0.8rem;
  color: #1ed81e;
}

.xp-bar__count {
  font-family: var(--font-mono);
  font-size: 0.5rem;
  color: var(--color-text-dim);
  letter-spacing: 0.06em;
}

.xp-bar__track {
  position: relative;
  width: 100%;
  height: 12px;
  background: var(--color-bg);
  border: 1px solid rgba(89, 177, 1, 0.38);
  overflow: hidden;
}

.xp-bar__fill {
  height: 100%;
  background: #1ed81e;
  transition: width 600ms ease-out;
}

</style>