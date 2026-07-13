<script setup lang="ts">

import { computed, nextTick, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import { resolveAssetUrl } from '@/api/client'
import * as friendsApi from '@/api/friends'
import * as chatApi from '@/api/chat'
import { useAuthStore } from '@/stores/auth'

import { useChatSocket, WS_TYPE_CHAT_MESSAGE, type WsEnvelope } from '@/composables/useChatSocket'
import type { ChatMessage, User } from '@/types/api'

const auth = useAuthStore()
const { t } = useI18n()
const chat = useChatSocket()

const friends = ref<User[]>([])
const activeFriend = ref<User | null>(null)
const messages = ref<ChatMessage[]>([])
const unread = ref<Record<string, number>>({})
const draft = ref('')
const threadRef = ref<HTMLElement | null>(null)

const myId = computed(() => auth.user?.id ?? '')

onMounted(async () => {

  const [f, u] = await Promise.all([friendsApi.getFriends(), chatApi.getUnread()])
  friends.value = f.friends
  unread.value = Object.fromEntries(u.unread.map((c) => [c.sender_id, c.count]))
  chat.connect(onEnvelope)
})

function onEnvelope(env: WsEnvelope): void {
  if (env.type !== WS_TYPE_CHAT_MESSAGE) return
  const msg = env.payload as ChatMessage
  const otherId = msg.sender_id === myId.value ? msg.recipient_id : msg.sender_id

  if (activeFriend.value && otherId === activeFriend.value.id) {
    messages.value.push(msg)
    scrollToBottom()
    if (msg.sender_id !== myId.value) void chatApi.markConversationRead(otherId)
  } else if (msg.sender_id !== myId.value) {
    unread.value[msg.sender_id] = (unread.value[msg.sender_id] ?? 0) + 1
  }
}

async function openConversation(friend: User): Promise<void> {
  activeFriend.value = friend

  const res = await chatApi.getConversation(friend.id)
  messages.value = res.messages.slice().reverse() // API returns newest-first
  unread.value[friend.id] = 0
  await chatApi.markConversationRead(friend.id)
  scrollToBottom()
}

function onSend(): void {
  const body = draft.value.trim()
  if (!body || !activeFriend.value) return
  if (chat.send(activeFriend.value.id, body)) draft.value = ''
}

function scrollToBottom(): void {
  void nextTick(() => {
    if (threadRef.value) threadRef.value.scrollTop = threadRef.value.scrollHeight
  })
}

function avatarOf(u: User): string | null {
  return resolveAssetUrl(u.avatar_url)
}
</script>

<template>
  <div class="chat-view">
    <aside class="chat-view__list">
      <h1 class="chat-view__title">{{ t('chat.title') }}</h1>
      <p v-if="friends.length === 0" class="chat-view__empty">{{ t('chat.noFriends') }}</p>
      <button
        v-for="f in friends"
        :key="f.id"

        type="button"
        class="chat-view__friend"
        :class="{ 'chat-view__friend--active': activeFriend?.id === f.id }"
        @click="openConversation(f)"
      >
        <span class="chat-view__avatar">
          <img v-if="avatarOf(f)" :src="avatarOf(f)!" alt="" />
          <span v-else>{{ f.username.slice(0, 2).toUpperCase() }}</span>
        </span>
        <span class="chat-view__name">{{ f.username }}</span>
        <span v-if="f.is_online" class="chat-view__dot" :title="t('chat.online')" />
        <span v-if="unread[f.id]" class="chat-view__badge">{{ unread[f.id] }}</span>
      </button>
    </aside>

    <section class="chat-view__pane">
      <template v-if="activeFriend">
        <header class="chat-view__pane-header">{{ activeFriend.username }}</header>
        <div ref="threadRef" class="chat-view__thread">

          <p v-if="messages.length === 0" class="chat-view__empty">{{ t('chat.sayHi') }}</p>
          <div
            v-for="m in messages"
            :key="m.id"
            class="chat-view__bubble"
            :class="m.sender_id === myId ? 'chat-view__bubble--mine' : 'chat-view__bubble--theirs'"
          >
            {{ m.body }}
          </div>
        </div>
        <form class="chat-view__composer" @submit.prevent="onSend">
          <input
            v-model="draft"
            type="text"

            maxlength="2000"
            class="chat-view__input"
            :placeholder="t('chat.placeholder')"
          />
          <button type="submit" class="chat-view__send" :disabled="!draft.trim()">
            {{ t('chat.send') }}
          </button>
        </form>
      </template>
      <p v-else class="chat-view__empty chat-view__empty--center">{{ t('chat.selectFriend') }}</p>
    </section>
  </div>
</template>

<style scoped>
.chat-view { display: flex; height: 70vh; gap: 1rem; }
.chat-view__list { width: 240px; overflow-y: auto; border-right: 1px solid rgba(255, 255, 255, 0.1); }
.chat-view__title { font-size: 1.1rem; padding: 0.5rem; }

.chat-view__friend { display: flex; align-items: center; gap: 0.5rem; width: 100%; padding: 0.5rem; background: none; border: none; color: inherit; cursor: pointer; }
.chat-view__friend--active { background: rgba(255, 255, 255, 0.08); }
.chat-view__avatar img { width: 28px; height: 28px; border-radius: 50%; object-fit: cover; }
.chat-view__name { flex: 1; text-align: left; }
.chat-view__dot { width: 8px; height: 8px; border-radius: 50%; background: #3fb950; }
.chat-view__badge { background: #f85149; color: #fff; border-radius: 999px; padding: 0 0.4rem; font-size: 0.75rem; }
.chat-view__pane { flex: 1; display: flex; flex-direction: column; }
.chat-view__pane-header { padding: 0.5rem; font-weight: 600; border-bottom: 1px solid rgba(255, 255, 255, 0.1); }
.chat-view__thread { flex: 1; overflow-y: auto; display: flex; flex-direction: column; gap: 0.4rem; padding: 0.5rem; }
.chat-view__bubble { max-width: 70%; padding: 0.4rem 0.7rem; border-radius: 0.8rem; word-break: break-word; }
.chat-view__bubble--mine { align-self: flex-end; background: #2f81f7; color: #fff; }
.chat-view__bubble--theirs { align-self: flex-start; background: rgba(255, 255, 255, 0.12); }
.chat-view__composer { display: flex; gap: 0.5rem; padding: 0.5rem; }

.chat-view__input { flex: 1; }
.chat-view__empty { opacity: 0.6; padding: 1rem; }
.chat-view__empty--center { margin: auto; }
</style>
