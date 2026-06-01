<script setup lang="ts">
import { useI18n } from 'vue-i18n'

import LegalPage from '@/components/layout/LegalPage.vue'

const REPO_URL = 'https://github.com/raainshe/Transcendence'
const LICENSE_URL = 'https://github.com/raainshe/Transcendence/blob/main/LICENSE'
const { t } = useI18n()
</script>

<template>
  <LegalPage :title="t('legal.terms.title')" :updated="t('legal.termsPageUpdated')">
    <h2>{{ t('legal.terms.acceptanceTitle') }}</h2>
    <p>{{ t('legal.terms.acceptanceBody') }}</p>

    <h2>{{ t('legal.terms.educationTitle') }}</h2>
    <p>{{ t('legal.terms.educationBody') }}</p>

    <h2>{{ t('legal.terms.licenseTitle') }}</h2>
    <p>
      {{ t('legal.terms.licenseBody') }}
      <a :href="LICENSE_URL" target="_blank" rel="noopener noreferrer">{{ LICENSE_URL }}</a>.
    </p>

    <h2>{{ t('legal.terms.trademarkTitle') }}</h2>
    <p>{{ t('legal.terms.trademarkBody') }}</p>

    <h2>{{ t('legal.terms.warrantyTitle') }}</h2>
    <p>{{ t('legal.terms.warrantyBody') }}</p>

    <h2>{{ t('legal.terms.liabilityTitle') }}</h2>
    <p>{{ t('legal.terms.liabilityBody') }}</p>

    <h2>{{ t('legal.terms.useTitle') }}</h2>
    <p>{{ t('legal.terms.useBody') }}</p>

    <h2>{{ t('legal.terms.changesTitle') }}</h2>
    <p>{{ t('legal.terms.changesBody') }}</p>

    <h2>{{ t('legal.terms.contactTitle') }}</h2>
    <p>
      {{ t('legal.terms.contactBody') }}
      <a :href="REPO_URL" target="_blank" rel="noopener noreferrer">{{ REPO_URL }}</a>.
    </p>
  </LegalPage>
</template>
