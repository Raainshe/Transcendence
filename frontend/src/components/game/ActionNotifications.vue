<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { useI18n } from 'vue-i18n'

import '@/assets/styles/game/action-notifications.css'
import type { NotificationTone } from '@/game/fx/types'
import { useGameFxStore } from '@/stores/gameFx'

const fx = useGameFxStore()
const { notifications } = storeToRefs(fx)
const { t } = useI18n()

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
      {{ t(n.textKey, n.textParams ?? {}) }}
    </p>
  </div>
</template>
