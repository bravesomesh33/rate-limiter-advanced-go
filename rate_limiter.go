package ratelimiter

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/xid"
	"golang.org/x/net/context"
)

const (
	windowSizeInHours        = 24
	maxWindowRequestCount    = 110
	windowLogIntervalInHours = 1
)

var (
	redisClient *redis.Client
	ctx         = context.Background()
)

func init() {
	redisClient = redis.NewClient(&redis.Options{
		Addr: "localhost:6379", // Adjust if necessary
	})
}

type requestLog struct {
	RequestTimeStamp int64 `json:"requestTimeStamp"`
	RequestCount     int   `json:"requestCount"`
}

// RateLimiterUsingRedis is a middleware for rate limiting using Redis.
func RateLimiterUsingRedis() fiber.Handler {
	return func(c *fiber.Ctx) error {
		currentRequestTime := time.Now().Unix()
		ip := c.IP()

		// Fetch records of the current user using IP address
		record, err := redisClient.Get(ctx, ip).Result()
		if err == redis.Nil {
			// No record found, create a new record
			newRecord := []requestLog{{RequestTimeStamp: currentRequestTime, RequestCount: 1}}
			recordJSON, _ := json.Marshal(newRecord)
			redisClient.Set(ctx, ip, recordJSON, windowSizeInHours*time.Hour)
			return c.Next()
		} else if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Internal Server Error"})
		}

		// Parse the record and calculate the number of requests within the last window
		var data []requestLog
		json.Unmarshal([]byte(record), &data)
		windowStartTimestamp := time.Now().Add(-windowSizeInHours * time.Hour).Unix()
		var requestsWithinWindow []requestLog
		var totalWindowRequestsCount int

		for _, entry := range data {
			if entry.RequestTimeStamp > windowStartTimestamp {
				requestsWithinWindow = append(requestsWithinWindow, entry)
				totalWindowRequestsCount += entry.RequestCount
			}
		}

		// If the number of requests made is greater than or equal to the maximum, return an error
		if totalWindowRequestsCount >= maxWindowRequestCount {
			return c.Status(http.StatusTooManyRequests).JSON(fiber.Map{"error": "You have exceeded the maximum number of requests in the allowed time frame."})
		}

		// Log a new entry or increment the counter if the interval has not passed
		lastRequestLog := requestsWithinWindow[len(requestsWithinWindow)-1]
		potentialCurrentWindowIntervalStartTimeStamp := time.Now().Add(-windowLogIntervalInHours * time.Hour).Unix()

		if lastRequestLog.RequestTimeStamp > potentialCurrentWindowIntervalStartTimeStamp {
			lastRequestLog.RequestCount++
			data[len(data)-1] = lastRequestLog
		} else {
			data = append(data, requestLog{RequestTimeStamp: currentRequestTime, RequestCount: 1})
		}

		recordJSON, _ := json.Marshal(data)
		redisClient.Set(ctx, ip, recordJSON, windowSizeInHours*time.Hour)

		return c.Next()
	}
}
