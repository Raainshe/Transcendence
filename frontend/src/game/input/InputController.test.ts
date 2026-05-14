import { describe, expect, it, vi } from 'vitest'

import { InputController, type EngineLike } from '@/game/input/InputController'
import { InputCommand } from '@/game/types'

function stubEngine(overrides: Partial<EngineLike> = {}): EngineLike {
  return {
    moveLeft: vi.fn(() => true),
    moveRight: vi.fn(() => true),
    rotateCW: vi.fn(() => true),
    rotateCCW: vi.fn(() => true),
    softDrop: vi.fn(),
    hardDrop: vi.fn(() => true),
    hold: vi.fn(() => true),
    ...overrides,
  }
}

describe('InputController', () => {
  it('suppresses press and update when isInputBlocked is true', () => {
    const moveLeft = vi.fn(() => true)
    const moveRight = vi.fn(() => true)
    const engine: EngineLike = {
      moveLeft,
      moveRight,
      rotateCW: vi.fn(() => true),
      rotateCCW: vi.fn(() => true),
      softDrop: vi.fn(),
      hardDrop: vi.fn(() => true),
      hold: vi.fn(() => true),
    }
    let blocked = false
    const ctrl = new InputController(engine, { isInputBlocked: () => blocked })

    ctrl.press(InputCommand.MoveLeft)
    expect(moveLeft).toHaveBeenCalledTimes(1)

    blocked = true
    ctrl.press(InputCommand.MoveRight)
    expect(moveRight).not.toHaveBeenCalled()

    ctrl.update(100)
    expect(moveLeft).toHaveBeenCalledTimes(1)

    blocked = false
    ctrl.update(200)
    expect(moveLeft.mock.calls.length).toBeGreaterThan(1)
  })

  it('releaseAll clears soft drop even when press is currently blocked', () => {
    const softDrop = vi.fn()
    const engine = stubEngine({ softDrop })
    let blocked = false
    const ctrl = new InputController(engine, { isInputBlocked: () => blocked })
    ctrl.press(InputCommand.SoftDrop)
    expect(softDrop).toHaveBeenCalledWith(true)
    blocked = true
    ctrl.releaseAll()
    expect(softDrop).toHaveBeenLastCalledWith(false)
  })

  it('fires hold once per key press', () => {
    const hold = vi.fn(() => true)
    const engine = stubEngine({ hold })
    const ctrl = new InputController(engine)
    ctrl.press(InputCommand.Hold)
    ctrl.press(InputCommand.Hold)
    expect(hold).toHaveBeenCalledTimes(1)
    ctrl.release(InputCommand.Hold)
    ctrl.press(InputCommand.Hold)
    expect(hold).toHaveBeenCalledTimes(2)
  })

  it('releaseAll clears hold latch', () => {
    const hold = vi.fn(() => true)
    const engine = stubEngine({ hold })
    const ctrl = new InputController(engine)
    ctrl.press(InputCommand.Hold)
    ctrl.releaseAll()
    ctrl.press(InputCommand.Hold)
    expect(hold).toHaveBeenCalledTimes(2)
  })
})
