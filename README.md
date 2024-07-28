
# Rate Limiter Package

This package provides middleware for rate limiting in a Go application using Redis. It includes a rate limiting middleware implemented using Redis.

## Installation

To install the package, you need to have Go installed. Then run the following command:

```sh
go get github.com/yourusername/rate-limiter-go/ratelimiter
```

## Usage

### Importing the Middleware

Import the middleware function in your Go application:

```go
import (
    "github.com/yourusername/rate-limiter-go/ratelimiter"
)
```

### Middleware Function

#### Rate Limiter Using Redis

This middleware uses Redis to limit the number of requests a user can make within a specified time window.

```go
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
```

### Applying the Middleware

To use the middleware function in your Go application, apply it to your routes:

```go
import (
    "github.com/gofiber/fiber/v2"
    "github.com/yourusername/rate-limiter-go/ratelimiter"
)

func main() {
    app := fiber.New()

    app.Use(ratelimiter.RateLimiterUsingRedis())

    app.Get("/", func(c *fiber.Ctx) error {
        return c.SendString("Welcome to the Rate Limited API")
    })

    app.Listen(":3000")
}
```

## Configuration

The rate limiting configuration is defined by the following constants in the `rate_limiter.go` file:

- `windowSizeInHours`: The time window size in hours.
- `maxWindowRequestCount`: The maximum number of requests allowed within the time window.
- `windowLogIntervalInHours`: The interval in hours for logging request timestamps.

## Dependencies

- `github.com/go-redis/redis/v8`: Redis client for Go.
- `github.com/gofiber/fiber/v2`: An Express-inspired web framework for Go.
- `github.com/rs/xid`: A globally unique id generator.

## License

This package is licensed under the ISC License.
