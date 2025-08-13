package main

import (
	// "html/template"
	"net/http"

	"github.com/gin-gonic/gin"
	moderncSqlite "gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/yesakov/lukyasha-tracker/handlers"
	"github.com/yesakov/lukyasha-tracker/models"

	_ "modernc.org/sqlite"
)

var DB *gorm.DB

// func LoadTemplates() *template.Template {
// 	return template.Must(template.ParseFiles(
// 		"templates/layout.html",
// 		"templates/home.html",
// 		"templates/events.html",
// 		"templates/events_new.html",
// 		"templates/event_detail.html",
// 	))

// }

func InitDB() {
	var err error
	DB, err = gorm.Open(moderncSqlite.New(moderncSqlite.Config{
		DSN:        "data.db",
		DriverName: "sqlite",
	}), &gorm.Config{})

	if err != nil {
		panic("failed to connect database: " + err.Error())
	}

	DB.AutoMigrate(&models.Event{}, &models.Game{}, &models.GamePlayerStat{}, &models.Player{}, &models.Team{})
}

// func SeedDB() {
// 	var count int64
// 	DB.Model(&models.Event{}).Count(&count)
// 	if count == 0 {
// 		DB.Create(&models.Event{Name: "Spring Tournament", Date: "2025-04-15"})
// 		DB.Create(&models.Event{Name: "Summer League", Date: "2025-07-01"})
// 	}
// }

func main() {
	r := gin.Default()

	// Load HTML templates
	r.LoadHTMLGlob("templates/*")

	InitDB()
	// SeedDB()

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "home.html", gin.H{
			"title": "Main website",
		})
	})

	r.GET("/events/new", handlers.NewEventForm())
	r.GET("/events", handlers.ListEvents(DB))
	r.GET("/events/:id", handlers.ShowEvent(DB))
	r.POST("/events", handlers.CreateEvent(DB))

	r.POST("/teams", handlers.CreateTeamHTMX(DB))
	r.POST("/players", handlers.CreatePlayerHTMX(DB))
	r.DELETE("/teams/:id", handlers.DeleteTeam(DB))
	r.DELETE("/players/:id", handlers.DeletePlayer(DB))

	// r.GET("/events/:id", handlers.ShowEvent(DB))

	// r.PUT("/events/:id", handlers.UpdateEvent(DB))
	// r.DELETE("/events/:id", handlers.DeleteEvent(DB))

	// r.GET("/teams", handlers.GetTeams(DB))
	// r.GET("/teams/:id", handlers.GetTeam(DB))
	// r.POST("/teams", handlers.CreateTeamJSON(DB))

	// r.PUT("/teams/:id", handlers.UpdateTeam(DB))
	// r.DELETE("/teams/:id", handlers.DeleteTeam(DB))

	// r.GET("/players", handlers.GetPlayers(DB))
	// r.GET("/players/:id", handlers.GetPlayer(DB))
	// r.POST("/players", handlers.CreatePlayerJSON(DB))
	// r.PUT("/players/:id", handlers.UpdatePlayer(DB))
	// r.DELETE("/players/:id", handlers.DeletePlayer(DB))

	// r.GET("/games", handlers.GetGames(DB))
	// r.GET("/games/:id", handlers.GetGame(DB))
	// r.POST("/games", handlers.CreateGame(DB))
	// r.PUT("/games/:id", handlers.UpdateGame(DB))
	// r.DELETE("/games/:id", handlers.DeleteGame(DB))

	// r.GET("/stats", handlers.GetStats(DB))
	// r.GET("/stats/:id", handlers.GetStat(DB))
	// r.POST("/stats", handlers.CreateStat(DB))
	// r.PUT("/stats/:id", handlers.UpdateStat(DB))
	// r.DELETE("/stats/:id", handlers.DeleteStat(DB))

	r.Run(":8080")
}
