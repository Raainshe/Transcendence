package model

type Achievements struct {
	AvatarChange 	bool `json:"avatar_change"`
	HighestScore2K 	bool `json:"highest_score_2_k"`
	HighestScore5K 	bool `json:"highest_score_5_k"`
	HighestScore10K bool `json:"highest_score_10_k"`
	TotalPoints50K 	bool `json:"total_points_50_k"`
	TotalPoints100K bool `json:"total_points_100_k"`
	Level2        	bool `json:"level_2"`
	Level5        	bool `json:"level_5"`
	Level10        	bool `json:"level_10"`
	Streak2			bool `json:"streak_2"`
	Streak5			bool `json:"streak_5"`
	FirstMpGame		bool `json:"first_mp_game"`
	Played10		bool `json:"played_10"`
	Played50		bool `json:"played_50"`
	Played100		bool `json:"played_100"`
	FirstFriend		bool `json:"first_friend"`
	FirstYear		bool `json:"first_year"`
	FirstClear		bool `json:"first_clear"`
	FirstTetris		bool `json:"first_tetris"`
}