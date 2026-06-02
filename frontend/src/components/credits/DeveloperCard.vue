<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

import '@/assets/styles/credits/developer-card.css'

const props = defineProps<{
  name: string
  handle: string
  url: string
}>()

const avatarUrl = computed(() => `https://avatars.githubusercontent.com/${props.handle}?s=240`)
const { t } = useI18n()
</script>

<template>
  <a
    class="developer-card"
    :href="url"
    target="_blank"
    rel="noopener noreferrer"
    :aria-label="t('credits.developerGithubAria', { name })"
  >
    <img
      class="developer-card__avatar"
      :src="avatarUrl"
      :alt="t('credits.developerAvatarAlt', { name })"
      width="120"
      height="120"
      decoding="async"
      referrerpolicy="no-referrer"
    />
    <p class="developer-card__name">{{ name }}</p>
    <p class="developer-card__handle">@{{ handle }}</p>
  </a>
</template>
