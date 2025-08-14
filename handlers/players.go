package handlers

import (
	"net/http"

	"github.com/yesakov/lukyasha-tracker/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func CreatePlayerHTMX(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var player models.Player
		if err := c.ShouldBind(&player); err != nil {
			c.String(http.StatusBadRequest, "Invalid data")
			return
		}

		if player.Name == "" || player.TeamID == 0 {
			c.String(http.StatusBadRequest, "Name and TeamID required")
			return
		}

		// Check for duplicate
		var existing models.Player
		if err := db.Where("team_id = ? AND name = ?", player.TeamID, player.Name).First(&existing).Error; err == nil {
			c.String(http.StatusConflict, "Player already exists")
			return
		}

		db.Create(&player)

		// This is all you need:
		c.HTML(http.StatusOK, "player_item.html", player)
	}
}

func GetPlayers(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var players []models.Player
		if err := db.Find(&players).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, players)
	}
}

func GetPlayer(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var player models.Player
		id := c.Param("id")
		if err := db.First(&player, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Player not found"})
			return
		}
		c.JSON(http.StatusOK, player)
	}
}

func CreatePlayerJSON(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var player models.Player
		if err := c.ShouldBindJSON(&player); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		if player.Name == "" || player.TeamID == 0 {
			c.JSON(400, gin.H{"error": "Name and TeamID are required"})
			return
		}

		// Check for existing player in team
		var existing models.Player
		if err := db.
			Where("team_id = ? AND name = ?", player.TeamID, player.Name).
			First(&existing).Error; err == nil {
			c.JSON(409, gin.H{"error": "Player with this name already exists in the team"})
			return
		}

		if err := db.Create(&player).Error; err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		db.First(&player, player.ID)
		c.JSON(201, player)
	}
}

func UpdatePlayer(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var existing models.Player
		if err := db.First(&existing, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Player not found"})
			return
		}

		var updated models.Player
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

func DeletePlayer(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if err := db.Delete(&models.Player{}, id).Error; err != nil {
			c.String(http.StatusInternalServerError, "Delete error")
			return
		}
		c.Status(http.StatusOK) // HTMX will remove the target from DOM
	}
}
