<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { useI18n } from 'vue-i18n'

import { SUPPORTED_LOCALES, type AppLocale } from '@/i18n'
import { useGameSettingsStore } from '@/stores/gameSettings'

import '@/assets/styles/layout/site-header.css'

const blocks = ['yellow', 'cyan', 'purple', 'orange', 'blue', 'green', 'red'] as const

const { t } = useI18n()
const settings = useGameSettingsStore()
const { locale } = storeToRefs(settings)

function setLocale(nextLocale: AppLocale): void {
  locale.value = nextLocale
}

function languageButtonLabel(option: AppLocale): string {
  return t(`header.languageCodes.${option}`)
}
</script>

<template>
  <header class="site-header">
    <div class="site-header__inner">
      <h1 class="site-header__wordmark">
        Transcendence<sup class="site-header__registered" aria-hidden="true">&reg;</sup>
      </h1>
      <div class="site-header__language" role="group" :aria-label="t('header.languageLabel')">
        <button
          v-for="option in SUPPORTED_LOCALES"
          :key="option"
          type="button"
          class="site-header__language-option"
          :class="{ 'site-header__language-option--active': locale === option }"
          :aria-pressed="locale === option"
          @click="setLocale(option)"
        >
          {{ languageButtonLabel(option) }}
        </button>
      </div>
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
