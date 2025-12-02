package main

import (
	"fmt"
	"log"
	"misw-api/db"
	"misw-api/endpoints"
	"misw-api/middleware"
	"net/http"

	"github.com/joho/godotenv"
)

const VERSION = "0.0.1"
const PORT = "8321"

func handle(route string, method string, handler http.HandlerFunc) {
	http.Handle(method+" "+route, middleware.ApplyTo(handler))
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	err = db.Init()
	if err != nil {
		log.Fatal("Error initializing database: " + err.Error())
	}

	handle("/", "GET", endpoints.IndexHandlerFactory(VERSION))

	handle("/user", "PUT", endpoints.CreateUserHandler)
	handle("/user", "GET", endpoints.GetUserHandler)

	handle("/game", "PUT", endpoints.NewGameHandler)
	handle("/game", "GET", endpoints.GetGameHandler)
	handle("/game", "POST", endpoints.MakeMoveHandler)
	handle("/games", "GET", endpoints.GetUnfinishedGamesHandler)

	fmt.Printf("Server is running on port %s\n", PORT)
	log.Fatal(http.ListenAndServe(":"+PORT, nil))
}
