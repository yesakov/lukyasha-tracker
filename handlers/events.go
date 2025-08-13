package handlers

import (
	"fmt"
	"net/http"

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

		c.HTML(http.StatusOK, "event_detail.html", gin.H{
			"Title": "Event Details",
			"Event": event,
			"Teams": teams,
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
