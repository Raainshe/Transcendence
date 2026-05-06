<script setup lang="ts">
import DeveloperCard from '@/components/credits/DeveloperCard.vue'

import '@/assets/styles/views/credits-view.css'

type Developer = { name: string; handle: string; url: string }

const DEVELOPERS: readonly Developer[] = [
  { name: 'Anakin Pregitzer', handle: '4n4k1n', url: 'https://github.com/4n4k1n' },
  { name: 'Kevin Sinn', handle: 'Fearcon14', url: 'https://github.com/Fearcon14' },
  { name: 'Ryan Makoni', handle: 'Raainshe', url: 'https://github.com/Raainshe' },
] as const

const underlineBlocks = ['yellow', 'cyan', 'purple', 'orange', 'blue', 'green', 'red'] as const
</script>

<template>
  <section class="credits-view">
    <header class="credits-view__header">
      <h1 class="credits-view__title">Credits</h1>
      <div class="credits-view__underline" aria-hidden="true">
        <span
          v-for="color in underlineBlocks"
          :key="color"
          class="credits-view__underline-block"
          :class="`credits-view__underline-block--${color}`"
        />
      </div>
      <p class="credits-view__subtitle">Developed by</p>
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
