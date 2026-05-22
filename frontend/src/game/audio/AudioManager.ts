import type { EngineEvent } from '@/game/types'

import {
  LINE_CLEAR_PITCH,
  MENU_MUSIC_VOLUME,
  SOUND_IDS,
  SOUND_PATHS,
  THEME_PATH,
  type SoundId,
} from '@/game/audio/sounds'

const MOVE_THROTTLE_MS = 50
const SOFT_DROP_THROTTLE_MS = 40

export type AudioSettings = {
  sfxEnabled: boolean
  sfxVolume: number
}

export class AudioManager {
  private readonly clips = new Map<SoundId, HTMLAudioElement>()
  private readonly loaded = new Set<SoundId>()
  private unlocked = false
  private enabled = true
  private volume = 0.7
  private lastMoveAt = -MOVE_THROTTLE_MS
  private lastSoftDropAt = -SOFT_DROP_THROTTLE_MS
  private menuMusic: HTMLAudioElement | null = null
  private musicEnabled = true

  preload(): void {
    if (typeof Audio === 'undefined') return
    for (const id of SOUND_IDS) {
      const audio = new Audio(SOUND_PATHS[id])
      audio.preload = 'auto'
      audio.volume = this.volume
      audio.addEventListener(
        'canplaythrough',
        () => {
          this.loaded.add(id)
        },
        { once: true },
      )
      audio.addEventListener('error', () => {
        this.clips.delete(id)
      })
      this.clips.set(id, audio)
    }
  }

  unlock(): void {
    this.unlocked = true
    if (this.musicEnabled) {
      this.startMenuMusic()
    }
  }

  setMusicEnabled(on: boolean): void {
    this.musicEnabled = on
    if (!on) {
      this.stopMenuMusic()
      return
    }
    if (this.unlocked) {
      this.startMenuMusic()
    }
  }

  startMenuMusic(): void {
    if (!this.musicEnabled || typeof Audio === 'undefined') return
    if (!this.menuMusic) {
      this.menuMusic = new Audio(THEME_PATH)
      this.menuMusic.loop = true
      this.menuMusic.volume = MENU_MUSIC_VOLUME
      this.menuMusic.preload = 'auto'
    }
    if (!this.unlocked) return
    void this.menuMusic.play().catch(() => {
      /* autoplay policy */
    })
  }

  stopMenuMusic(): void {
    if (!this.menuMusic) return
    this.menuMusic.pause()
    this.menuMusic.currentTime = 0
  }

  isMusicPlaying(): boolean {
    return this.menuMusic !== null && !this.menuMusic.paused
  }

  /** Menu UI sounds (move / selected); respects menu SFX prefs, not in-game `enabled`. */
  playUi(
    id: SoundId,
    ui: { sfxEnabled: boolean; sfxVolume: number },
    opts?: { playbackRate?: number; volumeScale?: number },
  ): void {
    if (!ui.sfxEnabled || !this.unlocked) return
    const base = this.clips.get(id)
    if (!base) return

    const clone = base.cloneNode(true) as HTMLAudioElement
    clone.volume = ui.sfxVolume * (opts?.volumeScale ?? 1)
    if (opts?.playbackRate !== undefined) {
      clone.playbackRate = opts.playbackRate
    }
    clone.play().catch(() => {
      /* missing asset or autoplay */
    })
  }

  setEnabled(value: boolean): void {
    this.enabled = value
  }

  setVolume(value: number): void {
    this.volume = Math.max(0, Math.min(1, value))
    for (const clip of this.clips.values()) {
      clip.volume = this.volume
    }
  }

  play(id: SoundId, opts?: { playbackRate?: number; volumeScale?: number }): void {
    if (!this.enabled || !this.unlocked) return
    const base = this.clips.get(id)
    if (!base) return

    const clone = base.cloneNode(true) as HTMLAudioElement
    clone.volume = this.volume * (opts?.volumeScale ?? 1)
    if (opts?.playbackRate !== undefined) {
      clone.playbackRate = opts.playbackRate
    }
    clone.play().catch(() => {
      /* autoplay or missing asset */
    })
  }

  private playLineClear(linesCleared: number, _tSpinKind: 'none' | 'mini' | 'full'): void {
    if (!this.loaded.has('line_clear')) return

    const n = Math.min(4, Math.max(1, linesCleared)) as 1 | 2 | 3 | 4
    this.play('line_clear', { playbackRate: LINE_CLEAR_PITCH[n] })
  }

  handleEvents(events: readonly EngineEvent[], settings: AudioSettings): void {
    this.enabled = settings.sfxEnabled
    this.volume = settings.sfxVolume
    this.setVolume(settings.sfxVolume)

    if (!this.enabled) return

    const hasLineClear = events.some((e) => e.type === 'lines-cleared')

    for (const event of events) {
      switch (event.type) {
        case 'piece-moved': {
          const now = performance.now()
          if (now - this.lastMoveAt < MOVE_THROTTLE_MS) break
          this.lastMoveAt = now
          this.play('move', { volumeScale: 0.35 })
          break
        }
        case 'piece-rotated':
          this.play('rotate')
          break
        case 'piece-soft-dropped': {
          const now = performance.now()
          if (now - this.lastSoftDropAt < SOFT_DROP_THROTTLE_MS) break
          this.lastSoftDropAt = now
          this.play('soft_drop', { volumeScale: 0.4 })
          break
        }
        case 'piece-hard-dropped':
          this.play('hard_drop')
          break
        case 'piece-held':
          this.play('hold')
          break
        case 'piece-locked':
          if (!hasLineClear) this.play('lock')
          break
        case 'lines-cleared':
          this.playLineClear(event.linesCleared, event.tSpinKind)
          break
        case 'score-awarded':
          if (
            event.breakdown.backToBackMultiplier > 1 &&
            (event.breakdown.reason === 'lineClear' || event.breakdown.reason === 'tSpinNoLines')
          ) {
            this.play('back_to_back', { volumeScale: 0.85 })
          }
          break
        case 'level-up':
          this.play('level_up')
          break
        case 'game-over':
          this.play('game_over')
          break
        case 'match-ended':
          if (event.kind === 'won') this.play('win')
          break
        default:
          break
      }
    }
  }
}

/** Shared instance for the play session. */
export const gameAudio = new AudioManager()

if (typeof window !== 'undefined') {
  gameAudio.preload()
}
