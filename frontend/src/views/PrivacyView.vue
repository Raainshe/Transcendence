<script setup lang="ts">
import { useI18n } from 'vue-i18n'

import LegalPage from '@/components/layout/LegalPage.vue'

const REPO_URL = 'https://github.com/raainshe/Transcendence'
const { t } = useI18n()
</script>

<template>
  <LegalPage :title="t('legal.privacy.title')" updated="May 2026">
    <h2>{{ t('legal.privacy.overviewTitle') }}</h2>
    <p>{{ t('legal.privacy.overviewBody') }}</p>

    <h2>{{ t('legal.privacy.notCollectTitle') }}</h2>
    <ul>
      <li>{{ t('legal.privacy.notCollect1') }}</li>
      <li>{{ t('legal.privacy.notCollect2') }}</li>
      <li>{{ t('legal.privacy.notCollect3') }}</li>
      <li>{{ t('legal.privacy.notCollect4') }}</li>
    </ul>

    <h2>{{ t('legal.privacy.localStorageTitle') }}</h2>
    <p>{{ t('legal.privacy.localStorageBody') }}</p>

    <h2>{{ t('legal.privacy.thirdPartyTitle') }}</h2>
    <p>{{ t('legal.privacy.thirdPartyBody', { handle: 'handle' }) }}</p>

    <h2>{{ t('legal.privacy.childrenTitle') }}</h2>
    <p>{{ t('legal.privacy.childrenBody') }}</p>

    <h2>{{ t('legal.privacy.controlsTitle') }}</h2>
    <p>{{ t('legal.privacy.controlsBody') }}</p>

    <h2>{{ t('legal.privacy.changesTitle') }}</h2>
    <p>{{ t('legal.privacy.changesBody') }}</p>

    <h2>{{ t('legal.privacy.contactTitle') }}</h2>
    <p>
      {{ t('legal.privacy.contactBody') }}
      <a :href="REPO_URL" target="_blank" rel="noopener noreferrer">{{ REPO_URL }}</a>.
    </p>
  </LegalPage>
</template>
