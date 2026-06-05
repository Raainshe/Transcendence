<script setup lang="ts">
import { nextTick, ref, useId, useTemplateRef, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'

import * as lobbiesApi from '@/api/lobbies'
import { useLobbyStore } from '@/stores/lobby'

import '@/assets/styles/auth/auth-modal.css'

const props = defineProps<{
  open: boolean
}>()

const emit = defineEmits<{
  close: []
}>()

const { t } = useI18n()
const router = useRouter()
const lobbyStore = useLobbyStore()

const titleId = useId()
const errorId = useId()
const codeFieldRef = useTemplateRef<HTMLInputElement>('codeField')

const inviteCode = ref('')
const errorMessage = ref<string | null>(null)
const isSubmitting = ref(false)

watch(
  () => props.open,
  (isOpen) => {
    if (!isOpen) return
    inviteCode.value = ''
    errorMessage.value = null
    isSubmitting.value = false
    void nextTick(() => {
      codeFieldRef.value?.focus()
    })
  },
)

function onBackdropKeydown(event: KeyboardEvent): void {
  if (event.key === 'Escape') {
    event.preventDefault()
    event.stopPropagation()
    emit('close')
  }
}

function normalizeCode(raw: string): string {
  return raw.trim().toUpperCase().replace(/[^A-Z0-9]/g, '')
}

async function submit(): Promise<void> {
  const code = normalizeCode(inviteCode.value)
  if (!code) {
    errorMessage.value = t('lobby.errors.codeRequired')
    return
  }
  isSubmitting.value = true
  errorMessage.value = null
  try {
    const res = await lobbiesApi.joinLobbyByCode({ invite_code: code })
    if (!res) {
      errorMessage.value = t('lobby.errors.notFound')
      return
    }
    emit('close')
    void router.push({ name: 'lobby', params: { id: res.lobby.id } })
  } catch (err) {
    errorMessage.value = lobbyStore.mapError(err)
  } finally {
    isSubmitting.value = false
  }
}
</script>

<template>
  <Teleport to="body">
    <div
      v-if="open"
      class="auth-modal-overlay"
      role="presentation"
      @click.self="emit('close')"
      @keydown="onBackdropKeydown"
    >
      <div
        class="auth-modal-panel"
        role="dialog"
        aria-modal="true"
        :aria-labelledby="titleId"
        :aria-describedby="errorMessage ? errorId : undefined"
        @click.stop
      >
        <h2 :id="titleId" class="auth-modal-panel__sr-title">{{ t('menu.joinLobby') }}</h2>
        <form class="auth-modal-panel__form" @submit.prevent="submit">
          <div class="auth-modal-panel__field">
            <label class="auth-modal-panel__label" for="join-lobby-code">
              {{ t('lobby.inviteCode') }}
            </label>
            <input
              id="join-lobby-code"
              ref="codeField"
              v-model="inviteCode"
              class="auth-modal-panel__input"
              type="text"
              autocomplete="off"
              maxlength="8"
              :placeholder="t('lobby.codePlaceholder')"
              :disabled="isSubmitting"
            />
          </div>
          <p v-if="errorMessage" :id="errorId" class="auth-modal-panel__error" role="alert">
            {{ errorMessage }}
          </p>
          <div class="auth-modal-panel__actions">
            <button type="submit" class="auth-modal-panel__submit" :disabled="isSubmitting">
              {{ t('lobby.joinSubmit') }}
            </button>
            <button
              type="button"
              class="auth-modal-panel__close"
              :disabled="isSubmitting"
              @click="emit('close')"
            >
              {{ t('auth.close') }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </Teleport>
</template>
