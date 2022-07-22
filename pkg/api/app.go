package api

import (
	"github.com/gin-gonic/gin"

	"github.com/johannaojeling/go-rest-api/pkg/api/endpoints"
	"github.com/johannaojeling/go-rest-api/pkg/repositories"
)

type App struct {
	router         *gin.Engine
	userRepository repositories.UserRepository
}

func NewApp(userRepository repositories.UserRepository) *App {
	app := &App{
		router:         gin.Default(),
		userRepository: userRepository,
	}
	app.registerHandlers()
	return app
}

func (app *App) registerHandlers() {
	userGroup := app.router.Group("/users")
	endpoints.NewUsersHandler(app.userRepository).Register(userGroup)
}

func (app *App) Run(port string) error {
	return app.router.Run(port)
}
