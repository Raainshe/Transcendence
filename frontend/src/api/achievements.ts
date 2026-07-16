import { apiFetch } from '@/api/client'
import type { Achievements, AchievementsResponse,} from '@/types/api'

export function getUserAchievements(userID: string): Promise<Achievements> {
    return apiFetch<AchievementsResponse>(`/users/${userID}/achievements`, {
        method: 'GET',
    }).then((res) => res.achievements)
}
