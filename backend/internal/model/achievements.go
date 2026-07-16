package model

type Achievements struct {
	AvatarChange    bool `json:"avatar_change"`
	HighestScore2K  bool `json:"highest_score_2_k"`
	HighestScore10K bool `json:"highest_score_10_k"`
	HighestScore50K bool `json:"highest_score_50_k"`
	TotalPoints30K  bool `json:"total_points_30_k"`
	TotalPoints100K bool `json:"total_points_100_k"`
	TotalPoints250K bool `json:"total_points_250_k"`
	Level2          bool `json:"level_2"`
	Level8          bool `json:"level_8"`
	Level15         bool `json:"level_15"`
	FirstMpGame     bool `json:"first_mp_game"`
	FirstWin        bool `json:"first_win"`
	HundrethWin     bool `json:"hundreth_win"`
	Played10        bool `json:"played_10"`
	Played50        bool `json:"played_50"`
	Played100       bool `json:"played_100"`
	FirstFriend     bool `json:"first_friend"`
	FirstClear      bool `json:"first_clear"`
}
