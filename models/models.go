package models

import "gorm.io/gorm"

type Event struct {
	gorm.Model
	Name     string `form:"name" json:"name"`
	Date     string `form:"date" json:"date"`
	EventURL string `form:"event_url" json:"event_url"`
}

type Team struct {
	gorm.Model
	Name    string `form:"name" json:"name"`
	EventID uint   `form:"event_id" json:"event_id"`
	Players []Player
}

type Player struct {
	gorm.Model
	Name   string `form:"name" json:"name"`
	TeamID uint   `form:"team_id" json:"team_id"`
}

type Game struct {
	gorm.Model
	EventID       uint `form:"event_id" json:"event_id"`
	HomeTeamID    uint `form:"home_team_id" json:"home_team_id"`
	AwayTeamID    uint `form:"away_team_id" json:"away_team_id"`
	HomeTeamGoals int  `form:"home_team_goals" json:"home_team_goals"`
	AwayTeamGoals int  `form:"away_team_goals" json:"away_team_goals"`
}

type GamePlayerStat struct {
	gorm.Model
	PlayerID uint   `form:"player_id" json:"player_id"`
	GameID   uint   `form:"game_id" json:"game_id"`
	TeamID   uint   `form:"team_id" json:"team_id"`
	Type     string `form:"type" json:"type"` // "goal" or "assist"
}

// Optional: constants for Type field
const (
	StatTypeGoal   = "goal"
	StatTypeAssist = "assist"
)
