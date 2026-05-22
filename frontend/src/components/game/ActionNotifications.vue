<script setup lang="ts">
import { storeToRefs } from 'pinia'

import '@/assets/styles/game/action-notifications.css'
import type { NotificationTone } from '@/game/fx/types'
import { useGameFxStore } from '@/stores/gameFx'

const fx = useGameFxStore()
const { notifications } = storeToRefs(fx)

function toneClass(tone: NotificationTone): string {
  return `action-notifications__item--${tone}`
}
</script>

<template>
  <div class="action-notifications" aria-live="polite" aria-atomic="false">
    <p
      v-for="n in notifications"
      :key="n.id"
      class="action-notifications__item"
      :class="toneClass(n.tone)"
    >
      {{ n.text }}
    </p>
  </div>
</template>
