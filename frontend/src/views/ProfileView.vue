<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'

import DeleteAccountModal from '@/components/profile/DeleteAccountModal.vue'
import { ApiError, resolveAssetUrl } from '@/api/client'
import * as usersApi from '@/api/users'
import { useAuthStore } from '@/stores/auth'
import type { UserStats } from '@/types/api'

import '@/assets/styles/views/profile-view.css'

const auth = useAuthStore()
const router = useRouter()
const { t } = useI18n()

const username = ref('')
const formError = ref<string | null>(null)
const formSuccess = ref(false)
const isSaving = ref(false)

const avatarError = ref<string | null>(null)
const isUploading = ref(false)
const isRemovingAvatar = ref(false)

const stats = ref<UserStats | null>(null)
const statsError = ref(false)
const statsLoading = ref(false)

const deleteModalOpen = ref(false)
const deleteError = ref<string | null>(null)
const isDeleting = ref(false)

const user = computed(() => auth.user)
const avatarSrc = computed(() => resolveAssetUrl(user.value?.avatar_url))
const initials = computed(() => {
  const name = user.value?.username ?? '?'
  return name.slice(0, 2).toUpperCase()
})
const usernameDirty = computed(
  () => user.value && username.value.trim() !== user.value.username,
)

watch(
  user,
  (next) => {
    if (next) username.value = next.username
  },
  { immediate: true },
)

onMounted(() => {
  void loadStats()
})

async function loadStats(): Promise<void> {
  if (!user.value?.id) return
  statsLoading.value = true
  statsError.value = false
  try {
    stats.value = await usersApi.getUserStats(user.value.id)
  } catch {
    stats.value = null
    statsError.value = true
  } finally {
    statsLoading.value = false
  }
}

function mapError(error: unknown): string {
  if (error instanceof ApiError) {
    const known: Record<string, string> = {
      'username already in use': 'profile.errors.usernameTaken',
      'only jpeg, png, webp, and gif images are allowed': 'profile.errors.avatarType',
      'file too large or invalid form data': 'profile.errors.avatarTooLarge',
    }
    const mapped = known[error.message.toLowerCase()]
    if (mapped) return t(mapped)
    if (error.message) return error.message
  }
  return t('profile.errors.generic')
}

async function onSaveUsername(): Promise<void> {
  formError.value = null
  formSuccess.value = false
  const next = username.value.trim()
  if (!next) {
    formError.value = t('profile.errors.usernameRequired')
    return
  }
  if (next.length > 32) {
    formError.value = t('profile.errors.usernameTooLong')
    return
  }
  if (!usernameDirty.value) return

  isSaving.value = true
  try {
    await auth.updateProfile(next)
    formSuccess.value = true
  } catch (error) {
    formError.value = mapError(error)
  } finally {
    isSaving.value = false
  }
}

async function onAvatarSelected(event: Event): Promise<void> {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (!file) return

  avatarError.value = null
  isUploading.value = true
  try {
    await auth.uploadAvatar(file)
  } catch (error) {
    avatarError.value = mapError(error)
  } finally {
    isUploading.value = false
  }
}

async function onRemoveAvatar(): Promise<void> {
  if (!user.value?.avatar_url) return
  avatarError.value = null
  isRemovingAvatar.value = true
  try {
    await auth.removeAvatar()
  } catch (error) {
    avatarError.value = mapError(error)
  } finally {
    isRemovingAvatar.value = false
  }
}

async function onDeleteAccount(): Promise<void> {
  deleteError.value = null
  isDeleting.value = true
  try {
    await auth.deleteAccount()
    deleteModalOpen.value = false
    await router.push({ name: 'home' })
  } catch (error) {
    deleteError.value = mapError(error)
  } finally {
    isDeleting.value = false
  }
}
</script>

<template>
  <div v-if="user" class="profile-view">
    <header class="profile-view__header">
      <h1 class="profile-view__title">{{ t('profile.title') }}</h1>
    </header>

    <section class="profile-view__panel" aria-labelledby="profile-avatar-heading">
      <h2 id="profile-avatar-heading" class="profile-view__section-title">
        {{ t('profile.avatar') }}
      </h2>
      <div class="profile-view__avatar-row">
        <div class="profile-view__avatar-preview">
          <img
            v-if="avatarSrc"
            :src="avatarSrc"
            :alt="t('profile.avatarAlt')"
            class="profile-view__avatar-image"
          />
          <span v-else class="profile-view__avatar-initials" aria-hidden="true">{{ initials }}</span>
        </div>
        <div class="profile-view__avatar-actions">
          <label class="profile-view__button profile-view__button--secondary">
            {{ isUploading ? t('profile.saving') : t('profile.uploadAvatar') }}
            <input
              type="file"
              accept="image/jpeg,image/png,image/webp,image/gif"
              class="profile-view__file-input"
              :disabled="isUploading || isRemovingAvatar"
              @change="onAvatarSelected"
            />
          </label>
          <button
            v-if="user.avatar_url"
            type="button"
            class="profile-view__button profile-view__button--secondary"
            :disabled="isUploading || isRemovingAvatar"
            @click="onRemoveAvatar"
          >
            {{ isRemovingAvatar ? t('profile.saving') : t('profile.removeAvatar') }}
          </button>
          <p class="profile-view__hint">{{ t('profile.avatarHint') }}</p>
          <p v-if="avatarError" class="profile-view__error" role="alert">{{ avatarError }}</p>
        </div>
      </div>
    </section>

    <section class="profile-view__panel" aria-labelledby="profile-identity-heading">
      <h2 id="profile-identity-heading" class="profile-view__section-title">
        {{ t('profile.username') }}
      </h2>
      <form class="profile-view__field" @submit.prevent="onSaveUsername">
        <label class="profile-view__label" for="profile-username">{{ t('profile.username') }}</label>
        <input
          id="profile-username"
          v-model="username"
          type="text"
          class="profile-view__input"
          maxlength="32"
          autocomplete="username"
          :disabled="isSaving"
        />
        <label class="profile-view__label" for="profile-email">{{ t('profile.email') }}</label>
        <input
          id="profile-email"
          :value="user.email"
          type="email"
          class="profile-view__input"
          disabled
          autocomplete="email"
        />
        <p class="profile-view__hint">{{ t('profile.emailReadOnly') }}</p>
        <p v-if="formSuccess" class="profile-view__message" role="status">{{ t('profile.saved') }}</p>
        <p v-if="formError" class="profile-view__error" role="alert">{{ formError }}</p>
        <button
          type="submit"
          class="profile-view__button"
          :disabled="isSaving || !usernameDirty"
        >
          {{ isSaving ? t('profile.saving') : t('profile.save') }}
        </button>
      </form>
    </section>

    <section class="profile-view__panel" aria-labelledby="profile-stats-heading">
      <h2 id="profile-stats-heading" class="profile-view__section-title">
        {{ t('profile.statsTitle') }}
      </h2>
      <p v-if="statsLoading" class="profile-view__hint">{{ t('profile.saving') }}</p>
      <p v-else-if="statsError" class="profile-view__error">{{ t('profile.statsLoadError') }}</p>
      <p v-else-if="stats && stats.games_played === 0" class="profile-view__hint">
        {{ t('profile.statsEmpty') }}
      </p>
      <div v-else-if="stats" class="profile-view__stats-grid">
        <div class="profile-view__stat">
          <span class="profile-view__stat-value">{{ stats.games_played }}</span>
          <span class="profile-view__stat-label">{{ t('profile.statsGames') }}</span>
        </div>
        <div class="profile-view__stat">
          <span class="profile-view__stat-value">{{ stats.wins }}</span>
          <span class="profile-view__stat-label">{{ t('profile.statsWins') }}</span>
        </div>
        <div class="profile-view__stat">
          <span class="profile-view__stat-value">{{ stats.best_score }}</span>
          <span class="profile-view__stat-label">{{ t('profile.statsBest') }}</span>
        </div>
        <div class="profile-view__stat">
          <span class="profile-view__stat-value">{{ stats.total_lines }}</span>
          <span class="profile-view__stat-label">{{ t('profile.statsLines') }}</span>
        </div>
        <div class="profile-view__stat">
          <span class="profile-view__stat-value">{{ stats.avg_score }}</span>
          <span class="profile-view__stat-label">{{ t('profile.statsAvg') }}</span>
        </div>
      </div>
    </section>

    <section class="profile-view__panel profile-view__danger" aria-labelledby="profile-danger-heading">
      <h2 id="profile-danger-heading" class="profile-view__section-title">
        {{ t('profile.dangerTitle') }}
      </h2>
      <p v-if="deleteError" class="profile-view__error" role="alert">{{ deleteError }}</p>
      <button
        type="button"
        class="profile-view__button profile-view__button--danger"
        @click="deleteModalOpen = true"
      >
        {{ t('profile.deleteAccount') }}
      </button>
    </section>

    <DeleteAccountModal
      :open="deleteModalOpen"
      :loading="isDeleting"
      @close="deleteModalOpen = false"
      @confirm="onDeleteAccount"
    />
  </div>
</template>
