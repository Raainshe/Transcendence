<script setup lang="ts">
import { computed } from 'vue'
import { RouterView, useRoute } from 'vue-router'

import SiteFooter from '@/components/layout/SiteFooter.vue'
import SiteHeader from '@/components/layout/SiteHeader.vue'

import '@/assets/styles/app.css'

const route = useRoute()
const isPlayRoute = computed(() => route.name === 'play')
</script>

<template>
  <div class="app" :class="{ 'app--play': isPlayRoute }">
    <SiteHeader v-if="!isPlayRoute" />
    <main class="app__main">
      <RouterView />
    </main>
    <SiteFooter v-if="!isPlayRoute" />
  </div>
</template>
