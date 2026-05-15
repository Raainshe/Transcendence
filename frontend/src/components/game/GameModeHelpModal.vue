<script setup lang="ts">
import { computed, nextTick, useId, useTemplateRef, watch } from 'vue'

import '@/assets/styles/game/game-mode-help.css'
import { getGameVariationInfo, type GameVariation } from '@/types/game'

const props = defineProps<{
  variation: GameVariation
  open: boolean
}>()

const emit = defineEmits<{
  close: []
}>()

const titleId = useId()
const backButtonRef = useTemplateRef<HTMLButtonElement>('backButton')

const info = computed(() => getGameVariationInfo(props.variation))

watch(
  () => props.open,
  (isOpen) => {
    if (!isOpen) return
    void nextTick(() => {
      backButtonRef.value?.focus()
    })
  },
)

function onBackdropKeydown(event: KeyboardEvent): void {
  if (event.key === 'Escape') {
    event.preventDefault()
    event.stopPropagation()
    emit('close')
  }
}

function onBackKeydown(event: KeyboardEvent): void {
  if (event.key === 'Enter' || event.key === ' ') {
    event.preventDefault()
    emit('close')
  }
}
</script>

<template>
  <div
    v-if="open"
    class="game-mode-help-overlay"
    tabindex="-1"
    @keydown="onBackdropKeydown"
  >
    <div
      class="game-mode-help-panel"
      role="dialog"
      aria-modal="true"
      :aria-labelledby="titleId"
      @click.stop
    >
      <h2 :id="titleId" class="game-mode-help-panel__title">{{ info.label }}</h2>
      <p class="game-mode-help-panel__body">{{ info.description }}</p>
      <p class="game-mode-help-panel__subtitle">How to win</p>
      <p class="game-mode-help-panel__win">{{ info.howToWin }}</p>
      <div class="game-mode-help-panel__actions">
        <button
          ref="backButton"
          type="button"
          class="game-mode-help-panel__back"
          @click="emit('close')"
          @keydown="onBackKeydown"
        >
          Back to game
        </button>
      </div>
    </div>
  </div>
</template>
