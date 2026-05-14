<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref, useTemplateRef, watch } from 'vue'
import { useRouter } from 'vue-router'

import GameBoard from '@/components/game/GameBoard.vue'
import GameHud from '@/components/game/GameHud.vue'
import MenuItem from '@/components/menu/MenuItem.vue'
import { useGameSessionStore } from '@/stores/gameSession'

import '@/assets/styles/views/game-play-view.css'

const router = useRouter()
const store = useGameSessionStore()

const PAUSE_ITEM_COUNT = 2
const pauseFocusedIndex = ref(0)
const pauseMenuRef = useTemplateRef<HTMLElement>('pauseMenu')

watch(
  () => store.paused,
  (paused) => {
    if (!paused) return
    pauseFocusedIndex.value = 0
    void nextTick(() => {
      pauseMenuRef.value?.focus()
    })
  },
)

function movePauseFocus(delta: 1 | -1): void {
  pauseFocusedIndex.value =
    (pauseFocusedIndex.value + delta + PAUSE_ITEM_COUNT) % PAUSE_ITEM_COUNT
}

function activatePauseMenu(): void {
  if (pauseFocusedIndex.value === 0) {
    store.resume()
  } else {
    void router.push({ name: 'home' })
  }
}

function onPauseMenuKeydown(event: KeyboardEvent): void {
  switch (event.key) {
    case 'ArrowUp':
      event.preventDefault()
      event.stopPropagation()
      movePauseFocus(-1)
      break
    case 'ArrowDown':
      event.preventDefault()
      event.stopPropagation()
      movePauseFocus(1)
      break
    case 'Enter':
    case ' ':
      event.preventDefault()
      event.stopPropagation()
      activatePauseMenu()
      break
    default:
      break
  }
}

function onGlobalKeydown(e: KeyboardEvent): void {
  if (e.key === 'Escape') {
    e.preventDefault()
    if (!store.paused) {
      store.pause()
    } else {
      void router.push({ name: 'home' })
    }
    return
  }
  if (e.key === 'p' || e.key === 'P') {
    if (store.paused) {
      e.preventDefault()
      store.resume()
    }
  }
}

onMounted(() => {
  window.addEventListener('keydown', onGlobalKeydown)
})

onBeforeUnmount(() => {
  window.removeEventListener('keydown', onGlobalKeydown)
})
</script>

<template>
  <div class="game-play-view">
    <h1 class="visually-hidden">Play</h1>
    <GameHud band="top" />
    <div class="game-play-view__canvas-slot">
      <GameBoard />
    </div>
    <GameHud band="bottom">
      <p class="game-play-view__controls">
        Move: Left / Right · Rotate: Up, X, Z / Ctrl · Soft: Down · Hard: Space · Hold: C · Pause:
        Esc · Resume: P · Paused: Up/Down + Enter · Menu (when paused): Esc
      </p>
    </GameHud>
    <div
      v-if="store.paused"
      class="game-play-view__pause-overlay"
      role="dialog"
      aria-modal="true"
      aria-labelledby="pause-title"
    >
      <div class="game-play-view__pause-panel">
        <p id="pause-title" class="game-play-view__pause-title">Paused</p>
        <p class="game-play-view__pause-copy">
          Resume: P · Menu: Esc · Move: Up / Down · Select: Enter or Space
        </p>
        <section
          ref="pauseMenu"
          class="game-play-view__pause-menu"
          role="menu"
          aria-label="Pause menu"
          tabindex="0"
          @keydown="onPauseMenuKeydown"
        >
          <ul class="game-play-view__pause-list">
            <MenuItem
              label="Resume"
              kind="action"
              :selected="pauseFocusedIndex === 0"
              @select="pauseFocusedIndex = 0"
              @activate="store.resume()"
            />
            <MenuItem
              label="Quit to menu"
              kind="action"
              :selected="pauseFocusedIndex === 1"
              @select="pauseFocusedIndex = 1"
              @activate="void router.push({ name: 'home' })"
            />
          </ul>
        </section>
      </div>
    </div>
  </div>
</template>

<style scoped>
.visually-hidden {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}
</style>
