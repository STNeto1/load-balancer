package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sync/atomic"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = 100

	end1 := &Endpoint{
		Label:       "end1",
		Destination: "http://127.0.0.1:3000/a",
	}
	end2 := &Endpoint{
		Label:       "end2",
		Destination: "http://127.0.0.1:3000/b",
	}

	balancer := CreateBalancer()
	balancer.Register(end1)
	balancer.Register(end2)

	app := fiber.New()

	app.Use(logger.New())

	app.All("/", func(c *fiber.Ctx) error {
		end := balancer.Next()
		defer balancer.Return(end)

		httpReq, err := http.NewRequest(http.MethodGet, end.Destination, c.Request().BodyStream())
		if err != nil {
			fmt.Println(err.Error())
			return c.Status(http.StatusInternalServerError).SendString(err.Error())
		}

		res, err := http.DefaultClient.Do(httpReq)
		if err != nil {
			fmt.Println(err.Error())
			return c.Status(http.StatusInternalServerError).SendString(err.Error())
		}

		resBody, err := io.ReadAll(res.Body)
		if err != nil {
			fmt.Println(err.Error())
			return c.Status(http.StatusInternalServerError).SendString(err.Error())
		}

		return c.Status(res.StatusCode).Send(resBody)
	})

	app.Listen(":8080")
}

type Endpoint struct {
	Label       string
	Destination string
	Weight      int64
}

type Balancer struct {
	Endpoints   []*Endpoint
	Load        map[string]*int64
	TotalWeight int64
}

func CreateBalancer() *Balancer {
	return &Balancer{
		Endpoints:   []*Endpoint{},
		Load:        map[string]*int64{},
		TotalWeight: 0,
	}
}

func (b *Balancer) Register(e *Endpoint) {
	b.Endpoints = append(b.Endpoints, e)
	b.Load[e.Label] = new(int64)
	b.TotalWeight += e.Weight
}

func (b *Balancer) Next() *Endpoint {

	val := randomValue(b.TotalWeight)

	for _, e := range b.Endpoints {
		val -= e.Weight
		if val <= 0 {
			return e
		}
	}

	// Fallback
	return b.Endpoints[0]
}

func (b *Balancer) Return(e *Endpoint) {
	atomic.AddInt64(b.Load[e.Label], -1)
}

func randomValue(max int64) int64 {
	return 1 + rand.Int63n(max)
}
