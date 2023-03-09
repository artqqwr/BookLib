package main

import (
	"log"

	"github.com/artqqwr/bookslib/internal/app"
	"github.com/artqqwr/bookslib/internal/db"
	"github.com/artqqwr/bookslib/internal/handlers"
	"github.com/artqqwr/bookslib/pkg/config"
)

func main() {
	config := config.GetConfig()

	d := db.New(config)
	db.AutoMigrate(d)

	app := app.CreateApp()

	handlers := []handlers.Handler{}

	for _, handler := range handlers {
		handler.Routes(app)
	}

	err := app.Listen(config.Server.Address)
	if err != nil {
		log.Fatal(err)
	}
}
