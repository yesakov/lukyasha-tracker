package handlers

import (
    "fmt"
    "net/http"
    "sort"

    "github.com/yesakov/lukyasha-tracker/models"

    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
)

func NewEventForm() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "events_new.html", gin.H{
			"Title": "Create New Event",
		})
	}
}

func ListEvents(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var events []models.Event
		db.Find(&events)
		c.HTML(http.StatusOK, "events.html", gin.H{
			"Title":  "Events",
			"Events": events,
		})
	}
}

func ShowEvent(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        id := c.Param("id")

        var event models.Event
        if err := db.First(&event, id).Error; err != nil {
            c.String(http.StatusNotFound, "Event not found")
            return
        }

        var teams []models.Team
        db.Where("event_id = ?", event.ID).Find(&teams)

		// Load players for each team
		for i := range teams {
			var players []models.Player
			db.Where("team_id = ?", teams[i].ID).Find(&players)
			teams[i].Players = players
		}

        // Load games for this event
        var games []models.Game
        db.Where("event_id = ?", event.ID).Find(&games)

        // Build standings
        type StandRow struct {
            Team      models.Team
            Played    int
            Wins      int
            Draws     int
            Losses    int
            GF        int
            GA        int
            GD        int
            Points    int
        }

        standMap := make(map[uint]*StandRow)
        for _, t := range teams {
            standMap[t.ID] = &StandRow{Team: t}
        }
        for _, g := range games {
            home := standMap[g.HomeTeamID]
            away := standMap[g.AwayTeamID]
            if home == nil || away == nil { continue }
            home.Played++; away.Played++
            home.GF += g.HomeTeamGoals; home.GA += g.AwayTeamGoals
            away.GF += g.AwayTeamGoals; away.GA += g.HomeTeamGoals
            if g.HomeTeamGoals > g.AwayTeamGoals {
                home.Wins++; home.Points += 3; away.Losses++
            } else if g.HomeTeamGoals < g.AwayTeamGoals {
                away.Wins++; away.Points += 3; home.Losses++
            } else {
                home.Draws++; away.Draws++; home.Points++; away.Points++
            }
        }
        for _, s := range standMap { s.GD = s.GF - s.GA }
        standings := make([]*StandRow, 0, len(standMap))
        for _, s := range standMap { standings = append(standings, s) }
        sort.Slice(standings, func(i, j int) bool {
            if standings[i].Points != standings[j].Points { return standings[i].Points > standings[j].Points }
            if standings[i].GD != standings[j].GD { return standings[i].GD > standings[j].GD }
            if standings[i].GF != standings[j].GF { return standings[i].GF > standings[j].GF }
            return standings[i].Team.Name < standings[j].Team.Name
        })

        // Leaderboards (top scorers/assistants)
        type aggRow struct { PlayerID uint; Cnt int }
        var gameIDs []uint
        db.Model(&models.Game{}).Where("event_id = ?", event.ID).Pluck("id", &gameIDs)
        topScorers := []gin.H{}
        topAssists := []gin.H{}
        if len(gameIDs) > 0 {
            var gr []aggRow
            db.Model(&models.GamePlayerStat{}).
                Select("player_id, COUNT(*) as cnt").
                Where("type = ? AND game_id IN ?", models.StatTypeGoal, gameIDs).
                Group("player_id").Order("cnt DESC").Limit(10).Scan(&gr)
            for _, r := range gr {
                var p models.Player; var t models.Team
                db.First(&p, r.PlayerID); db.First(&t, p.TeamID)
                topScorers = append(topScorers, gin.H{"player": p.Name, "team": t.Name, "count": r.Cnt})
            }
            gr = nil
            db.Model(&models.GamePlayerStat{}).
                Select("player_id, COUNT(*) as cnt").
                Where("type = ? AND game_id IN ?", models.StatTypeAssist, gameIDs).
                Group("player_id").Order("cnt DESC").Limit(10).Scan(&gr)
            for _, r := range gr {
                var p models.Player; var t models.Team
                db.First(&p, r.PlayerID); db.First(&t, p.TeamID)
                topAssists = append(topAssists, gin.H{"player": p.Name, "team": t.Name, "count": r.Cnt})
            }
        }

        c.HTML(http.StatusOK, "event_detail.html", gin.H{
            "Title": "Event Details",
            "Event": event,
            "Teams": teams,
            "Games": games,
            "Standings": standings,
            "TopScorers": topScorers,
            "TopAssists": topAssists,
        })
    }
}

func CreateEvent(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input models.Event
		if err := c.ShouldBind(&input); err != nil {
			c.HTML(http.StatusOK, "events_new_form.html", gin.H{
				"Title": "Create New Event",
				"Error": "Invalid form data",
			})
			return
		}

		if input.Name == "" || input.Date == "" || input.EventURL == "" {
			fmt.Printf("--------------- %+v", input)
			c.HTML(http.StatusOK, "events_new_form.html", gin.H{
				"Title":    "Create New Event",
				"Error":    "All fields are required",
				"Name":     input.Name,
				"Date":     input.Date,
				"EventURL": input.EventURL,
			})
			return
		}

		if err := db.Create(&input).Error; err != nil {
			c.HTML(http.StatusOK, "events_new_form.html", gin.H{
				"Title": "Create New Event",
				"Error": "Database error",
			})
			return
		}

		c.Redirect(http.StatusSeeOther, fmt.Sprintf("/events/%d", input.ID))
	}
}

func GetEvents(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var events []models.Event
		if err := db.Find(&events).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, events)
	}
}

func GetEvent(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var event models.Event
		id := c.Param("id")
		if err := db.First(&event, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
			return
		}
		c.JSON(http.StatusOK, event)
	}
}

func UpdateEvent(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var existing models.Event
		if err := db.First(&existing, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
			return
		}

		var updated models.Event
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

func DeleteEvent(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if err := db.Delete(&models.Event{}, id).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusNoContent)
	}
}
