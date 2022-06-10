package api

import (
	"bytes"
	"dcard-pretest/pkg/model"
	"dcard-pretest/pkg/store"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/go-redis/redismock/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TestSuite struct {
	suite.Suite
	mockClient redismock.ClientMock
	Store      *store.Store
}

func (t *TestSuite) SetupTest() {
	db, mock := redismock.NewClientMock()
	t.mockClient = mock
	t.Store = &store.Store{Client: db}
}

func (t *TestSuite) TestGetTop10WithoutValue() {
	router := setupRouter(t.Store)
	res := []redis.Z{}
	t.mockClient.ExpectZRevRangeWithScores(leaderboardKey, 0, 9).SetVal(res)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/leaderboard", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t.T(), http.StatusOK, w.Code)
	assert.Contains(t.T(), w.Body.String(), "{\"topPlayers\":[]}")

}
func (t *TestSuite) TestGetTop10WithValue() {
	router := setupRouter(t.Store)
	res := []redis.Z{
		{
			Member: "a",
			Score:  10,
		},
		{
			Member: "b",
			Score:  20,
		},
		{
			Member: "c",
			Score:  30,
		},
	}
	t.mockClient.ExpectZRevRangeWithScores(leaderboardKey, 0, 9).SetVal(res)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/leaderboard", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t.T(), http.StatusOK, w.Code)
	assert.Contains(t.T(), w.Body.String(), "{\"topPlayers\":[{\"clientId\":\"a\",\"score\":10},{\"clientId\":\"b\",\"score\":20},{\"clientId\":\"c\",\"score\":30}]}")
}

func (t *TestSuite) TestPostScore() {
	router := setupRouter(t.Store)
	expectedScore := model.Score{
		ClientId: "client1",
		Score:    1.23,
	}
	z := &redis.Z{
		Member: expectedScore.ClientId,
		Score:  expectedScore.Score,
	}
	t.mockClient.ExpectZAdd(leaderboardKey, z).SetVal(1)

	w := httptest.NewRecorder()
	body, _ := json.Marshal(model.Score{Score: expectedScore.Score})
	req, _ := http.NewRequest("POST", "/api/v1/score", bytes.NewReader(body))
	req.Header.Add("ClientId", expectedScore.ClientId)

	router.ServeHTTP(w, req)

	assert.Equal(t.T(), http.StatusOK, w.Code)

	assert.Contains(t.T(), w.Body.String(), "{\"status\":\"ok\"}")
}

func (t *TestSuite) TestPostScoreWithInvalidHeader() {
	router := setupRouter(t.Store)
	expectedScore := model.Score{
		ClientId: "",
		Score:    1.23,
	}
	z := &redis.Z{
		Member: expectedScore.ClientId,
		Score:  expectedScore.Score,
	}
	t.mockClient.ExpectZAdd(leaderboardKey, z).SetVal(1)

	w := httptest.NewRecorder()
	body, _ := json.Marshal(model.Score{Score: expectedScore.Score})
	req, _ := http.NewRequest("POST", "/api/v1/score", bytes.NewReader(body))
	req.Header.Add("ClientId", expectedScore.ClientId)

	router.ServeHTTP(w, req)

	assert.Equal(t.T(), http.StatusUnauthorized, w.Code)
}

func TestExampleTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
