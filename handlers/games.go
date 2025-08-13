package handlers

import (
    "fmt"
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/yesakov/lukyasha-tracker/models"
    "gorm.io/gorm"
)

func GetGames(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var games []models.Game
		if err := db.Find(&games).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, games)
	}
}

func GetGame(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var game models.Game
		id := c.Param("id")
		if err := db.First(&game, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Game not found"})
			return
		}
		c.JSON(http.StatusOK, game)
	}
}

func CreateGame(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var game models.Game
		if err := c.ShouldBindJSON(&game); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := db.Create(&game).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, game)
	}
}

func UpdateGame(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var existing models.Game
		if err := db.First(&existing, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Game not found"})
			return
		}

		var updated models.Game
		if err := c.ShouldBindJSON(&updated); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := db.Model(&existing).Updates(updated).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, existing)
	}
}

func DeleteGame(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        id := c.Param("id")
        if err := db.Delete(&models.Game{}, id).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        c.Status(http.StatusNoContent)
    }
}

// CreateGameForm creates a game from form-encoded data and redirects to its page
func CreateGameForm(db *gorm.DB) gin.HandlerFunc {
    type input struct {
        EventID    uint `form:"event_id"`
        HomeTeamID uint `form:"home_team_id"`
        AwayTeamID uint `form:"away_team_id"`
    }
    return func(c *gin.Context) {
        var in input
        if err := c.ShouldBind(&in); err != nil || in.EventID == 0 || in.HomeTeamID == 0 || in.AwayTeamID == 0 {
            c.String(http.StatusBadRequest, "Invalid game data")
            return
        }
        if in.HomeTeamID == in.AwayTeamID {
            c.String(http.StatusBadRequest, "Teams must be different")
            return
        }

        var home, away models.Team
        if err := db.First(&home, in.HomeTeamID).Error; err != nil {
            c.String(http.StatusBadRequest, "Home team not found")
            return
        }
        if err := db.First(&away, in.AwayTeamID).Error; err != nil {
            c.String(http.StatusBadRequest, "Away team not found")
            return
        }
        if home.EventID != in.EventID || away.EventID != in.EventID {
            c.String(http.StatusBadRequest, "Teams must belong to the event")
            return
        }

        game := models.Game{EventID: in.EventID, HomeTeamID: in.HomeTeamID, AwayTeamID: in.AwayTeamID}
        if err := db.Create(&game).Error; err != nil {
            c.String(http.StatusInternalServerError, "Database error")
            return
        }
        c.Redirect(http.StatusSeeOther, "/games/"+itoa(game.ID))
    }
}

// ShowGame renders a game page with score and goals
func ShowGame(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        id := c.Param("id")
        var game models.Game
        if err := db.First(&game, id).Error; err != nil {
            c.String(http.StatusNotFound, "Game not found")
            return
        }

        var event models.Event
        db.First(&event, game.EventID)

        var home, away models.Team
        db.First(&home, game.HomeTeamID)
        db.First(&away, game.AwayTeamID)

        var homePlayers, awayPlayers []models.Player
        db.Where("team_id = ?", home.ID).Find(&homePlayers)
        db.Where("team_id = ?", away.ID).Find(&awayPlayers)

        // Load stats (goals/assists) ordered by creation
        var stats []models.GamePlayerStat
        db.Where("game_id = ?", game.ID).Order("created_at ASC").Find(&stats)

        // Attach player names
        type StatItem struct {
            ID        uint
            Type      string
            PlayerID  uint
            Player    string
            TeamID    uint
            TeamName  string
        }
        items := make([]StatItem, 0, len(stats))
        for _, s := range stats {
            var p models.Player
            var t models.Team
            db.First(&p, s.PlayerID)
            db.First(&t, s.TeamID)
            items = append(items, StatItem{
                ID:       s.ID,
                Type:     s.Type,
                PlayerID: s.PlayerID,
                Player:   p.Name,
                TeamID:   s.TeamID,
                TeamName: t.Name,
            })
        }

        c.HTML(http.StatusOK, "game_detail.html", gin.H{
            "Title":       "Game",
            "Event":       event,
            "Game":        game,
            "HomeTeam":    home,
            "AwayTeam":    away,
            "HomePlayers": homePlayers,
            "AwayPlayers": awayPlayers,
            "Stats":       items,
        })
    }
}

// AddGoalHTMX creates goal (and optional assist) via HTMX and returns the refreshed goals list
func AddGoalHTMX(db *gorm.DB) gin.HandlerFunc {
    type input struct {
        PlayerID       uint `form:"player_id"`
        AssistPlayerID uint `form:"assist_player_id"`
    }
    return func(c *gin.Context) {
        id := c.Param("id")
        var game models.Game
        if err := db.First(&game, id).Error; err != nil {
            c.String(http.StatusNotFound, "Game not found")
            return
        }
        var in input
        if err := c.ShouldBind(&in); err != nil || in.PlayerID == 0 {
            c.String(http.StatusBadRequest, "Invalid data")
            return
        }

        var scorer models.Player
        if err := db.First(&scorer, in.PlayerID).Error; err != nil {
            c.String(http.StatusBadRequest, "Scorer not found")
            return
        }
        var scorerTeam models.Team
        db.First(&scorerTeam, scorer.TeamID)

        // validate scorer belongs to this game
        if scorer.TeamID != game.HomeTeamID && scorer.TeamID != game.AwayTeamID {
            c.String(http.StatusBadRequest, "Player not in this game")
            return
        }

        // Create goal stat
        goal := models.GamePlayerStat{PlayerID: scorer.ID, GameID: game.ID, TeamID: scorer.TeamID, Type: models.StatTypeGoal}
        if err := db.Create(&goal).Error; err != nil {
            c.String(http.StatusInternalServerError, "DB error")
            return
        }

        // Optional assist
        if in.AssistPlayerID != 0 && in.AssistPlayerID != in.PlayerID {
            var assist models.Player
            if err := db.First(&assist, in.AssistPlayerID).Error; err == nil {
                if assist.TeamID == game.HomeTeamID || assist.TeamID == game.AwayTeamID {
                    _ = db.Create(&models.GamePlayerStat{PlayerID: assist.ID, GameID: game.ID, TeamID: assist.TeamID, Type: models.StatTypeAssist}).Error
                }
            }
        }

        // Update game score
        if scorer.TeamID == game.HomeTeamID {
            db.Model(&game).UpdateColumn("home_team_goals", gorm.Expr("home_team_goals + 1"))
        } else {
            db.Model(&game).UpdateColumn("away_team_goals", gorm.Expr("away_team_goals + 1"))
        }
        db.First(&game, game.ID)

        // Reload items
        var stats []models.GamePlayerStat
        db.Where("game_id = ?", game.ID).Order("created_at ASC").Find(&stats)
        type StatItem struct {
            ID       uint
            Type     string
            Player   string
            TeamName string
        }
        items := make([]StatItem, 0, len(stats))
        for _, s := range stats {
            var p models.Player
            var t models.Team
            db.First(&p, s.PlayerID)
            db.First(&t, s.TeamID)
            items = append(items, StatItem{ID: s.ID, Type: s.Type, Player: p.Name, TeamName: t.Name})
        }

        c.HTML(http.StatusOK, "game_goals_list.html", gin.H{
            "Game":  game,
            "Stats": items,
        })
    }
}

// helper to convert uint to string without importing strconv everywhere
func itoa(u uint) string {
    // simple and safe for IDs
    return fmt.Sprintf("%d", u)
}
