<script setup lang="ts">
import '@/assets/styles/layout/site-header.css'

const blocks = ['yellow', 'cyan', 'purple', 'orange', 'blue', 'green', 'red'] as const
</script>

<template>
  <header class="site-header">
    <div class="site-header__inner">
      <h1 class="site-header__wordmark">
        Transcendence<sup class="site-header__registered" aria-hidden="true">&reg;</sup>
      </h1>
      <div class="site-header__underline" aria-hidden="true">
        <span
          v-for="color in blocks"
          :key="color"
          class="site-header__underline-block"
          :class="`site-header__underline-block--${color}`"
        />
      </div>
    </div>
  </header>
</template>
