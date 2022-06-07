package api

import (
	"dcard-pretest/pkg/store"
	"net/http"
	"strconv"

	logger "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
)

// TODO: type Score = store.Score
type Score struct {
	id       int64   `gorm:"primary_key;auto_increment;not_null"`
	ClientID string  `json:"client_id" gorm:"primaryKey,index"`
	Score    float64 `json:"score"`
	// created_at, updated_at
	// expired_at
}

func Run() {
	main()
}

func main() {
	db, err := store.NewPostgres()
	if err != nil {
		panic("failed to connect database")
	}

	sqlDB, err := db.DB()
	if err != nil {
		panic("failed to get generic database")
	}
	defer sqlDB.Close()

	logger.WithFields(logger.Fields{}).Info("Start Indexer")

	err = db.AutoMigrate(&Score{})
	if err != nil {
		panic("failed to migration score table")
	}

	r := gin.Default()
	v1 := r.Group("/api/v1")

	// ?limit=n, default = 10
	// curl -X GET http://localhost/api/v1/leaderboard
	v1.GET("/leaderboard", func(c *gin.Context) {
		scores := []Score{}
		limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "10"), 10, 64)
		db.Limit(int(limit)).Order("score desc").Find(&scores)

		c.JSON(http.StatusOK, scores)
	})

	// TODO: create record
	v1.POST("/score", func(c *gin.Context) {
		clientID := c.Request.Header["ClientId"]

		c.JSON(http.StatusAccepted, gin.H{
			"client_id": clientID,
		})
	})

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
