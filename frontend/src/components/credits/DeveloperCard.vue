<script setup lang="ts">
import { computed } from 'vue'

import '@/assets/styles/credits/developer-card.css'

const props = defineProps<{
  name: string
  handle: string
  url: string
}>()

const avatarUrl = computed(() => `https://github.com/${props.handle}.png?size=240`)
</script>

<template>
  <a
    class="developer-card"
    :href="url"
    target="_blank"
    rel="noopener noreferrer"
    :aria-label="`${name} on GitHub`"
  >
    <img
      class="developer-card__avatar"
      :src="avatarUrl"
      :alt="`${name} GitHub avatar`"
      width="120"
      height="120"
      loading="lazy"
      decoding="async"
    />
    <p class="developer-card__name">{{ name }}</p>
    <p class="developer-card__handle">@{{ handle }}</p>
  </a>
</template>
