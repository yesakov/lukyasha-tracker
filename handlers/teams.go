package handlers

import (
	"net/http"

	"github.com/yesakov/lukyasha-tracker/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func CreateTeamHTMX(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var team models.Team
		if err := c.ShouldBind(&team); err != nil {
			c.String(http.StatusBadRequest, "Invalid data")
			return
		}

		if team.Name == "" || team.EventID == 0 {
			c.String(http.StatusBadRequest, "Name and EventID required")
			return
		}

		// Check for existing
		var existing models.Team
		if err := db.Where("event_id = ? AND name = ?", team.EventID, team.Name).First(&existing).Error; err == nil {
			c.String(http.StatusConflict, "Team already exists")
			return
		}

		db.Create(&team)
		team.Players = []models.Player{} // empty

		// This is the magic:
		c.HTML(http.StatusOK, "team_card.html", team)
	}
}

func DeleteTeam(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		// Delete players first (foreign key)
		db.Where("team_id = ?", id).Delete(&models.Player{})

		// Delete team
		if err := db.Delete(&models.Team{}, id).Error; err != nil {
			c.String(http.StatusInternalServerError, "Delete error")
			return
		}

		c.Status(http.StatusNoContent) // HTMX will remove the target from DOM
	}
}

func CreateTeamJSON(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var team models.Team
		if err := c.ShouldBindJSON(&team); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		if team.Name == "" || team.EventID == 0 {
			c.JSON(400, gin.H{"error": "Name and EventID are required"})
			return
		}

		// Check for existing team in same event
		var existing models.Team
		if err := db.
			Where("event_id = ? AND name = ?", team.EventID, team.Name).
			First(&existing).Error; err == nil {
			// Found duplicate
			c.JSON(409, gin.H{"error": "Team with this name already exists in the event"})
			return
		}

		// Create new team
		if err := db.Create(&team).Error; err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		db.First(&team, team.ID)
		c.JSON(201, team)
	}
}

func GetTeams(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var teams []models.Team
		if err := db.Find(&teams).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, teams)
	}
}

func GetTeam(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var team models.Team
		id := c.Param("id")
		if err := db.First(&team, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
			return
		}
		c.JSON(http.StatusOK, team)
	}
}

func CreateTeam(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var team models.Team
		if err := c.ShouldBindJSON(&team); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := db.Create(&team).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, team)
	}
}

func UpdateTeam(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var existing models.Team
		if err := db.First(&existing, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
			return
		}

		var updated models.Team
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
