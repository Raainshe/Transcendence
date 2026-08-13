<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, useTemplateRef, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'

import * as lobbiesApi from '@/api/lobbies'
import JoinLobbyModal from '@/components/menu/JoinLobbyModal.vue'
import CycleSelector from '@/components/menu/CycleSelector.vue'
import MenuItem from '@/components/menu/MenuItem.vue'
import { gameAudio } from '@/game/audio/AudioManager'
import { useAuthStore } from '@/stores/auth'
import { useLobbyStore } from '@/stores/lobby'
import { useGameSettingsStore } from '@/stores/gameSettings'
import {
  GAME_VARIATIONS,
  GAME_VARIATION_LABELS,
  type GameVariation,
  type PlayerCount,
} from '@/types/game'

import '@/assets/styles/menu/main-menu.css'

type CyclerHandle = { prev: () => void; next: () => void }

const LOBBY_PLAYER_COUNTS = [2, 3, 4] as const
type LobbyPlayerCount = (typeof LOBBY_PLAYER_COUNTS)[number]

const settings = useGameSettingsStore()
const auth = useAuthStore()
const lobbyStore = useLobbyStore()
const router = useRouter()
const { t } = useI18n()

const focusedIndex = ref(0)
const settingsReady = ref(false)
const joinModalOpen = ref(false)
const createBusy = ref(false)
const createError = ref<string | null>(null)

const menuRef = useTemplateRef<HTMLElement>('menu')
const variationCyclerRef = useTemplateRef<CyclerHandle>('variationCycler')
const playersCyclerRef = useTemplateRef<CyclerHandle>('playersCycler')

const isMultiplayer = computed(() => settings.variation === 'multiplayer')

const itemCount = computed(() => (isMultiplayer.value ? 5 : 3))

const playerOptions = LOBBY_PLAYER_COUNTS

const primaryActionLabel = computed(() =>
  isMultiplayer.value ? t('menu.createLobby') : t('menu.newGame'),
)

function uiPrefs() {
  return { sfxEnabled: settings.sfxEnabled, sfxVolume: settings.sfxVolume }
}

const audioOn = computed(() => settings.musicEnabled && settings.sfxEnabled)

const audioMenuLabel = computed(() =>
  audioOn.value ? t('menu.audioOn') : t('menu.audioOff'),
)

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

function requireAuth(): boolean {
  if (auth.isAuthenticated) return true
  void router.push({ name: 'home', query: { login: '1' } })
  return false
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
    if (focusedIndex.value >= itemCount.value) {
      focusedIndex.value = itemCount.value - 1
    }
  },
)

watch(
  () => settings.playerCount,
  () => {
    if (!settingsReady.value) return
    playSelected()
  },
)

async function startNewGame() {
  playSelected()
  createError.value = null

  if (isMultiplayer.value) {
    if (!requireAuth()) return
    createBusy.value = true
    try {
      const res = await lobbiesApi.createLobby({ max_players: settings.playerCount })
      gameAudio.stopMenuMusic()
      void router.push({ name: 'lobby', params: { id: res.lobby.id } })
    } catch (err) {
      const existingId = await lobbyStore.existingLobbyRedirect(err)
      if (existingId) {
        gameAudio.stopMenuMusic()
        void router.push({ name: 'lobby', params: { id: existingId } })
        return
      }
      createError.value = lobbyStore.mapError(err)
    } finally {
      createBusy.value = false
    }
    return
  }

  gameAudio.stopMenuMusic()
  void router.push({ name: 'play' })
}

function openJoinModal(): void {
  playSelected()
  if (!requireAuth()) return
  joinModalOpen.value = true
}

function move(delta: 1 | -1): void {
  ensureAudioUnlocked()
  const next = (focusedIndex.value + delta + itemCount.value) % itemCount.value
  focusItem(next)
}

function activate() {
  ensureAudioUnlocked()
  if (focusedIndex.value === 0) {
    void startNewGame()
  } else if (focusedIndex.value === 3 && isMultiplayer.value) {
    openJoinModal()
  } else if (focusedIndex.value === itemCount.value - 1) {
    toggleAudio()
  }
}

function cycle(delta: 1 | -1): void {
  ensureAudioUnlocked()
  if (focusedIndex.value === 1) {
    if (delta === 1) variationCyclerRef.value?.next()
    else variationCyclerRef.value?.prev()
  } else if (focusedIndex.value === 2 && isMultiplayer.value) {
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

function formatPlayers(count: PlayerCount | LobbyPlayerCount): string {
  return t('menu.playersCount', { count })
}
</script>

<template>
  <section
    ref="menu"
    class="main-menu"
    role="menu"
    :aria-label="t('menu.ariaLabel')"
    tabindex="0"
    @keydown="onKeydown"
    @click="ensureAudioUnlocked"
  >
    <h2 class="main-menu__title">{{ t('menu.title') }}</h2>
    <p v-if="createError" class="main-menu__error" role="alert">{{ createError }}</p>
    <ul class="main-menu__list">
      <MenuItem
        :label="primaryActionLabel"
        kind="action"
        :selected="focusedIndex === 0"
        @select="focusItem(0)"
        @activate="void startNewGame()"
      />

      <MenuItem
        :label="t('menu.variation')"
        kind="cycler"
        :selected="focusedIndex === 1"
        @select="focusItem(1)"
      >
        <CycleSelector
          ref="variationCycler"
          v-model="settings.variation"
          :options="GAME_VARIATIONS"
          :format-label="formatVariation"
          :aria-label="t('menu.variationAriaLabel')"
          :prev-aria-label="t('cycler.previous')"
          :next-aria-label="t('cycler.next')"
        />
      </MenuItem>
      <MenuItem
        v-if="isMultiplayer"
        :label="t('menu.players')"
        kind="cycler"
        :selected="focusedIndex === 2"
        @select="focusItem(2)"
      >
        <CycleSelector
          ref="playersCycler"
          v-model="settings.playerCount"
          :options="playerOptions"
          :format-label="formatPlayers"
          :aria-label="t('menu.playersAriaLabel')"
          :prev-aria-label="t('cycler.previous')"
          :next-aria-label="t('cycler.next')"
        />
      </MenuItem>
      <MenuItem
        v-if="isMultiplayer"
        :label="t('menu.joinLobby')"
        kind="action"
        :selected="focusedIndex === 3"
        @select="focusItem(3)"
        @activate="openJoinModal()"
      />
      <MenuItem
        :label="audioMenuLabel"
        kind="action"
        :selected="focusedIndex === itemCount - 1"
        @select="focusItem(itemCount - 1)"
        @activate="toggleAudio()"
      />
    </ul>
    <JoinLobbyModal :open="joinModalOpen" @close="joinModalOpen = false" />
  </section>
</template>
