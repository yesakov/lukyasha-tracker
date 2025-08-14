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
		var game models.Game
		if err := db.First(&game, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Game not found"})
			return
		}
		// Delete all stats for this game
		db.Where("game_id = ?", game.ID).Delete(&models.GamePlayerStat{})
		// Delete the game
		if err := db.Delete(&models.Game{}, game.ID).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		// If htmx, trigger events so event page can refresh and show a toast
		if c.GetHeader("HX-Request") == "true" {
			c.Header("HX-Trigger", "{\"game-removed\":true,\"toast\":\"Game deleted\"}")
			c.Status(http.StatusOK)
			return
		}
		c.Status(http.StatusOK)
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

		// Load all teams and their players within the event
		var allTeams []models.Team
		db.Where("event_id = ?", event.ID).Find(&allTeams)
		type TeamGroup struct {
			Team    models.Team
			Players []models.Player
		}
		groups := make([]TeamGroup, 0, len(allTeams))
		for _, t := range allTeams {
			var pls []models.Player
			db.Where("team_id = ?", t.ID).Find(&pls)
			groups = append(groups, TeamGroup{Team: t, Players: pls})
		}

		// Build grouped rows: each goal with its (optional) assist
		type GoalRow struct {
			ID           uint
			Minute       int
			GoalType     string
			Scorer       string
			ScoringTeam  string
			AssistID     *uint
			AssistPlayer string
			AssistTeam   string
		}
		var goals []models.GamePlayerStat
		db.Where("game_id = ? AND type IN ?", game.ID, []string{models.StatTypeGoal, models.StatTypePenalty, models.StatTypeOwnGoal}).
			Order("created_at ASC").Find(&goals)
		rows := make([]GoalRow, 0, len(goals))
		for _, g := range goals {
			var sp models.Player
			var st models.Team
			db.First(&sp, g.PlayerID)
			db.First(&st, g.TeamID)
			// find assist linked to this goal
			var a models.GamePlayerStat
			var assistID *uint
			var assistPlayer, assistTeam string
			if err := db.Where("game_id = ? AND type = ? AND goal_stat_id = ?", game.ID, models.StatTypeAssist, g.ID).First(&a).Error; err == nil {
				assistID = &a.ID
				var ap models.Player
				var at models.Team
				db.First(&ap, a.PlayerID)
				db.First(&at, a.TeamID)
				assistPlayer = ap.Name
				assistTeam = at.Name
			}
			rows = append(rows, GoalRow{
				ID:           g.ID,
				Minute:       g.Minute,
				GoalType:     g.Type,
				Scorer:       sp.Name,
				ScoringTeam:  st.Name,
				AssistID:     assistID,
				AssistPlayer: assistPlayer,
				AssistTeam:   assistTeam,
			})
		}

		c.HTML(http.StatusOK, "game_detail.html", gin.H{
			"Title":     "Game",
			"Event":     event,
			"Game":      game,
			"HomeTeam":  home,
			"AwayTeam":  away,
			"AllTeams":  groups,
			"GoalRows":  rows,
			"ActiveTab": "events",
			"Content":   "content_game_detail",
		})
	}
}

// AddGoalHTMX creates goal (and optional assist) via HTMX and returns the refreshed goals list
func AddGoalHTMX(db *gorm.DB) gin.HandlerFunc {
	type input struct {
		PlayerID       uint   `form:"player_id"`
		AssistPlayerID uint   `form:"assist_player_id"`
		TeamID         uint   `form:"team_id"`
		Minute         int    `form:"minute"`
		GoalType       string `form:"goal_type"`
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

		// Normalize/validate input
		if in.TeamID != game.HomeTeamID && in.TeamID != game.AwayTeamID {
			// default to home if invalid
			in.TeamID = game.HomeTeamID
		}
		if in.Minute < 0 {
			in.Minute = 0
		}
		if in.Minute > 200 {
			in.Minute = 200
		}
		switch in.GoalType {
		case models.StatTypeGoal, models.StatTypePenalty, models.StatTypeOwnGoal:
			// ok
		default:
			in.GoalType = models.StatTypeGoal
		}

		// Ensure scorer exists
		var scorer models.Player
		if err := db.First(&scorer, in.PlayerID).Error; err != nil {
			c.String(http.StatusBadRequest, "Scorer not found")
			return
		}

		// Create goal stat (TeamID is credited team, not necessarily player's registered team)
		goal := models.GamePlayerStat{PlayerID: scorer.ID, GameID: game.ID, TeamID: in.TeamID, Type: in.GoalType, Minute: in.Minute}
		if err := db.Create(&goal).Error; err != nil {
			c.String(http.StatusInternalServerError, "DB error")
			return
		}

		// Optional assist (skip for own goals)
		if in.AssistPlayerID != 0 && in.AssistPlayerID != in.PlayerID && in.GoalType != models.StatTypeOwnGoal {
			var assist models.Player
			if err := db.First(&assist, in.AssistPlayerID).Error; err == nil {
				_ = db.Create(&models.GamePlayerStat{PlayerID: assist.ID, GameID: game.ID, TeamID: in.TeamID, Type: models.StatTypeAssist, Minute: in.Minute, GoalStatID: &goal.ID}).Error
			}
		}

		// Update game score for credited team
		if in.TeamID == game.HomeTeamID {
			db.Model(&game).UpdateColumn("home_team_goals", gorm.Expr("home_team_goals + 1"))
		} else {
			db.Model(&game).UpdateColumn("away_team_goals", gorm.Expr("away_team_goals + 1"))
		}
		db.First(&game, game.ID)

		// Reload grouped rows
		var goals []models.GamePlayerStat
		db.Where("game_id = ? AND type IN ?", game.ID, []string{models.StatTypeGoal, models.StatTypePenalty, models.StatTypeOwnGoal}).
			Order("created_at ASC").Find(&goals)
		type GoalRow struct {
			ID           uint
			Minute       int
			GoalType     string
			Scorer       string
			ScoringTeam  string
			AssistID     *uint
			AssistPlayer string
			AssistTeam   string
		}
		rows := make([]GoalRow, 0, len(goals))
		for _, g := range goals {
			var sp models.Player
			var st models.Team
			db.First(&sp, g.PlayerID)
			db.First(&st, g.TeamID)
			var a models.GamePlayerStat
			var assistID *uint
			var assistPlayer, assistTeam string
			if err := db.Where("game_id = ? AND type = ? AND goal_stat_id = ?", game.ID, models.StatTypeAssist, g.ID).First(&a).Error; err == nil {
				assistID = &a.ID
				var ap models.Player
				var at models.Team
				db.First(&ap, a.PlayerID)
				db.First(&at, a.TeamID)
				assistPlayer = ap.Name
				assistTeam = at.Name
			}
			rows = append(rows, GoalRow{
				ID:           g.ID,
				Minute:       g.Minute,
				GoalType:     g.Type,
				Scorer:       sp.Name,
				ScoringTeam:  st.Name,
				AssistID:     assistID,
				AssistPlayer: assistPlayer,
				AssistTeam:   assistTeam,
			})
		}

		c.HTML(http.StatusOK, "game_goals_list.html", gin.H{
			"Game":     game,
			"GoalRows": rows,
		})
	}
}

// helper to convert uint to string without importing strconv everywhere
func itoa(u uint) string {
	// simple and safe for IDs
	return fmt.Sprintf("%d", u)
}
