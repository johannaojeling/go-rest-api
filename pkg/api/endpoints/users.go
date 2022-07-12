package endpoints

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/johannaojeling/go-rest-api/pkg/models"
	"github.com/johannaojeling/go-rest-api/pkg/repositories"
	"github.com/johannaojeling/go-rest-api/pkg/schemas"
)

type UsersHandler struct {
	UsersRepository repositories.UserRepository
}

func NewUsersHandler(usersRepository repositories.UserRepository) *UsersHandler {
	return &UsersHandler{
		UsersRepository: usersRepository,
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
		ctx.AbortWithStatusJSON(
			http.StatusBadRequest,
			models.NewErrorMessage("invalid request body"),
		)
		return
	}

	user := userRequestToUserModel(userRequest)
	err = handler.UsersRepository.CreateUser(user)
	if err != nil {
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
		ctx.AbortWithStatusJSON(
			http.StatusBadRequest,
			models.NewErrorMessage("invalid uri, expecting id"),
		)
		return
	}
	id := userUri.Id

	user, err := handler.UsersRepository.GetUserById(id)
	if errors.Is(err, repositories.ErrUserNotFound) {
		ctx.AbortWithStatusJSON(
			http.StatusNotFound,
			models.NewErrorMessage("no user with id %q exists", id),
		)
		return
	}
	if err != nil {
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
	allUsers, err := handler.UsersRepository.GetAllUsers()
	if err != nil {
		ctx.AbortWithStatusJSON(
			http.StatusInternalServerError,
			models.NewErrorMessage("error retrieving users"),
		)
		return
	}

	userResponseList := make([]schemas.UserResponse, len(allUsers))
	for i, user := range allUsers {
		userResponseList[i] = userModelToUserResponse(user)
	}

	ctx.JSON(http.StatusOK, userResponseList)
}

func (handler *UsersHandler) UpdateUser(ctx *gin.Context) {
	var userUri schemas.UserURI
	err := ctx.BindUri(&userUri)
	if err != nil {
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
		ctx.AbortWithStatusJSON(
			http.StatusBadRequest,
			models.NewErrorMessage("invalid request body"),
		)
		return
	}

	updates := userRequestToUserModel(userRequest)
	updatedUser, err := handler.UsersRepository.UpdateUserById(id, updates)

	if errors.Is(err, repositories.ErrUserNotFound) {
		newUser := &models.User{
			Id:        id,
			FirstName: updates.FirstName,
			LastName:  updates.LastName,
			Email:     updates.Email,
		}
		err = handler.UsersRepository.CreateUser(newUser)
		if err != nil {
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
		ctx.AbortWithStatusJSON(
			http.StatusBadRequest,
			models.NewErrorMessage("invalid uri, expecting id"),
		)
		return
	}
	id := userUri.Id

	err = handler.UsersRepository.DeleteUser(id)
	if errors.Is(err, repositories.ErrUserNotFound) {
		ctx.AbortWithStatusJSON(
			http.StatusNotFound,
			models.NewErrorMessage("no user with id %q exists", id),
		)
		return
	}
	if err != nil {
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
