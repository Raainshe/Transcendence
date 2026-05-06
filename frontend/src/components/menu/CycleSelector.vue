<script setup lang="ts" generic="T">
import { nextTick, ref } from 'vue'

import '@/assets/styles/menu/cycle-selector.css'

const props = defineProps<{
  modelValue: T
  options: readonly T[]
  formatLabel?: (value: T) => string
  ariaLabel?: string
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: T): void
}>()

const flash = ref(false)
let flashTimer: number | undefined

async function trigger(value: T) {
  emit('update:modelValue', value)
  flash.value = false
  await nextTick()
  flash.value = true
  if (flashTimer !== undefined) {
    window.clearTimeout(flashTimer)
  }
  flashTimer = window.setTimeout(() => {
    flash.value = false
  }, 240)
}

function step(delta: 1 | -1) {
  const len = props.options.length
  if (len === 0) return
  const currentIdx = props.options.indexOf(props.modelValue)
  const startIdx = currentIdx === -1 ? 0 : currentIdx
  const nextIdx = (startIdx + delta + len) % len
  const next = props.options[nextIdx]
  if (next !== undefined) {
    void trigger(next)
  }
}

function prev() {
  step(-1)
}

function next() {
  step(1)
}

function format(value: T): string {
  return props.formatLabel ? props.formatLabel(value) : String(value)
}

defineExpose({ prev, next })
</script>

<template>
  <div class="cycle-selector" role="group" :aria-label="ariaLabel">
    <button
      type="button"
      class="cycle-selector__arrow cycle-selector__arrow--prev"
      tabindex="-1"
      aria-label="Previous"
      @mousedown.prevent
      @click="prev"
    >
      &#9668;
    </button>
    <span
      class="cycle-selector__value"
      :class="{ 'cycle-selector__value--flash': flash }"
      aria-live="polite"
    >
      {{ format(modelValue) }}
    </span>
    <button
      type="button"
      class="cycle-selector__arrow cycle-selector__arrow--next"
      tabindex="-1"
      aria-label="Next"
      @mousedown.prevent
      @click="next"
    >
      &#9658;
    </button>
  </div>
</template>
