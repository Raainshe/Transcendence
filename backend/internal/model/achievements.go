package model

type Achievements struct {
	AvatarChange    bool `json:"avatar_change"`      //OnAvatarUploaded
	HighestScore2K  bool `json:"highest_score_2_k"`  //OnGameEnd - EndMatch(match_end.go) for MP & RecordMatch (game.go) for SP
	HighestScore10K bool `json:"highest_score_10_k"` //OnGameEnd - EndMatch(match_end.go) for MP & RecordMatch (game.go) for SP
	HighestScore50K bool `json:"highest_score_50_k"` //OnGameEnd - EndMatch(match_end.go) for MP & RecordMatch (game.go) for SP
	TotalPoints30K  bool `json:"total_points_50_k"`  //OnGameEnd - EndMatch(match_end.go) for MP & RecordMatch (game.go) for SP
	TotalPoints100K bool `json:"total_points_100_k"` //OnGameEnd - EndMatch(match_end.go) for MP & RecordMatch (game.go) for SP
	TotalPoints250K bool `json:"total_points_250_k"`
	Level2          bool `json:"level_2"`  //OnGameEnd - EndMatch(match_end.go) for MP & RecordMatch (game.go) for SP
	Level10         bool `json:"level_10"` //OnGameEnd - EndMatch(match_end.go) for MP & RecordMatch (game.go) for SP
	Level50         bool `json:"level_50"` //OnGameEnd - EndMatch(match_end.go) for MP & RecordMatch (game.go) for SP
	// Streak2			bool `json:"streak_2"`				//either games played or login
	// Streak5			bool `json:"streak_5"`				//either games played or login
	FirstMpGame bool `json:"first_mp_game"` //OnGameEnd - // ??
	FirstWin    bool `json:"first_win"`     //OnGameEnd - GetUserStats /MP&SP
	HundrethWin bool `json:"hundreth_win"`  //OnGameEnd - GetUserStats /MP&SP
	Played10    bool `json:"played_10"`     //OnGameEnd - GetUserStats /MP&SP
	Played50    bool `json:"played_50"`     //OnGameEnd - GetUserStats /MP&SP
	Played100   bool `json:"played_100"`    //OnGameEnd - GetUserStats /MP&SP
	FirstFriend bool `json:"first_friend"`  //OnFriendAdded
	// FirstYear		bool `json:"first_year"`			//needACCCreatedInfo
	FirstClear bool `json:"first_clear"` //OnGameEnd - EndMatch(match_end.go) for MP & RecordMatch (game.go) for SP
	// FirstTetris		bool `json:"first_tetris"`			//noData
}

//ideas:
// give game struct to on gameend for mode && time (eg. nightowl, streak)
// + add more mode specific achievements
// AVG_Score in GetUserStats -> some kind of elo calculation
