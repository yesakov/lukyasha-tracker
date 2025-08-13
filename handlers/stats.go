package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yesakov/lukyasha-tracker/models"
	"gorm.io/gorm"
)

func GetStats(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var stats []models.GamePlayerStat
		if err := db.Find(&stats).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, stats)
	}
}

func GetStat(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var stat models.GamePlayerStat
		id := c.Param("id")
		if err := db.First(&stat, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Stat not found"})
			return
		}
		c.JSON(http.StatusOK, stat)
	}
}

func CreateStat(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var stat models.GamePlayerStat
		if err := c.ShouldBindJSON(&stat); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := db.Create(&stat).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, stat)
	}
}

func UpdateStat(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var existing models.GamePlayerStat
		if err := db.First(&existing, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Stat not found"})
			return
		}

		var updated models.GamePlayerStat
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

func DeleteStat(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if err := db.Delete(&models.GamePlayerStat{}, id).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusNoContent)
	}
}
