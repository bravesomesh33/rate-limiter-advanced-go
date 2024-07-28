package main

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
