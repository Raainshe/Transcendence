import { defineStore } from 'pinia'
import { ref, watch } from 'vue'

import type { GameVariation, PlayerCount } from '@/types/game'

export const useGameSettingsStore = defineStore('gameSettings', () => {
  const variation = ref<GameVariation>('marathon')
  const playerCount = ref<PlayerCount>(1)

  watch(variation, (v) => {
    if (v === 'sprint' || v === 'ultra') {
      playerCount.value = 1
    }
  })

  return { variation, playerCount }
})
