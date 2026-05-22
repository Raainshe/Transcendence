import { defineStore } from 'pinia'
import { ref, watch } from 'vue'

import type { GameVariation, PlayerCount } from '@/types/game'

const AUDIO_STORAGE_KEY = 'transcendence-audio'

type StoredAudioPrefs = {
  sfxEnabled?: boolean
  sfxVolume?: number
  musicEnabled?: boolean
}

function loadAudioPrefs(): StoredAudioPrefs | null {
  if (typeof localStorage === 'undefined') return null
  try {
    const raw = localStorage.getItem(AUDIO_STORAGE_KEY)
    if (!raw) return null
    return JSON.parse(raw) as StoredAudioPrefs
  } catch {
    return null
  }
}

function saveAudioPrefs(prefs: StoredAudioPrefs): void {
  if (typeof localStorage === 'undefined') return
  try {
    localStorage.setItem(AUDIO_STORAGE_KEY, JSON.stringify(prefs))
  } catch {
    /* quota / private mode */
  }
}

const stored = loadAudioPrefs()

export const useGameSettingsStore = defineStore('gameSettings', () => {
  const variation = ref<GameVariation>('marathon')
  const playerCount = ref<PlayerCount>(1)
  /** Default ON when unset; persisted via localStorage after first change. */
  const sfxEnabled = ref(stored?.sfxEnabled !== false)
  const sfxVolume = ref(stored?.sfxVolume ?? 0.7)
  const musicEnabled = ref(stored?.musicEnabled !== false)

  watch(variation, (v) => {
    if (v === 'sprint' || v === 'ultra') {
      playerCount.value = 1
    }
  })

  watch(
    [sfxEnabled, sfxVolume, musicEnabled],
    () => {
      saveAudioPrefs({
        sfxEnabled: sfxEnabled.value,
        sfxVolume: sfxVolume.value,
        musicEnabled: musicEnabled.value,
      })
    },
    { deep: true },
  )

  return { variation, playerCount, sfxEnabled, sfxVolume, musicEnabled }
})
