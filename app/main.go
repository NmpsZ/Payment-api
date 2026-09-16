package main

import (
	"log"

	"payment-backend/configs"
	"payment-backend/modules/servers"
	"payment-backend/pkg/databases"
)

func main() {
	config := configs.Load()
	if config.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	db, err := databases.Connect(config.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := servers.NewRouter(db).Run(":" + config.Port); err != nil {
		log.Fatal(err)
	}
}
