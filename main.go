package main

import (
	"fmt"
	"log"
	"os"

	_ "github.com/GoogleCloudPlatform/cloudsql-proxy/proxy/dialers/postgres"
	"gorm.io/gorm"

	"github.com/johannaojeling/go-rest-api/pkg/api"
	"github.com/johannaojeling/go-rest-api/pkg/database"
	"github.com/johannaojeling/go-rest-api/pkg/models"
	"github.com/johannaojeling/go-rest-api/pkg/repositories"
)

var (
	dbDriver = os.Getenv("DB_DRIVER")
	dbUrl    = os.Getenv("DB_URL")
	port     = os.Getenv("PORT")
)

func init() {
	if dbDriver == "" {
		dbDriver = "postgres"
	}
	if port == "" {
		port = "8080"
	}
}

func main() {
	gormDB, err := setUpDatabase(dbDriver, dbUrl)
	if err != nil {
		log.Fatalf("error setting up database: %v", err)
	}

	userRepository := repositories.NewSQLUserRepository(gormDB)
	app := api.NewApp(userRepository)

	log.Printf("listening on port %s\n", port)
	if err := app.Run(":" + port); err != nil {
		log.Fatalf("error running app: %v", err)
	}
}

func setUpDatabase(driver string, dsn string) (*gorm.DB, error) {
	gormDB, err := database.GetConnection(driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("error getting database connection: %v", err)
	}

	err = gormDB.AutoMigrate(&models.User{})
	if err != nil {
		return nil, fmt.Errorf("error migrating database: %v", err)
	}
	return gormDB, nil
}
