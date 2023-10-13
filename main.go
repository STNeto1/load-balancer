package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gofiber/fiber/v2"
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

	// app.Use(logger.New())

	app.All("/", func(c *fiber.Ctx) error {
		end := balancer.Next()
		// defer balancer.Return(end)

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
}

type Balancer struct {
	Endpoints []*Endpoint
}

func CreateBalancer() *Balancer {
	return &Balancer{
		Endpoints: []*Endpoint{},
	}
}

func (b *Balancer) Register(e *Endpoint) {
	b.Endpoints = append(b.Endpoints, e)
}

func (b *Balancer) Next() *Endpoint {
	first := b.Endpoints[0]
	b.Endpoints = append(b.Endpoints[1:], first)

	return first
}
