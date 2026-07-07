import type { Achievements } from '@/types/api'

export type Tier = 'bronze' | 'silver' | 'gold' | 'purple'

export interface BadgeDefinition {
  key: keyof Achievements
  badgeName: string
  tier: Tier
  catImage: string
}

export const tierColors: Record<Tier, string> = {
  bronze: '#cd7f32',
  silver: '#a8a9ad',
  gold:   '#f7d308',
  purple: '#29bb55',
}

export const badgeDefinitions: BadgeDefinition[] = [
  {
    key: 'avatar_change', badgeName: 'Fashionista', tier: 'purple', catImage: 'cat3.png',
  },
  {
    key: 'first_clear', badgeName: 'First Brick', tier: 'purple', catImage: 'cat1.png',
  },
  {
    key: 'first_mp_game', badgeName: 'Challenger', tier: 'purple', catImage: 'cat9.png',
  },
  {
    key: 'first_win', badgeName: 'Victor', tier: 'purple', catImage: 'cat2.png',
  },
  {
    key: 'hundreth_win', badgeName: 'Champion', tier: 'purple', catImage: 'cat8.png',
  },
  {
    key: 'first_friend', badgeName: 'Friendly Neighbor', tier: 'purple', catImage: 'cat10-2.png',
  },
  {
    key: 'highest_score_2_k', badgeName: 'Apprentice Builder', tier: 'bronze', catImage: 'cat7.png',
  },
  {
    key: 'highest_score_10_k', badgeName: 'Builder', tier: 'silver', catImage: 'cat7.png',
  },
  {
    key: 'highest_score_50_k',  badgeName: 'Master Builder', tier: 'gold', catImage: 'cat7.png',
  },
  {
    key: 'total_points_30_k', badgeName: 'Block Collector', tier: 'bronze', catImage: 'cat4.png',
  },
  {
    key: 'total_points_100_k', badgeName: 'Block Hoarder', tier: 'silver', catImage: 'cat4.png',
  },
  {
    key: 'total_points_250_k', badgeName: 'Block Titan', tier: 'gold', catImage: 'cat4.png',
  },
  {
    key: 'level_2', badgeName: 'Rising Planner', tier: 'bronze', catImage: 'cat5.png',
  },
  {
    key: 'level_10', badgeName: 'Tower Architect', tier: 'silver', catImage: 'cat5.png',
  },
  {
    key: 'level_50', badgeName: 'Skyline Master', tier: 'gold', catImage: 'cat5.png',
  },
  {
    key: 'played_10', badgeName: 'Regular', tier: 'bronze', catImage: 'cat6.png',
  },
  {
    key: 'played_50', badgeName: 'Veteran', tier: 'silver', catImage: 'cat6.png',
  },
  {
    key: 'played_100', badgeName: 'Workaholic', tier: 'gold', catImage: 'cat6.png',
  },
  ]