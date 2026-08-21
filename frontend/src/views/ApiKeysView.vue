<script setup lang="ts">
import { computed, onMounted, ref, useId } from 'vue'
import { useI18n } from 'vue-i18n'

import * as apiKeysApi from '@/api/apikeys'
import { ApiError } from '@/api/client'
import type { ApiKeyCreated, ApiKeySummary } from '@/types/api'

import '@/assets/styles/views/api-keys-view.css'
import '@/assets/styles/auth/auth-modal.css'

const { t, locale } = useI18n()

const keys = ref<ApiKeySummary[]>([])
const isLoading = ref(true)
const loadError = ref(false)

const newKeyName = ref('')
const isCreating = ref(false)
const createError = ref<string | null>(null)

const createdKey = ref<ApiKeyCreated | null>(null)
const copied = ref(false)
const revealTitleId = useId()

const revokeTarget = ref<ApiKeySummary | null>(null)
const isRevoking = ref(false)
const revokeError = ref<string | null>(null)
const revokeTitleId = useId()

const dateFormatter = computed(
  () =>
    new Intl.DateTimeFormat(locale.value, {
      dateStyle: 'medium',
      timeStyle: 'short',
    }),
)

onMounted(() => {
  void loadKeys()
})

async function loadKeys(): Promise<void> {
  isLoading.value = true
  loadError.value = false
  try {
    const res = await apiKeysApi.getApiKeys()
    keys.value = res.api_keys
  } catch {
    keys.value = []
    loadError.value = true
  } finally {
    isLoading.value = false
  }
}

function formatDate(iso: string | null): string {
  if (!iso) return t('apiKeys.never')
  const date = new Date(iso)
  if (Number.isNaN(date.getTime())) return t('apiKeys.never')
  return dateFormatter.value.format(date)
}

function mapError(error: unknown): string {
  if (error instanceof ApiError && error.message) return error.message
  return t('apiKeys.errors.generic')
}

async function onCreate(): Promise<void> {
  createError.value = null
  const name = newKeyName.value.trim()
  if (!name) {
    createError.value = t('apiKeys.errors.nameRequired')
    return
  }
  if (name.length > 64) {
    createError.value = t('apiKeys.errors.nameTooLong')
    return
  }

  isCreating.value = true
  try {
    const res = await apiKeysApi.createApiKey(name)
    createdKey.value = res.api_key
    copied.value = false
    newKeyName.value = ''
    await loadKeys()
  } catch (error) {
    createError.value = mapError(error)
  } finally {
    isCreating.value = false
  }
}

async function onCopyKey(): Promise<void> {
  if (!createdKey.value) return
  try {
    await navigator.clipboard.writeText(createdKey.value.key)
    copied.value = true
  } catch {
    copied.value = false
  }
}

function closeCreatedModal(): void {
  createdKey.value = null
  copied.value = false
}

function askRevoke(key: ApiKeySummary): void {
  revokeTarget.value = key
  revokeError.value = null
}

function cancelRevoke(): void {
  if (isRevoking.value) return
  revokeTarget.value = null
  revokeError.value = null
}

async function onRevoke(): Promise<void> {
  const target = revokeTarget.value
  if (!target) return
  isRevoking.value = true
  revokeError.value = null
  try {
    await apiKeysApi.revokeApiKey(target.id)
    keys.value = keys.value.filter((k) => k.id !== target.id)
    revokeTarget.value = null
  } catch (error) {
    revokeError.value = mapError(error)
  } finally {
    isRevoking.value = false
  }
}
</script>

<template>
  <div class="api-keys-view">
    <header class="api-keys-view__header">
      <RouterLink :to="{ name: 'profile' }" class="api-keys-view__back">
        {{ t('apiKeys.backToProfile') }}
      </RouterLink>
      <h1 class="api-keys-view__title">{{ t('apiKeys.title') }}</h1>
    </header>

    <section class="api-keys-view__panel" aria-labelledby="api-keys-create-heading">
      <h2 id="api-keys-create-heading" class="api-keys-view__section-title">
        {{ t('apiKeys.createTitle') }}
      </h2>
      <form class="api-keys-view__field" @submit.prevent="onCreate">
        <label class="api-keys-view__label" for="api-key-name">{{ t('apiKeys.nameLabel') }}</label>
        <input
          id="api-key-name"
          v-model="newKeyName"
          type="text"
          class="api-keys-view__input"
          maxlength="64"
          :placeholder="t('apiKeys.namePlaceholder')"
          :disabled="isCreating"
        />
        <p v-if="createError" class="api-keys-view__error" role="alert">{{ createError }}</p>
        <button type="submit" class="api-keys-view__button" :disabled="isCreating">
          {{ isCreating ? t('apiKeys.creating') : t('apiKeys.create') }}
        </button>
      </form>
    </section>

    <section class="api-keys-view__panel" aria-labelledby="api-keys-list-heading">
      <h2 id="api-keys-list-heading" class="api-keys-view__section-title">
        {{ t('apiKeys.listTitle') }}
      </h2>
      <p v-if="isLoading" class="api-keys-view__hint">{{ t('apiKeys.loading') }}</p>
      <p v-else-if="loadError" class="api-keys-view__error">{{ t('apiKeys.loadError') }}</p>
      <p v-else-if="keys.length === 0" class="api-keys-view__hint">{{ t('apiKeys.empty') }}</p>
      <div v-else class="api-keys-view__list">
        <div v-for="key in keys" :key="key.id" class="api-keys-view__row">
          <div class="api-keys-view__row-meta">
            <span class="api-keys-view__row-name">{{ key.name }}</span>
            <div class="api-keys-view__row-dates">
              <span>{{ t('apiKeys.created') }}: {{ formatDate(key.created_at) }}</span>
              <span>{{ t('apiKeys.lastUsed') }}: {{ formatDate(key.last_used_at) }}</span>
            </div>
          </div>
          <button
            type="button"
            class="api-keys-view__button api-keys-view__button--danger"
            @click="askRevoke(key)"
          >
            {{ t('apiKeys.revoke') }}
          </button>
        </div>
      </div>
    </section>
    <div v-if="createdKey" class="auth-modal-overlay" tabindex="-1" @keydown.escape="closeCreatedModal">
      <div class="auth-modal-panel" role="alertdialog" aria-modal="true" :aria-labelledby="revealTitleId" @click.stop>
        <h2 :id="revealTitleId" class="auth-modal-panel__title">{{ t('apiKeys.revealTitle') }}</h2>
        <p class="api-keys-view__reveal-warning">{{ t('apiKeys.revealWarning') }}</p>
        <div class="api-keys-view__reveal-row">
          <code class="api-keys-view__reveal-code">{{ createdKey.key }}</code>
          <button
            type="button"
            class="api-keys-view__button api-keys-view__button--secondary"
            @click="onCopyKey"
          >
            {{ copied ? t('apiKeys.copied') : t('apiKeys.copy') }}
          </button>
        </div>
        <div class="auth-modal-panel__actions">
          <button type="button" class="auth-modal-panel__submit" @click="closeCreatedModal">
            {{ t('apiKeys.done') }}
          </button>
        </div>
      </div>
    </div>
    <div v-if="revokeTarget" class="auth-modal-overlay" tabindex="-1" @keydown.escape="cancelRevoke" @click.self="cancelRevoke">
      <div class="auth-modal-panel" role="alertdialog" aria-modal="true" :aria-labelledby="revokeTitleId" @click.stop>
        <h2 :id="revokeTitleId" class="auth-modal-panel__title">{{ t('apiKeys.revokeTitle') }}</h2>
        <p class="auth-modal-panel__body"> {{ t('apiKeys.revokeBody', { name: revokeTarget.name }) }}</p>
        <p v-if="revokeError" class="auth-modal-panel__error" role="alert">{{ revokeError }}</p>
        <div class="auth-modal-panel__actions">
          <button
            type="button"
            class="auth-modal-panel__submit"
            :disabled="isRevoking"
            @click="onRevoke"
          >
            {{ isRevoking ? t('apiKeys.revoking') : t('apiKeys.revokeConfirm') }}
          </button>
          <button
            type="button"
            class="auth-modal-panel__close"
            :disabled="isRevoking"
            @click="cancelRevoke"
          >
            {{ t('apiKeys.revokeCancel') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
