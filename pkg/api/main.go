package api

import (
	"context"
	"dcard-pretest/pkg/store"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-redis/redis/v9"
	logger "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
)

const (
	leaderboardKey = "dcard-leaderboard"
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

	v1.GET("/redis-leaderboard", func(c *gin.Context) {
		redisClient := redis.NewClient(&redis.Options{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		})
		scores, err := redisClient.ZRevRangeWithScores(c, leaderboardKey, 0, 9).Result()
		if err != nil {
			logger.Errorln(err)
		}
		c.JSON(http.StatusOK, scores)
	})

	v1.POST("/redis-score", func(c *gin.Context) {
		redisClient := redis.NewClient(&redis.Options{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		})
		params := map[string]interface{}{}
		clientID := c.GetHeader("ClientId")
		params["client_id"] = clientID
		err = json.NewDecoder(c.Request.Body).Decode(&params)

		_, err = addScore(redisClient, params)
		if err != nil {
			logger.Infoln(err)
		}
		c.JSON(http.StatusAccepted, gin.H{
			"resp": "hihi",
		})
	})

	// TODO: create record
	v1.POST("/score", func(c *gin.Context) {
		clientID := c.GetHeader("ClientId")

		c.JSON(http.StatusAccepted, gin.H{
			"client_id": clientID,
		})
	})

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

func addScore(c *redis.Client, p map[string]interface{}) (map[string]interface{}, error) {

	ctx := context.TODO()

	// clientId := p["clientId"].(string)
	clientId := p["client_id"].(string)
	score := p["score"].(float64)

	//Validate data here in a production environment

	err := c.ZAdd(ctx, leaderboardKey, redis.Z{
		Score:  score,
		Member: clientId,
	}).Err()

	if err != nil {
		return nil, err
	}

	rank := c.ZRank(ctx, leaderboardKey, clientId)

	if err != nil {
		return nil, err
	}

	response := map[string]interface{}{
		"data": map[string]interface{}{
			"nickname": p["nickname"].(string),
			"rank":     rank.Val(),
		},
	}

	return response, nil
}
