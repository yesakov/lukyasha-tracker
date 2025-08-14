package models

import "gorm.io/gorm"

type Event struct {
    gorm.Model
    Name     string `form:"name" json:"name" gorm:"not null"`
    Date     string `form:"date" json:"date" gorm:"not null"`
    EventURL string `form:"event_url" json:"event_url" gorm:"not null"`
}

type Team struct {
    gorm.Model
    Name    string   `form:"name" json:"name" gorm:"not null;index:idx_team_event_name,unique"`
    EventID uint     `form:"event_id" json:"event_id" gorm:"not null;index:idx_team_event_name,unique"`
    Players []Player `gorm:"constraint:OnDelete:CASCADE;"`
}

type Player struct {
    gorm.Model
    Name   string `form:"name" json:"name" gorm:"not null;index:idx_player_team_name,unique"`
    TeamID uint   `form:"team_id" json:"team_id" gorm:"not null;index:idx_player_team_name,unique"`
}

type Game struct {
    gorm.Model
    EventID       uint `form:"event_id" json:"event_id" gorm:"not null;index"`
    HomeTeamID    uint `form:"home_team_id" json:"home_team_id" gorm:"not null;index"`
    AwayTeamID    uint `form:"away_team_id" json:"away_team_id" gorm:"not null;index"`
    HomeTeamGoals int  `form:"home_team_goals" json:"home_team_goals"`
    AwayTeamGoals int  `form:"away_team_goals" json:"away_team_goals"`
}

type GamePlayerStat struct {
    gorm.Model
    PlayerID uint   `form:"player_id" json:"player_id" gorm:"not null;index"`
    GameID   uint   `form:"game_id" json:"game_id" gorm:"not null;index"`
    TeamID   uint   `form:"team_id" json:"team_id" gorm:"not null;index"`
    Type     string `form:"type" json:"type" gorm:"not null;index"` // "goal", "penalty", "own_goal", or "assist"
    Minute   int    `form:"minute" json:"minute" gorm:"index"`
    // For assists, reference the goal stat they belong to
    GoalStatID *uint `form:"goal_stat_id" json:"goal_stat_id" gorm:"index"`
}

// Optional: constants for Type field
const (
    StatTypeGoal   = "goal"
    StatTypeAssist = "assist"
    StatTypePenalty = "penalty"
    StatTypeOwnGoal = "own_goal"
)
