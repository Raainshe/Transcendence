<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { RouterLink } from 'vue-router'

import { resolveAssetUrl } from '@/api/client'
import type { User } from '@/types/api'

const props = defineProps<{
  user: User
  profileTo?: { name: 'userProfile'; params: { id: string } }
  youLabel?: string
}>()

const { t } = useI18n()

const avatarSrc = computed(() => resolveAssetUrl(props.user.avatar_url))
const initials = computed(() => props.user.username.slice(0, 2).toUpperCase())
const onlineLabel = computed(() =>
  props.user.is_online ? t('friends.online') : t('friends.offline'),
)
</script>

<template>
  <li class="friends-user-row">
    <div class="friends-user-row__identity">
      <img
        v-if="avatarSrc"
        :src="avatarSrc"
        :alt="t('friends.avatarAlt', { username: user.username })"
        class="friends-user-row__avatar"
        width="40"
        height="40"
      />
      <span
        v-else
        class="friends-user-row__avatar friends-user-row__avatar--placeholder"
        aria-hidden="true"
      >
        {{ initials }}
      </span>
      <div class="friends-user-row__meta">
        <RouterLink
          v-if="profileTo"
          :to="profileTo"
          class="friends-user-row__username friends-user-row__username--link"
        >
          @{{ user.username }}
        </RouterLink>
        <span v-else class="friends-user-row__username">@{{ user.username }}</span>
        <span v-if="youLabel" class="friends-user-row__you">{{ youLabel }}</span>
        <span
          class="friends-user-row__online"
          :class="{ 'friends-user-row__online--active': user.is_online }"
        >
          {{ onlineLabel }}
        </span>
      </div>
    </div>
    <div v-if="$slots.actions" class="friends-user-row__actions">
      <slot name="actions" />
    </div>
  </li>
</template>
