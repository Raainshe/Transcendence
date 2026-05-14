<script setup lang="ts">
import { onBeforeUnmount, onMounted } from 'vue'
import { useRouter } from 'vue-router'

import GameBoard from '@/components/game/GameBoard.vue'
import GameHud from '@/components/game/GameHud.vue'

import '@/assets/styles/views/game-play-view.css'

const router = useRouter()

function onGlobalKeydown(e: KeyboardEvent): void {
  if (e.key === 'Escape') {
    e.preventDefault()
    void router.push({ name: 'home' })
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
        Move: Left / Right · Rotate: Up, X, Z / Ctrl · Soft: Down · Hard: Space · Esc: menu
      </p>
    </GameHud>
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
