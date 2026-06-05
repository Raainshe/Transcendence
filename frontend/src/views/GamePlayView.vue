<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, useTemplateRef, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'

import ActionNotifications from '@/components/game/ActionNotifications.vue'
import GameBoard from '@/components/game/GameBoard.vue'
import MatchResultsOverlay from '@/components/game/MatchResultsOverlay.vue'
import OpponentBoard from '@/components/game/OpponentBoard.vue'
import { gameAudio } from '@/game/audio/AudioManager'
import GameHud from '@/components/game/GameHud.vue'
import GameModeBadge from '@/components/game/GameModeBadge.vue'
import GameModeHelpModal from '@/components/game/GameModeHelpModal.vue'
import MenuItem from '@/components/menu/MenuItem.vue'
import { readMultiplayerSeed, stashMultiplayerSeed } from '@/stores/lobby'
import { useGameSettingsStore } from '@/stores/gameSettings'
import { useGameSessionStore } from '@/stores/gameSession'
import { useMatchStore } from '@/stores/match'
import { getGameVariationInfo } from '@/types/game'

import '@/assets/styles/views/game-play-view.css'

const router = useRouter()
const route = useRoute()
const store = useGameSessionStore()
const matchStore = useMatchStore()
const { opponentList: opponents, matchEnded, matchResults, gameId: matchGameId } =
  storeToRefs(matchStore)
const settings = useGameSettingsStore()
const { t } = useI18n()

const PAUSE_ITEM_COUNT = 4
const pauseFocusedIndex = ref(0)
const showModeHelp = ref(false)
const sessionReady = ref(false)
const pauseMenuRef = useTemplateRef<HTMLElement>('pauseMenu')

const isMultiplayer = computed(() => store.multiplayerGameId != null)

const aboutMenuLabel = computed(
  () => t('game.pauseAbout', { mode: getGameVariationInfo(store.variant).label }),
)

const sfxMenuLabel = computed(() =>
  settings.sfxEnabled ? t('game.pauseSfxOn') : t('game.pauseSfxOff'),
)

function toggleSfx(): void {
  gameAudio.unlock()
  settings.sfxEnabled = !settings.sfxEnabled
  gameAudio.setEnabled(settings.sfxEnabled)
}

watch(
  () => store.paused,
  (paused) => {
    if (paused) {
      showModeHelp.value = false
      pauseFocusedIndex.value = 0
      void nextTick(() => {
        pauseMenuRef.value?.focus()
      })
      return
    }
    showModeHelp.value = false
  },
)

function movePauseFocus(delta: 1 | -1): void {
  pauseFocusedIndex.value =
    (pauseFocusedIndex.value + delta + PAUSE_ITEM_COUNT) % PAUSE_ITEM_COUNT
}

function openModeHelp(): void {
  showModeHelp.value = true
}

function closeModeHelp(): void {
  showModeHelp.value = false
  void nextTick(() => {
    pauseMenuRef.value?.focus()
  })
}

function activatePauseMenu(): void {
  switch (pauseFocusedIndex.value) {
    case 0:
      store.resume()
      break
    case 1:
      toggleSfx()
      break
    case 2:
      openModeHelp()
      break
    case 3:
      void router.push({ name: 'home' })
      break
    default:
      break
  }
}

function onPauseMenuKeydown(event: KeyboardEvent): void {
  if (showModeHelp.value) return

  switch (event.key) {
    case 'ArrowUp':
      event.preventDefault()
      event.stopPropagation()
      movePauseFocus(-1)
      break
    case 'ArrowDown':
      event.preventDefault()
      event.stopPropagation()
      movePauseFocus(1)
      break
    case 'Enter':
    case ' ':
      event.preventDefault()
      event.stopPropagation()
      activatePauseMenu()
      break
    default:
      break
  }
}

function onGlobalKeydown(e: KeyboardEvent): void {
  gameAudio.unlock()
  if ((e.key === 'm' || e.key === 'M') && store.paused && !showModeHelp.value) {
    e.preventDefault()
    toggleSfx()
    return
  }
  if (e.key === 'Escape') {
    e.preventDefault()
    if (showModeHelp.value) {
      closeModeHelp()
      return
    }
    if (!store.paused) {
      store.pause()
    } else {
      void router.push({ name: 'home' })
    }
    return
  }
  if (e.key === 'p' || e.key === 'P') {
    if (store.paused && !showModeHelp.value) {
      e.preventDefault()
      store.resume()
    }
  }
}

async function initSession(): Promise<void> {
  const matchId = typeof route.query.match === 'string' ? route.query.match : null
  try {
    if (matchId) {
      const stashedSeed = readMultiplayerSeed(matchId)
      await matchStore.bootstrap(matchId)
      matchStore.ensurePlayersRoster(matchId)
      const seed = stashedSeed ?? matchStore.getSharedSeed()
      if (!seed || seed <= 0) {
        void router.replace({ name: 'home' })
        return
      }
      stashMultiplayerSeed(matchId, seed)
      store.beginMultiplayerMatch(matchId, seed)
      matchStore.connect(matchId)
      sessionReady.value = true
    } else {
      store.beginSession()
      sessionReady.value = true
    }
  } catch {
    void router.replace({ name: 'home' })
  }
}

onMounted(() => {
  window.addEventListener('keydown', onGlobalKeydown)
  void initSession()
})

onBeforeUnmount(() => {
  window.removeEventListener('keydown', onGlobalKeydown)
  matchStore.disconnect()
  store.endSession()
})
</script>

<template>
  <div class="game-play-view">
    <h1 class="visually-hidden">{{ t('game.playHeading') }}</h1>
    <GameHud band="top" />
    <div class="game-play-view__arena">
      <div class="game-play-view__canvas-slot">
        <GameBoard v-if="sessionReady" />
        <ActionNotifications />
      </div>
      <aside
        v-if="sessionReady && isMultiplayer && opponents.length > 0"
        class="game-play-view__opponents"
        :aria-label="t('game.opponents.title')"
      >
        <OpponentBoard
          v-for="opp in opponents"
          :key="opp.user_id"
          :username="opp.username"
          :avatar-url="opp.avatar_url"
          :score="opp.score"
          :lines="opp.lines"
          :level="opp.level"
          :alive="opp.alive"
          :board-b64="opp.board"
        />
      </aside>
    </div>
    <GameHud band="bottom">
      <p class="game-play-view__controls">
        {{ t('game.controlsLine') }}
      </p>
    </GameHud>
    <MatchResultsOverlay
      v-if="matchEnded && matchResults && matchGameId"
      :game-id="matchGameId"
      :results="matchResults"
    />
    <div
      v-if="store.paused && !matchEnded"
      class="game-play-view__pause-overlay"
      role="dialog"
      aria-modal="true"
      aria-labelledby="pause-title"
    >
      <div class="game-play-view__pause-panel">
        <p id="pause-title" class="game-play-view__pause-title">{{ t('game.pauseTitle') }}</p>
        <div class="game-play-view__pause-mode">
          <GameModeBadge :variation="store.variant" size="md" />
        </div>
        <p class="game-play-view__pause-copy">{{ t('game.pauseCopy') }}</p>
        <section
          ref="pauseMenu"
          class="game-play-view__pause-menu"
          role="menu"
          :aria-label="t('game.pauseMenuAriaLabel')"
          tabindex="0"
          @keydown="onPauseMenuKeydown"
        >
          <ul class="game-play-view__pause-list">
            <MenuItem
              :label="t('game.pauseResume')"
              kind="action"
              :selected="pauseFocusedIndex === 0"
              @select="pauseFocusedIndex = 0"
              @activate="store.resume()"
            />
            <MenuItem
              :label="sfxMenuLabel"
              kind="action"
              :selected="pauseFocusedIndex === 1"
              @select="pauseFocusedIndex = 1"
              @activate="toggleSfx()"
            />
            <MenuItem
              :label="aboutMenuLabel"
              kind="action"
              :selected="pauseFocusedIndex === 2"
              @select="pauseFocusedIndex = 2"
              @activate="openModeHelp()"
            />
            <MenuItem
              :label="t('game.pauseQuit')"
              kind="action"
              :selected="pauseFocusedIndex === 3"
              @select="pauseFocusedIndex = 3"
              @activate="void router.push({ name: 'home' })"
            />
          </ul>
        </section>
      </div>
    </div>
    <GameModeHelpModal
      :variation="store.variant"
      :open="showModeHelp && store.paused"
      @close="closeModeHelp"
    />
  </div>
</template>

<style scoped>
.visually-hidden {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}
</style>
