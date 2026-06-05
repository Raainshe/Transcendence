import { defineStore } from 'pinia'
import { ref, watch } from 'vue'

import { DEFAULT_LOCALE, SUPPORTED_LOCALES, type AppLocale } from '@/i18n'
import type { GameVariation, PlayerCount } from '@/types/game'

const AUDIO_STORAGE_KEY = 'transcendence-audio'
const LOCALE_STORAGE_KEY = 'transcendence-locale'

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

function loadLocale(): AppLocale {
  if (typeof localStorage === 'undefined') return DEFAULT_LOCALE
  try {
    const raw = localStorage.getItem(LOCALE_STORAGE_KEY)
    if (!raw) return DEFAULT_LOCALE
    if (SUPPORTED_LOCALES.includes(raw as AppLocale)) {
      return raw as AppLocale
    }
    return DEFAULT_LOCALE
  } catch {
    return DEFAULT_LOCALE
  }
}

function saveLocale(locale: AppLocale): void {
  if (typeof localStorage === 'undefined') return
  try {
    localStorage.setItem(LOCALE_STORAGE_KEY, locale)
  } catch {
    /* quota / private mode */
  }
}

export const useGameSettingsStore = defineStore('gameSettings', () => {
  const variation = ref<GameVariation>('marathon')
  const playerCount = ref<PlayerCount>(1)
  /** Default ON when unset; persisted via localStorage after first change. */
  const sfxEnabled = ref(stored?.sfxEnabled !== false)
  const sfxVolume = ref(stored?.sfxVolume ?? 0.7)
  const musicEnabled = ref(stored?.musicEnabled !== false)
  const locale = ref<AppLocale>(loadLocale())

  watch(variation, (v) => {
    if (v === 'sprint' || v === 'ultra') {
      playerCount.value = 1
    } else if (v === 'multiplayer' && playerCount.value < 2) {
      playerCount.value = 2
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

  watch(locale, (nextLocale) => {
    saveLocale(nextLocale)
  })

  return { variation, playerCount, sfxEnabled, sfxVolume, musicEnabled, locale }
})
