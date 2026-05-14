<script setup lang="ts">
import { onMounted, ref, useTemplateRef } from 'vue'
import { useRouter } from 'vue-router'

import CycleSelector from '@/components/menu/CycleSelector.vue'
import MenuItem from '@/components/menu/MenuItem.vue'
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

const ITEM_COUNT = 3
const focusedIndex = ref(0)

const menuRef = useTemplateRef<HTMLElement>('menu')
const variationCyclerRef = useTemplateRef<CyclerHandle>('variationCycler')
const playersCyclerRef = useTemplateRef<CyclerHandle>('playersCycler')

onMounted(() => {
  menuRef.value?.focus()
})

function startNewGame() {
  void router.push({ name: 'play' })
}

function move(delta: 1 | -1) {
  focusedIndex.value = (focusedIndex.value + delta + ITEM_COUNT) % ITEM_COUNT
}

function activate() {
  if (focusedIndex.value === 0) {
    startNewGame()
  }
}

function cycle(delta: 1 | -1) {
  if (focusedIndex.value === 1) {
    if (delta === 1) variationCyclerRef.value?.next()
    else variationCyclerRef.value?.prev()
  } else if (focusedIndex.value === 2) {
    if (delta === 1) playersCyclerRef.value?.next()
    else playersCyclerRef.value?.prev()
  }
}

function onKeydown(event: KeyboardEvent) {
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
  >
    <h2 class="main-menu__title">Main Menu</h2>
    <ul class="main-menu__list">
      <MenuItem
        label="New Game"
        kind="action"
        :selected="focusedIndex === 0"
        @select="focusedIndex = 0"
        @activate="startNewGame"
      />
      <MenuItem
        label="Variation"
        kind="cycler"
        :selected="focusedIndex === 1"
        @select="focusedIndex = 1"
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
        @select="focusedIndex = 2"
      >
        <CycleSelector
          ref="playersCycler"
          v-model="settings.playerCount"
          :options="PLAYER_COUNTS"
          :format-label="formatPlayers"
          aria-label="Number of players"
        />
      </MenuItem>
    </ul>
  </section>
</template>
