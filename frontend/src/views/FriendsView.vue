<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import * as friendsApi from '@/api/friends'
import * as usersApi from '@/api/users'
import UserRow from '@/components/friends/UserRow.vue'
import { ApiError } from '@/api/client'
import { useAuthStore } from '@/stores/auth'
import type { User } from '@/types/api'

import '@/assets/styles/views/friends-view.css'

type FriendsTab = 'friends' | 'requests' | 'blocked' | 'add'

const { t } = useI18n()
const auth = useAuthStore()

const activeTab = ref<FriendsTab>('friends')
const friends = ref<User[]>([])
const requests = ref<User[]>([])
const blocked = ref<User[]>([])

const loading = ref(true)
const loadError = ref(false)
const actionError = ref<string | null>(null)
const actionBusyId = ref<string | null>(null)

const addUsername = ref('')
const addLoading = ref(false)
const addError = ref<string | null>(null)
const addSuccess = ref(false)

const tabs = computed(() => [
  { id: 'friends' as const, label: t('friends.tabFriends') },
  {
    id: 'requests' as const,
    label: t('friends.tabRequests'),
    badge: requests.value.length > 0 ? requests.value.length : undefined,
  },
  { id: 'blocked' as const, label: t('friends.tabBlocked') },
  { id: 'add' as const, label: t('friends.tabAdd') },
])

onMounted(() => {
  void loadAll()
})

function mapError(error: unknown): string {
  if (error instanceof ApiError) {
    const known: Record<string, string> = {
      'user not found': 'friends.errors.userNotFound',
      'cannot add yourself as a friend': 'friends.errors.cannotAddSelf',
      'relationship already exists': 'friends.errors.relationshipExists',
      'friend request not found': 'friends.errors.requestNotFound',
      'friendship not found': 'friends.errors.friendshipNotFound',
      'user is already blocked': 'friends.errors.alreadyBlocked',
      'user is not blocked': 'friends.errors.notBlocked',
      'cannot block yourself': 'friends.errors.cannotBlockSelf',
    }
    const mapped = known[error.message.toLowerCase()]
    if (mapped) return t(mapped)
    if (error.message) return error.message
  }
  return t('friends.errors.generic')
}

function normalizeUsername(raw: string): string {
  return raw.trim().replace(/^@+/, '')
}

async function loadAll(): Promise<void> {
  loading.value = true
  loadError.value = false
  actionError.value = null
  try {
    const [friendsRes, requestsRes, blockedRes] = await Promise.all([
      friendsApi.getFriends(),
      friendsApi.getFriendRequests(),
      friendsApi.getBlockedUsers(),
    ])
    friends.value = friendsRes.friends
    requests.value = requestsRes.requests
    blocked.value = blockedRes.blocked
  } catch {
    friends.value = []
    requests.value = []
    blocked.value = []
    loadError.value = true
  } finally {
    loading.value = false
  }
}

async function runAction(userId: string, fn: () => Promise<void>): Promise<void> {
  actionError.value = null
  actionBusyId.value = userId
  try {
    await fn()
    await loadAll()
  } catch (error) {
    actionError.value = mapError(error)
  } finally {
    actionBusyId.value = null
  }
}

function setTab(tab: FriendsTab): void {
  activeTab.value = tab
  actionError.value = null
  addError.value = null
  addSuccess.value = false
}

async function onAddFriend(): Promise<void> {
  addError.value = null
  addSuccess.value = false
  const name = normalizeUsername(addUsername.value)
  if (!name) {
    addError.value = t('friends.errors.usernameRequired')
    return
  }

  addLoading.value = true
  try {
    const target = await usersApi.findUserByUsername(name)
    if (!target) {
      addError.value = t('friends.errors.userNotFound')
      return
    }
    if (auth.user?.id === target.id) {
      addError.value = t('friends.errors.cannotAddSelf')
      return
    }
    await friendsApi.sendFriendRequest(target.id)
    addUsername.value = ''
    addSuccess.value = true
    await loadAll()
    activeTab.value = 'friends'
  } catch (error) {
    addError.value = mapError(error)
  } finally {
    addLoading.value = false
  }
}
</script>

<template>
  <section class="friends-view">
    <header class="friends-view__header">
      <h1 class="friends-view__title">{{ t('friends.title') }}</h1>
      <p class="friends-view__subtitle">{{ t('friends.subtitle') }}</p>
    </header>

    <div class="friends-view__panel">
      <nav class="friends-view__tabs" :aria-label="t('friends.tabsAria')">
        <button
          v-for="tab in tabs"
          :key="tab.id"
          type="button"
          class="friends-view__tab"
          :class="{ 'friends-view__tab--active': activeTab === tab.id }"
          :aria-current="activeTab === tab.id ? 'page' : undefined"
          @click="setTab(tab.id)"
        >
          {{ tab.label }}
          <span v-if="tab.badge" class="friends-view__tab-badge">{{ tab.badge }}</span>
        </button>
      </nav>

      <p v-if="loading" class="friends-view__status">{{ t('friends.loading') }}</p>
      <p v-else-if="loadError" class="friends-view__status">{{ t('friends.loadError') }}</p>

      <template v-else>
        <p v-if="actionError" class="friends-view__error" role="alert">{{ actionError }}</p>

        <div v-if="activeTab === 'friends'">
          <p v-if="friends.length === 0" class="friends-view__status">
            {{ t('friends.emptyFriends') }}
          </p>
          <ul v-else class="friends-view__list">
            <UserRow
              v-for="friend in friends"
              :key="friend.id"
              :user="friend"
              :profile-to="{ name: 'userProfile', params: { id: friend.id } }"
            >
              <template #actions>
                <button
                  type="button"
                  class="friends-user-row__action friends-user-row__action--danger"
                  :disabled="actionBusyId === friend.id"
                  @click="runAction(friend.id, () => friendsApi.removeFriend(friend.id))"
                >
                  {{ t('friends.remove') }}
                </button>
                <button
                  type="button"
                  class="friends-user-row__action friends-user-row__action--danger"
                  :disabled="actionBusyId === friend.id"
                  @click="runAction(friend.id, () => friendsApi.blockUser(friend.id))"
                >
                  {{ t('friends.block') }}
                </button>
              </template>
            </UserRow>
          </ul>
        </div>

        <div v-if="activeTab === 'requests'">
          <p class="friends-view__hint">{{ t('friends.requestsHint') }}</p>
          <p v-if="requests.length === 0" class="friends-view__status">
            {{ t('friends.emptyRequests') }}
          </p>
          <ul v-else class="friends-view__list">
            <UserRow
              v-for="requester in requests"
              :key="requester.id"
              :user="requester"
              :profile-to="{ name: 'userProfile', params: { id: requester.id } }"
            >
              <template #actions>
                <button
                  type="button"
                  class="friends-user-row__action friends-user-row__action--primary"
                  :disabled="actionBusyId === requester.id"
                  @click="
                    runAction(requester.id, () => friendsApi.acceptFriendRequest(requester.id))
                  "
                >
                  {{ t('friends.accept') }}
                </button>
                <button
                  type="button"
                  class="friends-user-row__action"
                  :disabled="actionBusyId === requester.id"
                  @click="runAction(requester.id, () => friendsApi.removeFriend(requester.id))"
                >
                  {{ t('friends.decline') }}
                </button>
              </template>
            </UserRow>
          </ul>
        </div>

        <div v-if="activeTab === 'blocked'">
          <p v-if="blocked.length === 0" class="friends-view__status">
            {{ t('friends.emptyBlocked') }}
          </p>
          <ul v-else class="friends-view__list">
            <UserRow
              v-for="blockedUser in blocked"
              :key="blockedUser.id"
              :user="blockedUser"
              :profile-to="{ name: 'userProfile', params: { id: blockedUser.id } }"
            >
              <template #actions>
                <button
                  type="button"
                  class="friends-user-row__action friends-user-row__action--primary"
                  :disabled="actionBusyId === blockedUser.id"
                  @click="runAction(blockedUser.id, () => friendsApi.unblockUser(blockedUser.id))"
                >
                  {{ t('friends.unblock') }}
                </button>
              </template>
            </UserRow>
          </ul>
        </div>

        <div v-if="activeTab === 'add'">
          <form class="friends-view__add-form" @submit.prevent="onAddFriend">
            <div class="friends-view__field">
              <label class="friends-view__label" for="friends-add-username">
                {{ t('friends.addLabel') }}
              </label>
              <input
                id="friends-add-username"
                v-model="addUsername"
                type="text"
                class="friends-view__input"
                :placeholder="t('friends.addPlaceholder')"
                autocomplete="username"
                :disabled="addLoading"
              />
            </div>
            <p class="friends-view__hint">{{ t('friends.addHint') }}</p>
            <p v-if="addError" class="friends-view__error" role="alert">{{ addError }}</p>
            <p v-if="addSuccess" class="friends-view__success">{{ t('friends.addSuccess') }}</p>
            <button type="submit" class="friends-view__button" :disabled="addLoading">
              {{ addLoading ? t('friends.addSending') : t('friends.addSubmit') }}
            </button>
          </form>
        </div>
      </template>
    </div>
  </section>
</template>
