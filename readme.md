## Getting Started

### Development Environment

- Go version go1.17.8 darwin/amd64
- Redis version 6.2.7

### Start the services

```
$ docker run --name some-redis -p 6379:6379 -d redis:6.2.7
$ go run ./service/api
$ go run ./service/cron
```

### Endpoints

- `curl -X GET http://localhost/api/v1/leaderboard`
- `curl -X POST -H "Content-Type:application/json" -H "ClientId: clientid" http://localhost/api/v1/score -d '{"score": 87.7}'`

### Notes

- 使用 gin framework 建立 API 配合 Redis Sorted Set 作為 storage
    - Sorted Set 有天然的 unique 效果一個 client 只能建立一個 score
    - 若有複雜的計算邏輯再考慮改用 RDS + Redis (快取)
- nice to have
    - 清除 leaderboard，目前是依照 cron 啟動時間每 10 分鐘清除一次
        - 可以考慮設定 20 分鐘以上的預設 TTL 避免 cron 失效，若 redis 被塞滿影響到其他 record 被清掉
        - 設定成整點開始計算 10 分鐘（00, 10, 20, ...）使用者比較能預期紀錄被清除的時間
    - redis config (password, port, db...) 目前是寫在 code 裡面，抽出 config 日後比較好維護
    - 執行 service 目前需要自行安裝 golang 環境以及編譯，包成 docker + makefile 體驗跟維護性會比較好一些