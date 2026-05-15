<script setup lang="ts">
import { RouterLink } from 'vue-router'

import '@/assets/styles/layout/site-footer.css'

const year = new Date().getFullYear()

const hints = [
  { keys: '\u2191\u2193', label: 'Navigate' },
  { keys: '\u2190\u2192', label: 'Change' },
  { keys: 'Enter', label: 'Start' },
] as const

const links = [
  { to: '/privacy', label: 'Privacy' },
  { to: '/terms', label: 'Terms' },
  { to: '/credits', label: 'Developed by' },
] as const
</script>

<template>
  <footer class="site-footer">
    <span class="site-footer__left">
      <span class="site-footer__brand">&copy; {{ year }} Transcendence</span>
      <RouterLink
        v-for="link in links"
        :key="link.to"
        :to="link.to"
        class="site-footer__link"
      >
        {{ link.label }}
      </RouterLink>
    </span>
    <span class="site-footer__hint" aria-hidden="true">
      <span v-for="hint in hints" :key="hint.label" class="site-footer__hint-item">
        <span class="site-footer__hint-key">{{ hint.keys }}</span>
        <span>{{ hint.label }}</span>
      </span>
    </span>
    <span class="site-footer__version">v0.1.0</span>
  </footer>
</template>
