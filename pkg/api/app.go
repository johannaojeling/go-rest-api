package api

import (
	_ "github.com/GoogleCloudPlatform/cloudsql-proxy/proxy/dialers/postgres"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/johannaojeling/go-rest-api/pkg/api/endpoints"
	"github.com/johannaojeling/go-rest-api/pkg/models"
)

type App struct {
	DB     *gorm.DB
	Router *gin.Engine
}

func NewApp(driver string, dsn string) (*App, error) {
	var app App
	err := app.connectToDB(driver, dsn)
	if err != nil {
		return nil, err
	}

	app.registerHandlers()
	return &app, nil
}

func (app *App) connectToDB(driver string, dbUrl string) error {
	dialector := postgres.New(postgres.Config{
		DriverName: driver,
		DSN:        dbUrl,
	})
	db, err := gorm.Open(dialector)
	if err != nil {
		return err
	}

	err = db.AutoMigrate(new(models.User))
	if err != nil {
		return err
	}

	app.DB = db
	return nil
}

func (app *App) registerHandlers() {
	app.Router = gin.Default()
	group := app.Router.Group("/users")
	endpoints.NewUsersHandler(app.DB).Register(group)
}

func (app *App) Run(port string) error {
	return app.Router.Run(port)
}
