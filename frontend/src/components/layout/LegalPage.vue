<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import '@/assets/styles/layout/legal-page.css'

defineProps<{
  title: string
  updated?: string
}>()

const { t } = useI18n()

const showScrollTop = ref(false)

function onScroll() {
  showScrollTop.value = window.scrollY > 300
}

function scrollToTop() {
  const reduceMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches
  window.scrollTo({ top: 0, behavior: reduceMotion ? 'auto' : 'smooth' })
}

onMounted(() => {
  onScroll()
  window.addEventListener('scroll', onScroll, { passive: true })
})

onBeforeUnmount(() => {
  window.removeEventListener('scroll', onScroll)
})
</script>

<template>
  <section class="legal-page">
    <header class="legal-page__header">
      <h1 class="legal-page__title">{{ title }}</h1>
      <p v-if="updated" class="legal-page__subtitle">
        {{ t('legal.lastUpdated', { date: updated }) }}
      </p>
    </header>
    <article class="legal-page__body">
      <slot />
    </article>

    <button
      type="button"
      class="legal-page__scroll-top"
      :class="{ 'legal-page__scroll-top--visible': showScrollTop }"
      :aria-label="t('legal.scrollToTop')"
      @click="scrollToTop"
    >
      <span aria-hidden="true">&#x2191;</span>
    </button>
  </section>
</template>
