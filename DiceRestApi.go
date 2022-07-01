package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"log"
	"math/rand"
	"time"
)

func main() {
	app := fiber.New()

	app.Get("/roll", func(c *fiber.Ctx) error {
		rand.Seed(time.Now().UnixNano())
		dice := rand.Intn(6) + 1
		msg := fmt.Sprintf("%d", dice)
		return c.SendString(msg)
	})

	log.Fatal(app.Listen(":3000"))
}
