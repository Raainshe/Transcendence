import { describe, expect, it } from 'vitest'

import { Engine } from '@/game/engine/Engine'
import { getOccupiedCells, spawn } from '@/game/engine/Tetrimino'
import { EnginePhase, MinoType } from '@/game/types'

/** Arm first piece: generation delay + spawn. */
function tickFirstPiece(engine: Engine, dt = 300): void {
  engine.update(dt)
  expect(engine.state.currentPiece).not.toBeNull()
}

describe('Engine hold', () => {
  it('allows only one hold per lock cycle', () => {
    const engine = new Engine({ seed: 42 })
    tickFirstPiece(engine)
    expect(engine.hold()).toBe(true)
    expect(engine.hold()).toBe(false)
    expect(engine.state.canHold).toBe(false)
  })

  it('clears hold lock after the piece locks', () => {
    const engine = new Engine({ seed: 42 })
    tickFirstPiece(engine)
    expect(engine.hold()).toBe(true)
    expect(engine.state.canHold).toBe(false)
    engine.hardDrop()
    expect(engine.state.phase).toBe(EnginePhase.Generation)
    engine.update(300)
    expect(engine.state.currentPiece).not.toBeNull()
    expect(engine.state.canHold).toBe(true)
  })

  it('emits piece-held on successful hold', () => {
    const engine = new Engine({ seed: 42 })
    tickFirstPiece(engine)
    engine.drainEvents()
    expect(engine.hold()).toBe(true)
    const events = engine.drainEvents()
    expect(events.some((e) => e.type === 'piece-held')).toBe(true)
  })

  it('stores the swapped-out type in hold when hold was empty', () => {
    const engine = new Engine({ seed: 42 })
    tickFirstPiece(engine)
    const firstType = engine.state.currentPiece!.type
    expect(engine.state.holdPiece).toBeNull()
    expect(engine.hold()).toBe(true)
    expect(engine.state.holdPiece).toBe(firstType)
    expect(engine.state.currentPiece).not.toBeNull()
    expect(engine.state.currentPiece!.type).not.toBe(firstType)
  })

  it('swaps with hold when the slot is occupied', () => {
    const engine = new Engine({ seed: 42 })
    tickFirstPiece(engine)
    expect(engine.hold()).toBe(true)
    const slotBeforeLock = engine.state.holdPiece
    expect(slotBeforeLock).not.toBeNull()
    engine.hardDrop()
    engine.update(300)
    const pieceBeforeHold = engine.state.currentPiece!.type
    expect(engine.hold()).toBe(true)
    expect(engine.state.holdPiece).toBe(pieceBeforeHold)
    expect(engine.state.currentPiece!.type).toBe(slotBeforeLock)
  })

  it('block out on hold when the bag spawn overlaps locked cells', () => {
    let engine: Engine | null = null
    for (let seed = 0; seed < 2000; seed++) {
      const e = new Engine({ seed })
      tickFirstPiece(e)
      const curCells = getOccupiedCells(e.state.currentPiece!)
      const taken = new Set(curCells.map((c) => `${c.x},${c.y}`))
      const nextT = e.state.nextQueue[0]
      if (nextT === undefined) continue
      const spawnCells = getOccupiedCells(spawn(nextT))
      if (spawnCells.some((c) => taken.has(`${c.x},${c.y}`))) continue
      engine = e
      for (const c of spawnCells) {
        e.matrixRef.set(c.x, c.y, MinoType.I)
      }
      break
    }
    expect(engine).not.toBeNull()
    expect(engine!.hold()).toBe(false)
    expect(engine!.state.gameOver).toBe(true)
    expect(engine!.state.gameOverReason).toBe('blockOut')
  })
})
