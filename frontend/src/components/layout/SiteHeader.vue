<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { computed, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { RouterLink, useRoute, useRouter } from 'vue-router'

import { resolveAssetUrl } from '@/api/client'
import * as friendsApi from '@/api/friends'
import AuthModal, { type AuthModalTab } from '@/components/auth/AuthModal.vue'
import { SUPPORTED_LOCALES, type AppLocale } from '@/i18n'
import { useAuthStore } from '@/stores/auth'
import { useGameSettingsStore } from '@/stores/gameSettings'

import '@/assets/styles/layout/site-header.css'

const blocks = ['yellow', 'cyan', 'purple', 'orange', 'blue', 'green', 'red'] as const

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const settings = useGameSettingsStore()
const auth = useAuthStore()
const { locale } = storeToRefs(settings)
const { isAuthenticated, user, status: authStatus } = storeToRefs(auth)

const authModalOpen = ref(false)
const authModalTab = ref<AuthModalTab>('login')
const pendingRequestCount = ref(0)

const avatarSrc = computed(() => resolveAssetUrl(user.value?.avatar_url))

async function loadPendingRequestCount(): Promise<void> {
  if (!isAuthenticated.value) {
    pendingRequestCount.value = 0
    return
  }
  try {
    const { requests } = await friendsApi.getFriendRequests()
    pendingRequestCount.value = requests.length
  } catch {
    pendingRequestCount.value = 0
  }
}

onMounted(() => {
  void loadPendingRequestCount()
})

watch(isAuthenticated, () => {
  void loadPendingRequestCount()
})

watch(
  () => route.name,
  () => {
    void loadPendingRequestCount()
  },
)

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
  window.location.assign('/')
}

function clearLoginQuery(): void {
  if (route.query.login === undefined) return
  const { login: _login, ...query } = route.query
  void router.replace({ query })
}

watch(
  () => [route.query.login, authStatus.value, isAuthenticated.value] as const,
  ([login, status, authed]) => {
    if (authed) {
      if (login !== undefined) clearLoginQuery()
      authModalOpen.value = false
      return
    }
    if (login === '1' && status !== 'loading') {
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
            :to="{ name: 'friends' }"
            class="site-header__nav-link"
          >
            {{ t('header.friends') }}
            <span
              v-if="pendingRequestCount > 0"
              class="site-header__nav-badge"
              :aria-label="t('header.friendsPending', { count: pendingRequestCount })"
            >
              {{ pendingRequestCount }}
            </span>
          </RouterLink>
          <RouterLink
            :to="{ name: 'chat' }"
            class="site-header__nav-link"
          >
            {{ t('header.chat') }}
          </RouterLink>
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
