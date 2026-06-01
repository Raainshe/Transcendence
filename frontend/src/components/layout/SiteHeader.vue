<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { RouterLink, useRoute } from 'vue-router'

import { resolveAssetUrl } from '@/api/client'
import AuthModal, { type AuthModalTab } from '@/components/auth/AuthModal.vue'
import { SUPPORTED_LOCALES, type AppLocale } from '@/i18n'
import { useAuthStore } from '@/stores/auth'
import { useGameSettingsStore } from '@/stores/gameSettings'

import '@/assets/styles/layout/site-header.css'

const blocks = ['yellow', 'cyan', 'purple', 'orange', 'blue', 'green', 'red'] as const

const route = useRoute()
const { t } = useI18n()
const settings = useGameSettingsStore()
const auth = useAuthStore()
const { locale } = storeToRefs(settings)
const { isAuthenticated, user, status: authStatus } = storeToRefs(auth)

const authModalOpen = ref(false)
const authModalTab = ref<AuthModalTab>('login')

const avatarSrc = computed(() => resolveAssetUrl(user.value?.avatar_url))

function setLocale(nextLocale: AppLocale): void {
  locale.value = nextLocale
}

function languageButtonLabel(option: AppLocale): string {
  return t(`header.languageCodes.${option}`)
}

function openAuthModal(tab: AuthModalTab): void {
  authModalTab.value = tab
  authModalOpen.value = true
}

function closeAuthModal(): void {
  authModalOpen.value = false
}

async function onLogout(): Promise<void> {
  await auth.logout()
}

watch(
  () => route.query.login,
  (value) => {
    if (value === '1' && !isAuthenticated.value) {
      openAuthModal('login')
    }
  },
  { immediate: true },
)
</script>

<template>
  <header class="site-header">
    <div class="site-header__inner">
      <RouterLink
        :to="{ name: 'home' }"
        class="site-header__wordmark-link"
        :aria-label="t('header.homeAria')"
      >
        <h1 class="site-header__wordmark">
          Transcendence<sup class="site-header__registered" aria-hidden="true">&reg;</sup>
        </h1>
      </RouterLink>

      <div class="site-header__auth">
        <template v-if="isAuthenticated && user">
          <RouterLink
            :to="{ name: 'profile' }"
            class="site-header__user"
            :title="t('header.signedInAs', { username: user.username })"
          >
            <img
              v-if="avatarSrc"
              :src="avatarSrc"
              :alt="t('profile.avatarAlt')"
              class="site-header__user-avatar"
            />
            <span class="site-header__user-name">@{{ user.username }}</span>
          </RouterLink>
          <button
            type="button"
            class="site-header__auth-button site-header__auth-button--secondary"
            :disabled="authStatus === 'loading'"
            @click="onLogout"
          >
            {{ t('header.logout') }}
          </button>
        </template>
        <template v-else>
          <button
            type="button"
            class="site-header__auth-button site-header__auth-button--primary"
            @click="openAuthModal('login')"
          >
            {{ t('header.login') }}
          </button>
          <button
            type="button"
            class="site-header__auth-button site-header__auth-button--link"
            @click="openAuthModal('register')"
          >
            {{ t('header.signUp') }}
          </button>
        </template>
      </div>

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

    <AuthModal :open="authModalOpen" :initial-tab="authModalTab" @close="closeAuthModal" />
  </header>
</template>
