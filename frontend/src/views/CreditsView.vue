<script setup lang="ts">
import { useI18n } from 'vue-i18n'

import DeveloperCard from '@/components/credits/DeveloperCard.vue'

import '@/assets/styles/views/credits-view.css'

type Developer = { name: string; handle: string; url: string }

const DEVELOPERS: readonly Developer[] = [
  { name: 'Anakin Pregitzer', handle: '4n4k1n', url: 'https://github.com/4n4k1n' },
  { name: 'Kevin Sinn', handle: 'Fearcon14', url: 'https://github.com/Fearcon14' },
  { name: 'Ryan Makoni', handle: 'Raainshe', url: 'https://github.com/Raainshe' },
] as const

const { t } = useI18n()
</script>

<template>
  <section class="credits-view">
    <header class="credits-view__header">
      <h1 class="credits-view__title">{{ t('credits.title') }}</h1>
      <p class="credits-view__subtitle">{{ t('credits.subtitle') }}</p>
    </header>

    <div class="credits-view__grid">
      <DeveloperCard
        v-for="dev in DEVELOPERS"
        :key="dev.handle"
        :name="dev.name"
        :handle="dev.handle"
        :url="dev.url"
      />
    </div>
  </section>
</template>
