package main

import (
	"net/http"
	"sync/atomic"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	stats := CreateStats()

	app := fiber.New()

	app.Use(logger.New())

	app.Get("/a", func(c *fiber.Ctx) error {
		stats.Increment("a")

		return c.Status(http.StatusNoContent).Send([]byte(``))
	})
	app.Get("/b", func(c *fiber.Ctx) error {
		stats.Increment("b")

		return c.Status(http.StatusNoContent).Send([]byte(``))
	})
	app.Get("/stats", func(c *fiber.Ctx) error {
		return c.JSON(stats)
	})

	app.Listen(":3000")
}

type Stats struct {
	Total      *int64            `json:"total"`
	Individual map[string]*int64 `json:"individual"`
}

func CreateStats() *Stats {
	val := Stats{
		Total:      new(int64),
		Individual: make(map[string]*int64),
	}

	val.Individual["a"] = new(int64)
	val.Individual["b"] = new(int64)

	return &val
}

func (s *Stats) Increment(key string) {
	if _, ok := s.Individual[key]; !ok {
		s.Individual[key] = new(int64)
	}

	atomic.AddInt64(s.Individual[key], 1)
	atomic.AddInt64(s.Total, 1)
}
