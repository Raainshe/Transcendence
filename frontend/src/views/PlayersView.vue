<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import * as usersApi from '@/api/users'
import UserRow from '@/components/friends/UserRow.vue'
import { useAuthStore } from '@/stores/auth'
import type { User } from '@/types/api'

import '@/assets/styles/views/players-view.css'

const PAGE_SIZE = 20

const { t } = useI18n()
const auth = useAuthStore()

const users = ref<User[]>([])
const total = ref(0)
const page = ref(0)
const loading = ref(true)
const loadError = ref(false)

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / PAGE_SIZE)))
const canPrev = computed(() => page.value > 0)
const canNext = computed(() => (page.value + 1) * PAGE_SIZE < total.value)

type PlayerRow = { user: User; youLabel?: string }

const visibleUsers = computed((): PlayerRow[] => {
  const me = auth.user?.id
  return users.value.map((u) => ({
    user: u,
    youLabel: me && u.id === me ? t('players.you') : undefined,
  }))
})

onMounted(() => {
  void load()
})

async function load(): Promise<void> {
  loading.value = true
  loadError.value = false
  try {
    const res = await usersApi.listUsers({
      limit: PAGE_SIZE,
      offset: page.value * PAGE_SIZE,
    })
    users.value = res.users
    total.value = res.total
  } catch {
    users.value = []
    total.value = 0
    loadError.value = true
  } finally {
    loading.value = false
  }
}

function profileRoute(id: string) {
  return { name: 'userProfile' as const, params: { id } }
}

function prevPage(): void {
  if (!canPrev.value) return
  page.value -= 1
  void load()
}

function nextPage(): void {
  if (!canNext.value) return
  page.value += 1
  void load()
}
</script>

<template>
  <section class="players-view">
    <header class="players-view__header">
      <h1 class="players-view__title">{{ t('players.title') }}</h1>
      <p class="players-view__subtitle">{{ t('players.subtitle') }}</p>
    </header>

    <div class="players-view__panel">
      <p v-if="loading" class="players-view__status">{{ t('players.loading') }}</p>
      <p v-else-if="loadError" class="players-view__status">{{ t('players.loadError') }}</p>
      <p v-else-if="users.length === 0" class="players-view__status">{{ t('players.empty') }}</p>

      <template v-else>
        <ul class="players-view__list">
          <UserRow
            v-for="{ user, youLabel } in visibleUsers"
            :key="user.id"
            :user="user"
            :profile-to="profileRoute(user.id)"
            :you-label="youLabel"
          />
        </ul>

        <div v-if="total > PAGE_SIZE" class="players-view__pagination">
          <button
            type="button"
            class="players-view__page-btn"
            :disabled="!canPrev || loading"
            @click="prevPage"
          >
            {{ t('players.prev') }}
          </button>
          <span class="players-view__page-info">
            {{ t('players.page', { current: page + 1, total: totalPages }) }}
          </span>
          <button
            type="button"
            class="players-view__page-btn"
            :disabled="!canNext || loading"
            @click="nextPage"
          >
            {{ t('players.next') }}
          </button>
        </div>
      </template>
    </div>
  </section>
</template>
