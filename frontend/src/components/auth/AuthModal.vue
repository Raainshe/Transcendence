<script setup lang="ts">
import { computed, nextTick, ref, useId, useTemplateRef, watch } from 'vue'
import { useI18n } from 'vue-i18n'

import { ApiError } from '@/api/client'
import { useAuthStore } from '@/stores/auth'

import '@/assets/styles/auth/auth-modal.css'

export type AuthModalTab = 'login' | 'register'

const props = defineProps<{
  open: boolean
  initialTab?: AuthModalTab
}>()

const emit = defineEmits<{
  close: []
}>()

const auth = useAuthStore()
const { t } = useI18n()

const titleId = useId()
const errorId = useId()
const emailFieldRef = useTemplateRef<HTMLInputElement>('emailField')
const usernameFieldRef = useTemplateRef<HTMLInputElement>('usernameField')

const activeTab = ref<AuthModalTab>('login')
const email = ref('')
const password = ref('')
const username = ref('')
const confirmPassword = ref('')
const errorMessage = ref<string | null>(null)
const isSubmitting = ref(false)

const isLogin = computed(() => activeTab.value === 'login')

watch(
  () => props.open,
  (isOpen) => {
    if (!isOpen) return
    activeTab.value = props.initialTab ?? 'login'
    email.value = ''
    password.value = ''
    username.value = ''
    confirmPassword.value = ''
    errorMessage.value = null
    isSubmitting.value = false
    void nextTick(() => {
      if (activeTab.value === 'login') {
        emailFieldRef.value?.focus()
      } else {
        usernameFieldRef.value?.focus()
      }
    })
  },
)

watch(activeTab, (tab) => {
  errorMessage.value = null
  if (!props.open) return
  void nextTick(() => {
    if (tab === 'login') {
      emailFieldRef.value?.focus()
    } else {
      usernameFieldRef.value?.focus()
    }
  })
})

function onBackdropKeydown(event: KeyboardEvent): void {
  if (event.key === 'Escape') {
    event.preventDefault()
    event.stopPropagation()
    emit('close')
  }
}

function mapApiError(error: unknown): string {
  if (error instanceof ApiError) {
    const known: Record<string, string> = {
      'invalid credentials': 'auth.errors.invalidCredentials',
      'email already in use': 'auth.errors.emailTaken',
      'username already in use': 'auth.errors.usernameTaken',
      'password must be at least 8 characters': 'auth.errors.passwordTooShort',
      'username, email and password are required': 'auth.errors.required',
    }
    const mapped = known[error.message.toLowerCase()]
    if (mapped) {
      return t(mapped)
    }
    if (error.message) {
      return error.message
    }
  }
  return t('auth.errors.generic')
}

function validateLogin(): string | null {
  if (!email.value.trim() || !password.value) {
    return t('auth.errors.required')
  }
  return null
}

function validateRegister(): string | null {
  if (!username.value.trim() || !email.value.trim() || !password.value || !confirmPassword.value) {
    return t('auth.errors.required')
  }
  if (username.value.length > 32) {
    return t('auth.errors.usernameTooLong')
  }
  if (password.value.length < 8) {
    return t('auth.errors.passwordTooShort')
  }
  if (password.value !== confirmPassword.value) {
    return t('auth.errors.passwordMismatch')
  }
  return null
}

async function onSubmit(): Promise<void> {
  errorMessage.value = null
  const validationError = isLogin.value ? validateLogin() : validateRegister()
  if (validationError) {
    errorMessage.value = validationError
    return
  }

  isSubmitting.value = true
  try {
    if (isLogin.value) {
      await auth.login(email.value.trim(), password.value)
    } else {
      await auth.register(username.value.trim(), email.value.trim(), password.value)
    }
    emit('close')
  } catch (error) {
    errorMessage.value = mapApiError(error)
  } finally {
    isSubmitting.value = false
  }
}

function setTab(tab: AuthModalTab): void {
  activeTab.value = tab
}
</script>

<template>
  <div
    v-if="open"
    class="auth-modal-overlay"
    tabindex="-1"
    @keydown="onBackdropKeydown"
    @click.self="emit('close')"
  >
    <div
      class="auth-modal-panel"
      role="dialog"
      aria-modal="true"
      :aria-labelledby="titleId"
      :aria-describedby="errorMessage ? errorId : undefined"
      @click.stop
    >
      <h2 :id="titleId" class="auth-modal-panel__sr-title">{{ t('auth.dialogLabel') }}</h2>

      <div class="auth-modal-panel__tabs" role="tablist" :aria-label="t('auth.dialogLabel')">
        <button
          type="button"
          role="tab"
          class="auth-modal-panel__tab"
          :class="{ 'auth-modal-panel__tab--active': activeTab === 'login' }"
          :aria-selected="activeTab === 'login'"
          @click="setTab('login')"
        >
          {{ t('auth.tabLogin') }}
        </button>
        <button
          type="button"
          role="tab"
          class="auth-modal-panel__tab"
          :class="{ 'auth-modal-panel__tab--active': activeTab === 'register' }"
          :aria-selected="activeTab === 'register'"
          @click="setTab('register')"
        >
          {{ t('auth.tabRegister') }}
        </button>
      </div>

      <form class="auth-modal-panel__form" @submit.prevent="onSubmit">
        <div v-if="!isLogin" class="auth-modal-panel__field">
          <label class="auth-modal-panel__label" for="auth-username">{{ t('auth.username') }}</label>
          <input
            id="auth-username"
            ref="usernameField"
            v-model="username"
            type="text"
            class="auth-modal-panel__input"
            autocomplete="username"
            maxlength="32"
            :disabled="isSubmitting"
          />
        </div>

        <div class="auth-modal-panel__field">
          <label class="auth-modal-panel__label" for="auth-email">{{ t('auth.email') }}</label>
          <input
            id="auth-email"
            ref="emailField"
            v-model="email"
            type="email"
            class="auth-modal-panel__input"
            autocomplete="email"
            :disabled="isSubmitting"
          />
        </div>

        <div class="auth-modal-panel__field">
          <label class="auth-modal-panel__label" for="auth-password">{{ t('auth.password') }}</label>
          <input
            id="auth-password"
            v-model="password"
            type="password"
            class="auth-modal-panel__input"
            :autocomplete="isLogin ? 'current-password' : 'new-password'"
            :disabled="isSubmitting"
          />
        </div>

        <div v-if="!isLogin" class="auth-modal-panel__field">
          <label class="auth-modal-panel__label" for="auth-confirm">{{
            t('auth.confirmPassword')
          }}</label>
          <input
            id="auth-confirm"
            v-model="confirmPassword"
            type="password"
            class="auth-modal-panel__input"
            autocomplete="new-password"
            :disabled="isSubmitting"
          />
        </div>

        <p v-if="errorMessage" :id="errorId" class="auth-modal-panel__error" role="alert">
          {{ errorMessage }}
        </p>

        <div class="auth-modal-panel__actions">
          <button type="submit" class="auth-modal-panel__submit" :disabled="isSubmitting">
            {{ isLogin ? t('auth.submitLogin') : t('auth.submitRegister') }}
          </button>
          <button type="button" class="auth-modal-panel__close" :disabled="isSubmitting" @click="emit('close')">
            {{ t('auth.close') }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>
