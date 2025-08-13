package handlers

import (
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
