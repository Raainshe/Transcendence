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
    <h1 class="game-play-view__title">Play</h1>
    <div class="game-play-view__main">
      <GameHud />
      <GameBoard />
    </div>
    <p class="game-play-view__controls">
      Move: Arrow Left / Right · Rotate: Arrow Up, X (CW), Z / Left Ctrl (CCW) · Soft drop: Arrow Down ·
      Hard drop: Space · Esc: menu
    </p>
  </div>
</template>

<style scoped>
.game-play-view__title {
  font-family: var(--font-display);
  font-size: var(--fs-lg);
  margin: 0;
  text-align: center;
}
</style>
