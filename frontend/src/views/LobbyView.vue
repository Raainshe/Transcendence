<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'

import { resolveAssetUrl } from '@/api/client'
import {
  WS_TYPE_LOBBY_CLOSED,
  useLobbySocket,
} from '@/composables/useLobbySocket'
import { useGameSessionStore } from '@/stores/gameSession'
import { useLobbyStore } from '@/stores/lobby'
import { useAuthStore } from '@/stores/auth'
import type { LobbyMember, StartLobbyResult } from '@/types/api'

import '@/assets/styles/views/lobby-view.css'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const auth = useAuthStore()
const lobbyStore = useLobbyStore()
const gameSession = useGameSessionStore()
const { connect, disconnect } = useLobbySocket()

const copySuccess = ref(false)
const closedMessage = ref<string | null>(null)
let navigatingToMatch = false

const lobbyId = computed(() => String(route.params.id ?? ''))

const statusLabel = computed(() => {
  const status = lobbyStore.lobby?.status
  if (status === 'waiting') return t('lobby.statusWaiting')
  if (status === 'closed') return t('lobby.statusClosed')
  return ''
})

const readyButtonLabel = computed(() =>
  lobbyStore.myMember?.is_ready ? t('lobby.notReady') : t('lobby.ready'),
)

function memberInitials(member: LobbyMember): string {
  return member.username.slice(0, 2).toUpperCase()
}

function memberAvatar(member: LobbyMember): string | null {
  return resolveAssetUrl(member.avatar_url)
}

function isYou(member: LobbyMember): boolean {
  return member.user_id === auth.user?.id
}

function isHostMember(member: LobbyMember): boolean {
  return member.user_id === lobbyStore.lobby?.host_user_id
}

function goToMatch(result: StartLobbyResult): void {
  if (navigatingToMatch) return
  navigatingToMatch = true
  gameSession.beginMultiplayerMatch(result.game_id, result.shared_seed)
  void router.push({ name: 'play', query: { match: result.game_id } })
}

function handleWsEnvelope(env: { type: string; payload?: unknown }): void {
  const matchResult = lobbyStore.applyWsEnvelope(env)
  if (matchResult) {
    goToMatch(matchResult)
    return
  }
  if (env.type === WS_TYPE_LOBBY_CLOSED && lobbyStore.closedReason) {
    closedMessage.value =
      lobbyStore.closedReason === 'host_left'
        ? t('lobby.closedHostLeft')
        : t('lobby.closedStarted')
    void router.push({ name: 'home' })
  }
}

async function copyInviteCode(): Promise<void> {
  const code = lobbyStore.lobby?.invite_code
  if (!code || typeof navigator.clipboard === 'undefined') return
  try {
    await navigator.clipboard.writeText(code)
    copySuccess.value = true
    window.setTimeout(() => {
      copySuccess.value = false
    }, 2000)
  } catch {
    /* clipboard denied */
  }
}

async function handleToggleReady(): Promise<void> {
  await lobbyStore.toggleReady()
}

async function handleStart(): Promise<void> {
  const result = await lobbyStore.start()
  if (result) goToMatch(result)
}

async function handleLeave(): Promise<void> {
  await lobbyStore.leave()
  void router.push({ name: 'home' })
}

onMounted(async () => {
  const id = lobbyId.value
  if (!id) {
    void router.replace({ name: 'home' })
    return
  }
  const loaded = await lobbyStore.load(id)
  if (!loaded) {
    return
  }
  if (lobbyStore.lobby?.status === 'closed' && lobbyStore.lobby.game_id) {
    const seed = lobbyStore.lobby.shared_seed
    if (seed) {
      goToMatch({
        game_id: lobbyStore.lobby.game_id,
        shared_seed: seed,
        players: [],
      })
      return
    }
  }
  connect(id, handleWsEnvelope)
})

onBeforeUnmount(() => {
  disconnect()
  lobbyStore.reset()
})
</script>

<template>
  <div class="lobby-view">
    <header class="lobby-view__header">
      <h1 class="lobby-view__title">{{ t('lobby.title') }}</h1>
      <p v-if="lobbyStore.lobby" class="lobby-view__subtitle">
        {{ t('lobby.subtitle', { count: lobbyStore.lobby.members.length, max: lobbyStore.lobby.max_players }) }}
        · {{ statusLabel }}
      </p>
    </header>

    <p v-if="closedMessage" class="lobby-view__notice" role="status">{{ closedMessage }}</p>
    <p v-if="lobbyStore.loading" class="lobby-view__status">{{ t('lobby.loading') }}</p>
    <p v-else-if="lobbyStore.error" class="lobby-view__error">{{ lobbyStore.error }}</p>

    <section v-else-if="lobbyStore.lobby" class="lobby-view__panel">
      <div class="lobby-view__invite">
        <span class="lobby-view__invite-label">{{ t('lobby.inviteCode') }}</span>
        <div class="lobby-view__invite-row">
          <code class="lobby-view__code">{{ lobbyStore.lobby.invite_code }}</code>
          <button
            type="button"
            class="lobby-view__button lobby-view__button--secondary"
            @click="copyInviteCode"
          >
            {{ copySuccess ? t('lobby.copied') : t('lobby.copyCode') }}
          </button>
        </div>
        <p class="lobby-view__hint">{{ t('lobby.inviteHint') }}</p>
      </div>

      <ul class="lobby-view__list" :aria-label="t('lobby.rosterAria')">
        <li
          v-for="member in lobbyStore.lobby.members"
          :key="member.user_id"
          class="lobby-view__member"
        >
          <img
            v-if="memberAvatar(member)"
            :src="memberAvatar(member)!"
            :alt="t('lobby.avatarAlt', { username: member.username })"
            class="lobby-view__avatar"
            width="40"
            height="40"
          />
          <span v-else class="lobby-view__avatar lobby-view__avatar--placeholder" aria-hidden="true">
            {{ memberInitials(member) }}
          </span>
          <div class="lobby-view__member-meta">
            <span class="lobby-view__username">@{{ member.username }}</span>
            <span v-if="isYou(member)" class="lobby-view__badge">{{ t('lobby.you') }}</span>
            <span v-if="isHostMember(member)" class="lobby-view__badge lobby-view__badge--host">
              {{ t('lobby.host') }}
            </span>
          </div>
          <span
            class="lobby-view__ready"
            :class="{ 'lobby-view__ready--active': member.is_ready }"
          >
            {{ member.is_ready ? t('lobby.readyLabel') : t('lobby.notReadyLabel') }}
          </span>
        </li>
      </ul>

      <div class="lobby-view__actions">
        <button
          type="button"
          class="lobby-view__button"
          :disabled="lobbyStore.actionBusy || lobbyStore.lobby.status !== 'waiting'"
          @click="handleToggleReady"
        >
          {{ readyButtonLabel }}
        </button>
        <button
          v-if="lobbyStore.isHost"
          type="button"
          class="lobby-view__button lobby-view__button--primary"
          :disabled="!lobbyStore.canStart || lobbyStore.actionBusy"
          @click="handleStart"
        >
          {{ t('lobby.start') }}
        </button>
        <button
          type="button"
          class="lobby-view__button lobby-view__button--danger"
          :disabled="lobbyStore.actionBusy"
          @click="handleLeave"
        >
          {{ t('lobby.leave') }}
        </button>
      </div>
      <p v-if="lobbyStore.isHost && !lobbyStore.canStart" class="lobby-view__hint">
        {{ t('lobby.startHint') }}
      </p>
    </section>
  </div>
</template>
