<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import { useRoute } from 'vue-router'
import * as achievementsApi from '@/api/achievements'
import type { Achievements } from '@/types/api'
import { ApiError } from '@/api/client'
import { badgeDefinitions, type BadgeDefinition } from '@/types/badges'
import { useI18n } from 'vue-i18n'

import '@/assets/styles/views/achievements-view.css'

const { t } = useI18n()
const route = useRoute()

const achievements = ref<Achievements | null>(null)
const loading = ref(true)
const loadError = ref(false)
const notFound = ref(false)

const userId = computed(() => String(route.params.id ?? ''))

const entries = computed(() => {
  if (!achievements.value) return []
  const a = achievements.value
  return badgeDefinitions.map((def) => ({
    key: def.key,
    description: t(`achievements.descriptions.${def.key}`),
    unlocked: a[def.key] as boolean,
  }))
})

onMounted(() => {
  void loadAchievements()
})

async function loadAchievements(): Promise<void> {
  loading.value = true
  loadError.value = false
  notFound.value = false
  try {
    achievements.value = await achievementsApi.getUserAchievements(userId.value)
  } catch (error) {
    if (error instanceof ApiError && error.status === 404) {
      notFound.value = true
    } else {
      loadError.value = true
    }
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <section class="achievements-view">
    <header class="achievements-view__header">
      <h1 class="achievements-view__title">{{ t('achievements.title')}}</h1>
    </header>

    <div class="achievements-view__panel">
      <p v-if="loading" class="achievements-view__status">{{ t('achievements.loading') }}</p>
      <p v-else-if="loadError"class="achievements-view__status achievements-view__status--error">
        {{ t('achievements.loadError') }}</p>
      <div v-else class="achievements-view__table-wrap">
       <table class="achievements-view__table">
          <thead>
            <tr>
              <th scope="col">{{ t('achievements.colAchievements') }}</th>
              <th scope="col">{{ t('achievements.colDescription') }}</th>
              <th scope="col">{{ t('achievements.colStatus') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in entries" :key="row.key" :class="{ 'achievements-view__row--locked': !row.unlocked }">
              <td class="achievements-view__achievements">{{ row.key }}</td>
              <td class="achievements-view__description">{{ row.description }}</td>
              <td class="achievements-view__status-cell"> {{ row.unlocked ? t('achievements.unlocked') : t('achievements.locked') }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </section>
</template>
