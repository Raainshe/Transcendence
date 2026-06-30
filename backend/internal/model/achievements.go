package model

type Achievements struct {
	AvatarChange    bool `json:"avatar_change"`      //OnAvatarUploaded
	HighestScore2K  bool `json:"highest_score_2_k"`  //OnGameEnd - EndMatch(match_end.go) for MP & RecordMatch (game.go) for SP
	HighestScore10K bool `json:"highest_score_5_k"`  //OnGameEnd - EndMatch(match_end.go) for MP & RecordMatch (game.go) for SP
	HighestScore50K bool `json:"highest_score_10_k"` //OnGameEnd - EndMatch(match_end.go) for MP & RecordMatch (game.go) for SP
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

type Badge struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Unlocked    bool   `json:"unlocked"`
}

// need to think of nice badge names
var badgeDefinitions = []struct {
	Key         string
	Name        string
	Description string
}{
	{Key: "avatar_change", Name: "", Description: "Upload a profile picture"},
	{Key: "first_clear", Name: "", Description: "Clear your first line"},
	//{Key: "first_tetris", Name: "", Description: "Clear four lines at once"},
	{Key: "highest_score_2_k", Name: "", Description: "Score 2,000 points in a single game"},
	{Key: "highest_score_10_k", Name: "", Description: "Score 10,000 points in a single game"},
	{Key: "highest_score_50_k", Name: "", Description: "Score 50,000 points in a single game"},
	{Key: "total_points_50_k", Name: "", Description: "Earn 50,000 total points"},
	{Key: "total_points_100_k", Name: "", Description: "Earn 100,000 total points"},
	{Key: "total_points_250_k", Name: "", Description: "Earn 250,000 total points"},
	{Key: "level_2", Name: "", Description: "Reach level 2"},
	{Key: "level_10", Name: "", Description: "Reach level 10"},
	{Key: "level_50", Name: "", Description: "Reach level 50"},
	//{Key: "streak_2", Name: "", Description: "Win 2 matches in a row"},
	//{Key: "streak_5", Name: "", Description: "Win 5 matches in a row"},
	{Key: "first_mp_game", Name: "", Description: "Play your first multiplayer game"},
	{Key: "played_10", Name: "", Description: "Play 10 games"},
	{Key: "played_50", Name: "", Description: "Play 50 games"},
	{Key: "played_100", Name: "", Description: "Play 100 games"},
	{Key: "first_friend", Name: "", Description: "Add your first friend"},
	//{Key: "first_year", Name: "", Description: "Be a member for one year"},
}

func BadgesFromAchievements(a Achievements) []Badge {
	unlocked := map[string]bool{
		"avatar_change": a.AvatarChange,
		"first_clear":   a.FirstClear,
		//"first_tetris":       a.FirstTetris,
		"highest_score_2_k":  a.HighestScore2K,
		"highest_score_10_k": a.HighestScore10K,
		"highest_score_50_k": a.HighestScore50K,
		"total_points_30_k":  a.TotalPoints30K,
		"total_points_100_k": a.TotalPoints100K,
		"total_points_250_k": a.TotalPoints250K,
		"level_2":            a.Level2,
		"level_10":           a.Level10,
		"level_50":           a.Level50,
		//"streak_2":           a.Streak2,
		//"streak_5":           a.Streak5,
		"first_mp_game": a.FirstMpGame,
		"played_10":     a.Played10,
		"played_50":     a.Played50,
		"played_100":    a.Played100,
		"first_friend":  a.FirstFriend,
		//"first_year":         a.FirstYear,
	}

	badges := make([]Badge, 0, len(badgeDefinitions))
	for _, def := range badgeDefinitions {
		badges = append(badges, Badge{
			Key:         def.Key,
			Name:        def.Name,
			Description: def.Description,
			Unlocked:    unlocked[def.Key],
		})
	}
	return badges
}

//ideas:
// give game struct to on gameend for mode && time (eg. nightowl, streak)
// + add more mode specific achievements
// AVG_Score in GetUserStats -> some kind of elo calculation
