<script setup lang="ts">
import { computed } from 'vue'

import { gameAudio } from '@/game/audio/AudioManager'
import { useGameSettingsStore } from '@/stores/gameSettings'

import '@/assets/styles/menu/audio-mute-toggle.css'

const props = defineProps<{
  /** `music` = menu theme loop; `sfx` = in-game sound effects */
  target: 'music' | 'sfx'
}>()

const settings = useGameSettingsStore()

const enabled = computed(() =>
  props.target === 'music' ? settings.musicEnabled : settings.sfxEnabled,
)

const ariaLabel = computed(() =>
  props.target === 'music'
    ? enabled.value
      ? 'Mute menu music'
      : 'Unmute menu music'
    : enabled.value
      ? 'Mute sound effects'
      : 'Unmute sound effects',
)

const title = computed(() =>
  props.target === 'music'
    ? enabled.value
      ? 'Mute music (M)'
      : 'Unmute music (M)'
    : enabled.value
      ? 'Mute sound effects (M)'
      : 'Unmute sound effects (M)',
)

function toggle(): void {
  gameAudio.unlock()
  if (props.target === 'music') {
    settings.musicEnabled = !settings.musicEnabled
    gameAudio.setMusicEnabled(settings.musicEnabled)
  } else {
    settings.sfxEnabled = !settings.sfxEnabled
    gameAudio.setEnabled(settings.sfxEnabled)
  }
}

defineExpose({ toggle })
</script>

<template>
  <button
    type="button"
    class="audio-mute-toggle"
    :aria-label="ariaLabel"
    :aria-pressed="enabled"
    :title="title"
    @click="toggle"
  >
    <svg v-if="enabled" class="audio-mute-toggle__icon" viewBox="0 0 24 24" aria-hidden="true">
      <path
        d="M3 9v6h4l5 5V4L7 9H3zm13.5 3c0-1.77-1.02-3.29-2.5-4.03v8.06c1.48-.74 2.5-2.26 2.5-4.03zM14 3.23v2.06c2.89.86 5 3.54 5 6.71s-2.11 5.85-5 6.71v2.06c4.01-.91 7-4.49 7-8.77s-2.99-7.86-7-8.77z"
      />
    </svg>
    <svg v-else class="audio-mute-toggle__icon" viewBox="0 0 24 24" aria-hidden="true">
      <path
        d="M16.5 12c0-1.77-1.02-3.29-2.5-4.03v2.71l2.5 2.5c0-.23.03-.47.03-.71zm2.5 0c0 .94-.2 1.82-.54 2.64l1.51 1.51C20.63 14.91 21 13.5 21 12c0-4.28-2.99-7.86-7-8.77v2.06c2.89.86 5 3.54 5 6.71zM4.27 3L3 4.27 7.73 9H3v6h4l5 5v-6.73l4.25 4.25c-.67.52-1.42.93-2.25 1.18v2.06c1.48-.74 2.5-2.26 2.5-4.03 0-1.06-.27-2.05-.74-2.92l1.46 1.46C16.21 14.07 16.5 13.06 16.5 12c0-1.77-1.02-3.29-2.5-4.03v1.71l2.07 2.07L19.73 21 21 19.73 4.27 3z"
      />
    </svg>
  </button>
</template>
