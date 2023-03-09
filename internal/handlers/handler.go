package handlers

import "github.com/gofiber/fiber/v2"

type Handler interface {
	Routes(app *fiber.App)
}
