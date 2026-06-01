<script setup lang="ts">
import { useId } from 'vue'
import { useI18n } from 'vue-i18n'

import '@/assets/styles/auth/auth-modal.css'

const props = defineProps<{
  open: boolean
  loading?: boolean
}>()

const emit = defineEmits<{
  close: []
  confirm: []
}>()

const { t } = useI18n()

const titleId = useId()

function onBackdropKeydown(event: KeyboardEvent): void {
  if (event.key === 'Escape' && !props.loading) {
    event.preventDefault()
    event.stopPropagation()
    emit('close')
  }
}

function onConfirm(): void {
  emit('confirm')
}
</script>

<template>
  <div
    v-if="open"
    class="auth-modal-overlay"
    tabindex="-1"
    @keydown="onBackdropKeydown"
    @click.self="!loading && emit('close')"
  >
    <div
      class="auth-modal-panel"
      role="alertdialog"
      aria-modal="true"
      :aria-labelledby="titleId"
      @click.stop
    >
      <h2 :id="titleId" class="auth-modal-panel__title">{{ t('profile.deleteTitle') }}</h2>
      <p class="auth-modal-panel__body">{{ t('profile.deleteBody') }}</p>
      <div class="auth-modal-panel__actions">
        <button
          type="button"
          class="auth-modal-panel__submit"
          :disabled="loading"
          @click="onConfirm"
        >
          {{ loading ? t('profile.deleting') : t('profile.deleteConfirm') }}
        </button>
        <button
          type="button"
          class="auth-modal-panel__close"
          :disabled="loading"
          @click="emit('close')"
        >
          {{ t('profile.deleteCancel') }}
        </button>
      </div>
    </div>
  </div>
</template>
