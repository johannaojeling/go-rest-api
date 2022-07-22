package endpoints

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/johannaojeling/go-rest-api/pkg/models"
	"github.com/johannaojeling/go-rest-api/pkg/repositories"
	"github.com/johannaojeling/go-rest-api/pkg/schemas"
)

type UsersHandler struct {
	userRepository repositories.UserRepository
}

func NewUsersHandler(userRepository repositories.UserRepository) *UsersHandler {
	return &UsersHandler{
		userRepository: userRepository,
	}
}

func (handler *UsersHandler) Register(routerGroup *gin.RouterGroup) {
	{
		routerGroup.POST("/", handler.CreateUser)
		routerGroup.GET("/:id", handler.GetUser)
		routerGroup.GET("/", handler.GetAllUsers)
		routerGroup.PUT("/:id", handler.UpdateUser)
		routerGroup.DELETE("/:id", handler.DeleteUser)
	}
}

func (handler *UsersHandler) CreateUser(ctx *gin.Context) {
	var userRequest schemas.UserRequest
	err := ctx.ShouldBindJSON(&userRequest)
	if err != nil {
		log.Printf("invalid request body: %v", err)
		ctx.AbortWithStatusJSON(
			http.StatusBadRequest,
			models.NewErrorMessage("invalid request body"),
		)
		return
	}

	user := userRequestToUserModel(userRequest)
	err = handler.userRepository.CreateUser(user)
	if err != nil {
		log.Printf("error creating user: %v", err)
		ctx.AbortWithStatusJSON(
			http.StatusInternalServerError,
			models.NewErrorMessage("error creating user"),
		)
		return
	}

	userResponse := userModelToUserResponse(user)
	ctx.JSON(http.StatusCreated, userResponse)
}

func (handler *UsersHandler) GetUser(ctx *gin.Context) {
	var userUri schemas.UserURI
	err := ctx.BindUri(&userUri)
	if err != nil {
		log.Printf("invalid uri: %v", err)
		ctx.AbortWithStatusJSON(
			http.StatusBadRequest,
			models.NewErrorMessage("invalid uri, expecting id"),
		)
		return
	}
	id := userUri.Id

	user, err := handler.userRepository.GetUserById(id)
	if errors.Is(err, repositories.ErrUserNotFound) {
		log.Printf("user not found: %v", err)
		ctx.AbortWithStatusJSON(
			http.StatusNotFound,
			models.NewErrorMessage("no user with id %q exists", id),
		)
		return
	}
	if err != nil {
		log.Printf("error getting user: %v", err)
		ctx.AbortWithStatusJSON(
			http.StatusInternalServerError,
			models.NewErrorMessage("error retrieving user"),
		)
		return
	}

	userResponse := userModelToUserResponse(user)
	ctx.JSON(http.StatusOK, userResponse)
}

func (handler *UsersHandler) GetAllUsers(ctx *gin.Context) {
	users, err := handler.userRepository.GetAllUsers()
	if err != nil {
		log.Printf("error getting users: %v", err)
		ctx.AbortWithStatusJSON(
			http.StatusInternalServerError,
			models.NewErrorMessage("error retrieving users"),
		)
		return
	}

	userResponseList := make([]schemas.UserResponse, len(users))
	for i, user := range users {
		userResponseList[i] = userModelToUserResponse(user)
	}

	ctx.JSON(http.StatusOK, userResponseList)
}

func (handler *UsersHandler) UpdateUser(ctx *gin.Context) {
	var userUri schemas.UserURI
	err := ctx.BindUri(&userUri)
	if err != nil {
		log.Printf("invalid uri: %v", err)
		ctx.AbortWithStatusJSON(
			http.StatusBadRequest,
			models.NewErrorMessage("invalid uri, expecting id"),
		)
		return
	}
	id := userUri.Id

	var userRequest schemas.UserRequest
	err = ctx.ShouldBindJSON(&userRequest)
	if err != nil {
		log.Printf("invalid request body: %v", err)
		ctx.AbortWithStatusJSON(
			http.StatusBadRequest,
			models.NewErrorMessage("invalid request body"),
		)
		return
	}

	updates := userRequestToUserModel(userRequest)
	updatedUser, err := handler.userRepository.UpdateUserById(id, updates)

	if errors.Is(err, repositories.ErrUserNotFound) {
		newUser := &models.User{
			Id:        id,
			FirstName: updates.FirstName,
			LastName:  updates.LastName,
			Email:     updates.Email,
		}
		err = handler.userRepository.CreateUser(newUser)
		if err != nil {
			log.Printf("error creating user: %v", err)
			ctx.AbortWithStatusJSON(
				http.StatusInternalServerError,
				models.NewErrorMessage("error creating user"),
			)
			return
		}

		userResponse := userModelToUserResponse(newUser)
		ctx.JSON(http.StatusCreated, userResponse)
		return
	}

	if err != nil {
		log.Printf("error updating user: %v", err)
		ctx.AbortWithStatusJSON(
			http.StatusInternalServerError,
			models.NewErrorMessage("error updating user"),
		)
		return
	}

	userResponse := userModelToUserResponse(updatedUser)
	ctx.JSON(http.StatusOK, userResponse)
}

func (handler *UsersHandler) DeleteUser(ctx *gin.Context) {
	var userUri schemas.UserURI
	err := ctx.BindUri(&userUri)
	if err != nil {
		log.Printf("invalid uri: %v", err)
		ctx.AbortWithStatusJSON(
			http.StatusBadRequest,
			models.NewErrorMessage("invalid uri, expecting id"),
		)
		return
	}
	id := userUri.Id

	err = handler.userRepository.DeleteUserById(id)
	if errors.Is(err, repositories.ErrUserNotFound) {
		log.Printf("user not found: %v", err)
		ctx.AbortWithStatusJSON(
			http.StatusNotFound,
			models.NewErrorMessage("no user with id %q exists", id),
		)
		return
	}
	if err != nil {
		log.Printf("error deleting user: %v", err)
		ctx.AbortWithStatusJSON(
			http.StatusInternalServerError,
			models.NewErrorMessage("error deleting user"),
		)
		return
	}

	ctx.Status(http.StatusNoContent)
}

func userRequestToUserModel(userRequest schemas.UserRequest) *models.User {
	return &models.User{
		FirstName: userRequest.FirstName,
		LastName:  userRequest.LastName,
		Email:     userRequest.Email,
	}
}

func userModelToUserResponse(user *models.User) schemas.UserResponse {
	return schemas.UserResponse{
		Id:        user.Id,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
	}
}
