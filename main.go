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

func InitDB() {
	var err error
	DB, err = gorm.Open(moderncSqlite.New(moderncSqlite.Config{
		DSN:        "data.db",
		DriverName: "sqlite",
	}), &gorm.Config{})

	if err != nil {
		panic("failed to connect database: " + err.Error())
	}

	// Ensure SQLite enforces foreign keys
	DB.Exec("PRAGMA foreign_keys = ON;")

	DB.AutoMigrate(&models.Event{}, &models.Game{}, &models.GamePlayerStat{}, &models.Player{}, &models.Team{})
}

func main() {
	r := gin.Default()

	// Load HTML templates
	r.LoadHTMLGlob("templates/*")

	// Serve static assets (CSS, JS, images)
	r.Static("/static", "static")

	InitDB()

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "home.html", gin.H{
			"Title":     "Main website",
			"ActiveTab": "home",
			"Content":   "content_home",
		})
	})

	r.GET("/events/new", handlers.NewEventForm())
	r.GET("/events", handlers.ListEvents(DB))
	r.GET("/events/:id", handlers.ShowEvent(DB))
	r.GET("/events/:id/games_partial", handlers.EventGamesPartial(DB))
	r.GET("/events/:id/stats_partial", handlers.EventStatsPartial(DB))
	r.POST("/events", handlers.CreateEvent(DB))
	r.GET("/events/:id/team_options", handlers.TeamOptions(DB))
	r.DELETE("/events/:id", handlers.DeleteEvent(DB))

	r.POST("/teams", handlers.CreateTeamHTMX(DB))
	r.POST("/players", handlers.CreatePlayerHTMX(DB))
	r.DELETE("/teams/:id", handlers.DeleteTeam(DB))
	r.DELETE("/players/:id", handlers.DeletePlayer(DB))

	// Games and scoring
	r.POST("/games", handlers.CreateGameForm(DB))
	r.GET("/games/:id", handlers.ShowGame(DB))
	r.DELETE("/games/:id", handlers.DeleteGame(DB))
	r.POST("/games/:id/goals", handlers.AddGoalHTMX(DB))
	r.DELETE("/stats/:id", handlers.DeleteStat(DB))

	r.Run(":8080")
}
