/**
 * DOM keyboard adapter: forwards `keydown` / `keyup` on an event target into
 * logical `InputCommand`s on an `InputController`.
 *
 * The adapter is intentionally minimal so it stays trivially testable in
 * Node (we duck-type the event target rather than depending on browser-only
 * globals). Slice 5 (the renderer) will simply instantiate it with `window`
 * and `document`; tests construct it with a fake target.
 *
 * Stuck-key safety: on `blur` (target) and `visibilitychange` to `hidden`
 * (document) the adapter calls `controller.releaseAll()`. Without this,
 * alt-tabbing while holding a key would never deliver the `keyup` and the
 * piece would auto-shift forever.
 */

import { InputCommand, type KeyMap } from '@/game/types'

/**
 * Minimal `InputController` interface the adapter calls into. Lets tests
 * pass a tiny spy without constructing a real controller + engine.
 */
export interface ControllerLike {
  press(cmd: InputCommand): void
  release(cmd: InputCommand): void
  releaseAll(): void
}

/** Default keyboard layout. Keyed on `KeyboardEvent.code` (layout-agnostic). */
export const DEFAULT_KEYMAP: KeyMap = {
  ArrowLeft: InputCommand.MoveLeft,
  ArrowRight: InputCommand.MoveRight,
  ArrowDown: InputCommand.SoftDrop,
  ArrowUp: InputCommand.RotateCW,
  Space: InputCommand.HardDrop,
  KeyC: InputCommand.Hold,
  KeyZ: InputCommand.RotateCCW,
  KeyX: InputCommand.RotateCW,
  ControlLeft: InputCommand.RotateCCW,
}

/** Subset of `EventTarget` we actually use — keeps fakes ergonomic. */
export type KeyTargetLike = Pick<EventTarget, 'addEventListener' | 'removeEventListener'>

/** Visibility target also exposes `visibilityState` (read on the change event). */
export type VisibilityTargetLike = KeyTargetLike & {
  readonly visibilityState?: string
}

export type AttachOptions = {
  /** Receives `keydown` / `keyup` / `blur`. Defaults to `window` in a browser. */
  keyTarget?: KeyTargetLike
  /** Receives `visibilitychange`. Defaults to `document` in a browser. */
  visibilityTarget?: VisibilityTargetLike
}

export class KeyboardAdapter {
  private readonly controller: ControllerLike
  private readonly keymap: KeyMap

  private keyTarget: KeyTargetLike | null = null
  private visibilityTarget: VisibilityTargetLike | null = null

  constructor(controller: ControllerLike, keymap: KeyMap = DEFAULT_KEYMAP) {
    this.controller = controller
    this.keymap = keymap
  }

  /**
   * Wires up the adapter. Safe to call repeatedly — re-attaching detaches
   * the previous targets first. With no `opts`, defaults to `window` and
   * `document` if those globals exist; otherwise throws (callers in Node
   * must pass explicit targets).
   */
  attach(opts: AttachOptions = {}): void {
    if (this.keyTarget) this.detach()

    const keyTarget =
      opts.keyTarget ?? (typeof window !== 'undefined' ? (window as KeyTargetLike) : null)
    if (!keyTarget) {
      throw new Error(
        'KeyboardAdapter.attach: no `keyTarget` provided and no global `window` is available.',
      )
    }
    const visibilityTarget =
      opts.visibilityTarget ??
      (typeof document !== 'undefined' ? (document as unknown as VisibilityTargetLike) : null)

    this.keyTarget = keyTarget
    this.visibilityTarget = visibilityTarget

    keyTarget.addEventListener('keydown', this.onKeyDown)
    keyTarget.addEventListener('keyup', this.onKeyUp)
    keyTarget.addEventListener('blur', this.onBlur)
    if (visibilityTarget) {
      visibilityTarget.addEventListener('visibilitychange', this.onVisibilityChange)
    }
  }

  /** Removes every listener attached by `attach`. Idempotent. */
  detach(): void {
    if (this.keyTarget) {
      this.keyTarget.removeEventListener('keydown', this.onKeyDown)
      this.keyTarget.removeEventListener('keyup', this.onKeyUp)
      this.keyTarget.removeEventListener('blur', this.onBlur)
    }
    if (this.visibilityTarget) {
      this.visibilityTarget.removeEventListener('visibilitychange', this.onVisibilityChange)
    }
    this.keyTarget = null
    this.visibilityTarget = null
  }

  // -------------------------------------------------------------------------
  // Handlers (bound class properties so they keep identity across attach/detach)
  // -------------------------------------------------------------------------

  private readonly onKeyDown = (e: Event): void => {
    const ke = e as KeyboardEventLike
    // OS-level auto-repeat fires `keydown` repeatedly while the key is held.
    // `press()` is already idempotent, but skipping repeats here also skips
    // the `preventDefault()` cost on non-game repeats.
    if (ke.repeat) return
    const cmd = this.keymap[ke.code]
    if (cmd === undefined) return
    this.controller.press(cmd)
    ke.preventDefault?.()
  }

  private readonly onKeyUp = (e: Event): void => {
    const ke = e as KeyboardEventLike
    const cmd = this.keymap[ke.code]
    if (cmd === undefined) return
    this.controller.release(cmd)
    ke.preventDefault?.()
  }

  private readonly onBlur = (): void => {
    this.controller.releaseAll()
  }

  private readonly onVisibilityChange = (): void => {
    if (this.visibilityTarget?.visibilityState === 'hidden') {
      this.controller.releaseAll()
    }
  }
}

/**
 * Duck-typed view of `KeyboardEvent` — the only fields the adapter touches.
 * Lets the tests dispatch a plain `Event` with `code` / `repeat` attached
 * without depending on jsdom or the browser `KeyboardEvent` constructor.
 */
type KeyboardEventLike = Event & {
  readonly code: string
  readonly repeat?: boolean
  preventDefault?: () => void
}
