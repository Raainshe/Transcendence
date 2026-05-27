<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, useTemplateRef, watch } from 'vue'
import { useRouter } from 'vue-router'

import CycleSelector from '@/components/menu/CycleSelector.vue'
import MenuItem from '@/components/menu/MenuItem.vue'
import { gameAudio } from '@/game/audio/AudioManager'
import { useGameSettingsStore } from '@/stores/gameSettings'
import {
  GAME_VARIATIONS,
  GAME_VARIATION_LABELS,
  PLAYER_COUNTS,
  type GameVariation,
  type PlayerCount,
} from '@/types/game'

import '@/assets/styles/menu/main-menu.css'

type CyclerHandle = { prev: () => void; next: () => void }

const settings = useGameSettingsStore()
const router = useRouter()

const ITEM_COUNT = 4
const focusedIndex = ref(0)
const settingsReady = ref(false)

const menuRef = useTemplateRef<HTMLElement>('menu')
const variationCyclerRef = useTemplateRef<CyclerHandle>('variationCycler')
const playersCyclerRef = useTemplateRef<CyclerHandle>('playersCycler')

function uiPrefs() {
  return { sfxEnabled: settings.sfxEnabled, sfxVolume: settings.sfxVolume }
}

const audioOn = computed(() => settings.musicEnabled && settings.sfxEnabled)

const audioMenuLabel = computed(() => (audioOn.value ? 'Music: ON' : 'Music: OFF'))

function ensureAudioUnlocked(): void {
  gameAudio.unlock()
  gameAudio.setEnabled(settings.sfxEnabled)
  gameAudio.setMusicEnabled(settings.musicEnabled)
}

function toggleAudio(): void {
  gameAudio.unlock()
  const next = !audioOn.value
  settings.musicEnabled = next
  settings.sfxEnabled = next
  gameAudio.setEnabled(next)
  gameAudio.setMusicEnabled(next)
}

function playMove(): void {
  gameAudio.playUi('move', uiPrefs(), { volumeScale: 0.5 })
}

function playSelected(): void {
  gameAudio.playUi('selected', uiPrefs())
}

function focusItem(index: number): void {
  if (focusedIndex.value !== index) {
    playMove()
  }
  focusedIndex.value = index
}

onMounted(() => {
  menuRef.value?.focus()
  gameAudio.setEnabled(settings.sfxEnabled)
  gameAudio.setMusicEnabled(settings.musicEnabled)
  if (settings.musicEnabled) {
    gameAudio.startMenuMusic()
  }
  settingsReady.value = true
})

onBeforeUnmount(() => {
  gameAudio.stopMenuMusic()
})

watch(
  () => settings.variation,
  () => {
    if (!settingsReady.value) return
    playSelected()
  },
)

watch(
  () => settings.playerCount,
  () => {
    if (!settingsReady.value) return
    playSelected()
  },
)

function startNewGame() {
  playSelected()
  gameAudio.stopMenuMusic()
  void router.push({ name: 'play' })
}

function move(delta: 1 | -1): void {
  ensureAudioUnlocked()
  const next = (focusedIndex.value + delta + ITEM_COUNT) % ITEM_COUNT
  focusItem(next)
}

function activate() {
  ensureAudioUnlocked()
  if (focusedIndex.value === 0) {
    startNewGame()
  } else if (focusedIndex.value === 1) {
    toggleAudio()
  }
}

function cycle(delta: 1 | -1): void {
  ensureAudioUnlocked()
  if (focusedIndex.value === 2) {
    if (delta === 1) variationCyclerRef.value?.next()
    else variationCyclerRef.value?.prev()
  } else if (focusedIndex.value === 3) {
    if (delta === 1) playersCyclerRef.value?.next()
    else playersCyclerRef.value?.prev()
  }
}

function onKeydown(event: KeyboardEvent) {
  ensureAudioUnlocked()
  if (event.key === 'm' || event.key === 'M') {
    event.preventDefault()
    toggleAudio()
    return
  }
  switch (event.key) {
    case 'ArrowUp':
      event.preventDefault()
      move(-1)
      break
    case 'ArrowDown':
      event.preventDefault()
      move(1)
      break
    case 'ArrowLeft':
      event.preventDefault()
      cycle(-1)
      break
    case 'ArrowRight':
      event.preventDefault()
      cycle(1)
      break
    case 'Enter':
    case ' ':
      event.preventDefault()
      activate()
      break
  }
}

function formatVariation(value: GameVariation): string {
  return GAME_VARIATION_LABELS[value]
}

function formatPlayers(count: PlayerCount): string {
  return `${count} PLAYER${count === 1 ? '' : 'S'}`
}
</script>

<template>
  <section
    ref="menu"
    class="main-menu"
    role="menu"
    aria-label="Main Menu"
    tabindex="0"
    @keydown="onKeydown"
    @click="ensureAudioUnlocked"
  >
    <h2 class="main-menu__title">Main Menu</h2>
    <ul class="main-menu__list">
      <MenuItem
        label="New Game"
        kind="action"
        :selected="focusedIndex === 0"
        @select="focusItem(0)"
        @activate="startNewGame"
      />
     
      <MenuItem
        label="Variation"
        kind="cycler"
        :selected="focusedIndex === 1"
        @select="focusItem(2)"
      >
        <CycleSelector
          ref="variationCycler"
          v-model="settings.variation"
          :options="GAME_VARIATIONS"
          :format-label="formatVariation"
          aria-label="Game variation"
        />
      </MenuItem>
      <MenuItem
        label="Players"
        kind="cycler"
        :selected="focusedIndex === 2"
        @select="focusItem(3)"
      >
        <CycleSelector
          ref="playersCycler"
          v-model="settings.playerCount"
          :options="PLAYER_COUNTS"
          :format-label="formatPlayers"
          aria-label="Number of players"
        />
      </MenuItem>
      <MenuItem
        :label="audioMenuLabel"
        kind="action"
        :selected="focusedIndex === 3"
        @select="focusItem(1)"
        @activate="toggleAudio()"
      />
    </ul>
  </section>
</template>
