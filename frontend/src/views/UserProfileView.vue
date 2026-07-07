<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter, RouterLink } from 'vue-router'

import GameDetailModal from '@/components/game/GameDetailModal.vue'
import MatchHistoryList from '@/components/game/MatchHistoryList.vue'
import { ApiError, resolveAssetUrl } from '@/api/client'
import * as usersApi from '@/api/users'
import { useAuthStore } from '@/stores/auth'
import type { User, UserStats } from '@/types/api'
import { badgeDefinitions } from '@/types/badges'
import AchievementBadge from '@/components/profile/AchievementBadge.vue'

import '@/assets/styles/views/user-profile-view.css'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const { t } = useI18n()

const profileUser = ref<User | null>(null)
const stats = ref<UserStats | null>(null)
const loading = ref(true)
const notFound = ref(false)
const loadError = ref(false)

const selectedGameId = ref<string | null>(null)
const gameDetailOpen = ref(false)

const userId = computed(() => String(route.params.id ?? ''))
const avatarSrc = computed(() => resolveAssetUrl(profileUser.value?.avatar_url))
const initials = computed(() => profileUser.value?.username.slice(0, 2).toUpperCase() ?? '?')
const onlineLabel = computed(() =>
  profileUser.value?.is_online ? t('userProfile.online') : t('userProfile.offline'),
)

const unlockedBadges = computed(() => {
  if (!profileUser.value?.achievement_list) return []
  const a = profileUser.value.achievement_list
  return badgeDefinitions
    .filter((def) => a[def.key] as boolean)
    .map((def) => ({
      key: def.key,
      badgeName: def.badgeName,
      tier: def.tier,
      catImage: def.catImage,
    }))
})

watch(
  userId,
  (id) => {
    if (auth.isAuthenticated && auth.user?.id === id) {
      void router.replace({ name: 'profile' })
      return
    }
    void loadProfile(id)
  },
  { immediate: true },
)

onMounted(() => {
  if (auth.isAuthenticated && auth.user?.id === userId.value) {
    void router.replace({ name: 'profile' })
  }
})

async function loadProfile(id: string): Promise<void> {
  if (!id) {
    notFound.value = true
    loading.value = false
    return
  }

  loading.value = true
  notFound.value = false
  loadError.value = false
  profileUser.value = null
  stats.value = null

  try {
    const [user, userStats] = await Promise.all([
      usersApi.getUser(id),
      usersApi.getUserStats(id),
    ])
    profileUser.value = user
    stats.value = userStats
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

function openGameDetail(gameId: string): void {
  selectedGameId.value = gameId
  gameDetailOpen.value = true
}

function closeGameDetail(): void {
  gameDetailOpen.value = false
  selectedGameId.value = null
}
</script>

<template>
  <section class="user-profile-view">
    <header class="user-profile-view__header">
      <RouterLink :to="{ name: 'home' }" class="user-profile-view__back">
        {{ t('userProfile.back') }}
      </RouterLink>
      <h1 class="user-profile-view__title">{{ t('userProfile.title') }}</h1>
    </header>

    <p v-if="loading" class="user-profile-view__status">{{ t('userProfile.loading') }}</p>
    <p v-else-if="notFound" class="user-profile-view__status">{{ t('userProfile.notFound') }}</p>
    <p v-else-if="loadError" class="user-profile-view__status">{{ t('userProfile.loadError') }}</p>

    <template v-else-if="profileUser">
      <section class="user-profile-view__panel">
        <div class="user-profile-view__identity">
          <img
            v-if="avatarSrc"
            :src="avatarSrc"
            :alt="t('friends.avatarAlt', { username: profileUser.username })"
            class="user-profile-view__avatar"
            width="80"
            height="80"
          />
          <span
            v-else
            class="user-profile-view__avatar user-profile-view__avatar--placeholder"
            aria-hidden="true"
          >
            {{ initials }}
          </span>
          <div class="user-profile-view__meta">
            <span class="user-profile-view__username">@{{ profileUser.username }}</span>
            <span
              class="user-profile-view__online"
              :class="{ 'user-profile-view__online--active': profileUser.is_online }"
            >
              {{ onlineLabel }}
            </span>
          </div>
        </div>
      </section>

      <section class="user-profile-view__panel" aria-labelledby="user-profile-stats-heading">
        <h2 id="user-profile-stats-heading" class="user-profile-view__section-title">
          {{ t('profile.statsTitle') }}
        </h2>
        <p v-if="stats && stats.games_played === 0" class="user-profile-view__status">
          {{ t('profile.statsEmpty') }}
        </p>
        <div v-else-if="stats" class="user-profile-view__stats-grid">
          <div class="user-profile-view__stat">
            <span class="user-profile-view__stat-value">{{ stats.games_played }}</span>
            <span class="user-profile-view__stat-label">{{ t('profile.statsGames') }}</span>
          </div>
          <div class="user-profile-view__stat">
            <span class="user-profile-view__stat-value">{{ stats.wins }}</span>
            <span class="user-profile-view__stat-label">{{ t('profile.statsWins') }}</span>
          </div>
          <div class="user-profile-view__stat">
            <span class="user-profile-view__stat-value">{{ stats.best_score }}</span>
            <span class="user-profile-view__stat-label">{{ t('profile.statsBest') }}</span>
          </div>
          <div class="user-profile-view__stat">
            <span class="user-profile-view__stat-value">{{ stats.total_lines }}</span>
            <span class="user-profile-view__stat-label">{{ t('profile.statsLines') }}</span>
          </div>
          <div class="user-profile-view__stat">
            <span class="user-profile-view__stat-value">{{ stats.avg_score }}</span>
            <span class="user-profile-view__stat-label">{{ t('profile.statsAvg') }}</span>
          </div>
        </div>
      </section>

      <section class="user-profile-view__panel" aria-labelledby="user-profile-badges-heading">
        <h2 id="user-profile-badges-heading" class="user-profile-view__section-title">
          {{ t('profile.badgesTitle') }}
        </h2>
        <div class="user-profile-view__badges">
          <div class="user-profile-view__badge-item" v-for="badge in unlockedBadges" :key="badge.key">
            <AchievementBadge
              :tier="badge.tier"
              :catImage="badge.catImage"
              :badgeName="badge.badgeName"
            />
            <span class="user-profile-view__badge-name">{{ badge.badgeName }}</span>
          </div>
        </div>
        <RouterLink
          :to="{ name: 'achievements', params: { id: profileUser.id } }"
          class="user-profile-view__link"
        >
          {{ t('profile.seeAllAchievements') }}
        </RouterLink>
      </section>

      <section class="user-profile-view__panel" aria-labelledby="user-profile-history-heading">
        <h2 id="user-profile-history-heading" class="user-profile-view__section-title">
          {{ t('matchHistory.title') }}
        </h2>
        <MatchHistoryList :user-id="profileUser.id" @select="openGameDetail" />
      </section>

      <GameDetailModal
        :open="gameDetailOpen"
        :game-id="selectedGameId"
        :highlight-user-id="profileUser.id"
        :highlight-label="`@${profileUser.username}`"
        @close="closeGameDetail"
      />
    </template>
  </section>
</template>
