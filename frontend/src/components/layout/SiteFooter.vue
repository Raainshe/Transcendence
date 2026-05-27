<script setup lang="ts">
import { RouterLink } from 'vue-router'
import { useI18n } from 'vue-i18n'

import '@/assets/styles/layout/site-footer.css'

const year = new Date().getFullYear()
const { t } = useI18n()

const hints = [
  { keys: '\u2191\u2193', labelKey: 'footer.hintNavigate' },
  { keys: '\u2190\u2192', labelKey: 'footer.hintChange' },
  { keys: 'Enter', labelKey: 'footer.hintStart' },
] as const

const links = [
  { to: '/privacy', labelKey: 'footer.privacy' },
  { to: '/terms', labelKey: 'footer.terms' },
  { to: '/credits', labelKey: 'footer.developedBy' },
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
        {{ t(link.labelKey) }}
      </RouterLink>
    </span>
    <span class="site-footer__hint" aria-hidden="true">
      <span v-for="hint in hints" :key="hint.labelKey" class="site-footer__hint-item">
        <span class="site-footer__hint-key">{{ hint.keys }}</span>
        <span>{{ t(hint.labelKey) }}</span>
      </span>
    </span>
    <span class="site-footer__version">{{ t('footer.version') }}</span>
  </footer>
</template>
