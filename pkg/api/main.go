package api

import (
	"context"
	"dcard-pretest/pkg/store"
	"net/http"
	"regexp"

	"github.com/go-redis/redis/v9"
	logger "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
)

const (
	leaderboardKey = "dcard-leaderboard"
)

var (
	start = int64(0)
	stop  = int64(9)
)

// TODO: type Score = store.Score
type Score struct {
	ClientID string  `json:"clientId"`
	Score    float64 `json:"score"`
}

func Run() {
	main()
}

func main() {
	logger.WithFields(logger.Fields{}).Info("Start leaderboard service")

	redisClient := store.NewRedis()
	defer redisClient.Close()

	r := gin.Default()
	v1 := r.Group("/api/v1")

	// curl -X GET http://localhost/api/v1/leaderboard
	v1.GET("/redis-leaderboard", func(c *gin.Context) {
		scores, err := redisClient.ZRevRangeWithScores(c, leaderboardKey, start, stop).Result()
		if err != nil {
			logger.Errorln(err)
			c.JSON(http.StatusInternalServerError, gin.H{"status": "fail"})
			return
		}

		result := make([]*Score, len(scores))

		for i, v := range scores {
			result[i] = &Score{
				ClientID: v.Member.(string),
				Score:    v.Score,
			}
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

		var score Score

		if err := c.BindJSON(&score); err != nil {
			logger.Infoln("fail to bind score", err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"status": "fail"})
			return
		}
		score.ClientID = clientID

		err := addScore(c, redisClient, score)
		if err != nil {
			logger.Infoln("fail to add score", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"status": "fail"})
		}
		c.JSON(http.StatusAccepted, gin.H{
			"status": "ok",
		})
	})

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

func addScore(ctx context.Context, c *redis.Client, p Score) error {
	clientId := p.ClientID
	score := p.Score

	err := c.ZAdd(ctx, leaderboardKey, redis.Z{
		Score:  score,
		Member: clientId,
	}).Err()

	if err != nil {
		return err
	}

	return nil
}
