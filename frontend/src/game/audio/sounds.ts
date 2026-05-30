/** Logical sound ids mapped to `/audio/*.ogg` in `public/audio/`. */
export const SOUND_IDS = [
  'move',
  'rotate',
  'soft_drop',
  'hard_drop',
  'lock',
  'line_clear',
  'hold',
  'level_up',
  'back_to_back',
  'game_over',
  'win',
  'selected',
] as const

export type SoundId = (typeof SOUND_IDS)[number]

export const SOUND_PATHS: Record<SoundId, string> = {
  move: '/audio/move.ogg',
  rotate: '/audio/rotate.ogg',
  soft_drop: '/audio/soft_drop.ogg',
  hard_drop: '/audio/hard_drop.ogg',
  lock: '/audio/lock.ogg',
  line_clear: '/audio/line_clear.ogg',
  hold: '/audio/hold.ogg',
  level_up: '/audio/level_up.ogg',
  back_to_back: '/audio/back_to_back.ogg',
  game_over: '/audio/game_over.ogg',
  win: '/audio/win.ogg',
  selected: '/audio/selected.ogg',
}

/** Menu loop — not played via `play()` clones. */
export const THEME_PATH = '/audio/theme.ogg'

export const MENU_MUSIC_VOLUME = 0.4

/** Playback rate when using shared `line_clear` fallback. */
export const LINE_CLEAR_PITCH: Record<1 | 2 | 3 | 4, number> = {
  1: 1,
  2: 1.08,
  3: 1.16,
  4: 1.28,
}
