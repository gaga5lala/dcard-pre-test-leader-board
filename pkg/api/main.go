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
	logger.Info("Start leaderboard service")

	s := store.NewRedis()
	defer s.Close()

	r := setupRouter(s)
	r.Run(":80")
}

func setupRouter(s *store.Store) *gin.Engine {
	r := gin.Default()
	v1 := r.Group("/api/v1").Use(JSONMiddleware())

	v1.GET("/leaderboard", GetLeaderboardHandler(s))
	v1.POST("/score", PostScoreHandler(s))
	return r
}

func GetLeaderboardHandler(s *store.Store) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		result, err := s.Top10(c, leaderboardKey)
		if err != nil {
			logger.Infoln("fail to get top10", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"status": "fail"})
		}

		c.JSON(http.StatusOK, gin.H{"topPlayers": result})
	}

	return gin.HandlerFunc(fn)
}

func PostScoreHandler(s *store.Store) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		clientID := c.GetHeader("ClientId")

		// arbitrary constrain by myself
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
			logger.Infoln("fail to insert score", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"status": "fail"})
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	}
	return gin.HandlerFunc(fn)
}

func JSONMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "application/json")
		c.Next()
	}
}
