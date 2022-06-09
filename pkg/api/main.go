package api

import (
	"dcard-pretest/pkg/model"
	"dcard-pretest/pkg/store"
	"net/http"
	"regexp"

	logger "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
)

const (
	leaderboardKey = "dcard-leaderboard"
)

func Run() {
	main()
}

func main() {
	logger.WithFields(logger.Fields{}).Info("Start leaderboard service")

	s := store.NewRedis()
	defer s.Close()

	r := gin.Default()
	v1 := r.Group("/api/v1")

	// curl -X GET http://localhost/api/v1/leaderboard
	v1.GET("/redis-leaderboard", func(c *gin.Context) {
		result, err := s.Top10(c, leaderboardKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "fail"})
		}

		c.JSON(http.StatusOK, gin.H{"topPlayers": result})
	})

	v1.POST("/redis-score", func(c *gin.Context) {
		clientID := c.GetHeader("ClientId")

		re := regexp.MustCompile("^[a-z0-9]{4,16}$")
		matched := re.MatchString(clientID)

		if matched == false {
			logger.Infoln("invalid clientID", clientID)
			c.JSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": "Invalid clientID. format: [a-z0-9]{4,16}"})
			return
		}

		var score model.Score

		if err := c.BindJSON(&score); err != nil {
			logger.Infoln("fail to bind score", err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"status": "fail"})
			return
		}

		score.ClientId = clientID
		err := s.Insert(c, leaderboardKey, score)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "fail"})
		}

		c.JSON(http.StatusAccepted, gin.H{
			"status": "ok",
		})
	})

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
