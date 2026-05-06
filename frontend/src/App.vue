<script setup lang="ts">
import { RouterView } from 'vue-router'

import SiteFooter from '@/components/layout/SiteFooter.vue'
import SiteHeader from '@/components/layout/SiteHeader.vue'

import '@/assets/styles/app.css'
</script>

<template>
  <div class="app">
    <SiteHeader />
    <main class="app__main">
      <RouterView />
    </main>
    <SiteFooter />
  </div>
</template>
