package endpoints

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/johannaojeling/go-rest-api/pkg/models"
	"github.com/johannaojeling/go-rest-api/pkg/schemas"
)

type UsersHandler struct {
	DB *gorm.DB
}

func NewUsersHandler(db *gorm.DB) *UsersHandler {
	return &UsersHandler{
		DB: db,
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

	user := userRequestToUser(userRequest)
	err = handler.DB.Create(&user).Error
	if err != nil {
		ctx.AbortWithStatusJSON(
			http.StatusInternalServerError,
			models.NewErrorMessage("error creating user"),
		)
		return
	}

	userResponse := userToUserResponse(user)
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

	var user models.User
	err = handler.DB.Where("id = ?", id).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
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

	userResponse := userToUserResponse(user)
	ctx.JSON(http.StatusOK, userResponse)
}

func (handler *UsersHandler) GetAllUsers(ctx *gin.Context) {
	var users []models.User
	handler.DB.Find(&users)

	userResponseList := make([]schemas.UserResponse, len(users))
	for i, user := range users {
		userResponseList[i] = userToUserResponse(user)
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

	var user models.User
	err = handler.DB.Where("id = ?", id).First(&user).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		user = userRequestToUser(userRequest)
		user.Id = id

		err = handler.DB.Create(&user).Error
		if err != nil {
			ctx.AbortWithStatusJSON(
				http.StatusInternalServerError,
				models.NewErrorMessage("error creating user"),
			)
			return
		}

		userResponse := userToUserResponse(user)
		ctx.JSON(http.StatusCreated, userResponse)
		return
	}

	if err != nil {
		ctx.AbortWithStatusJSON(
			http.StatusInternalServerError,
			models.NewErrorMessage("error retrieving user"),
		)
		return
	}

	userUpdates := userRequestToUser(userRequest)
	err = handler.DB.Model(&user).Updates(userUpdates).Error
	if err != nil {
		ctx.AbortWithStatusJSON(
			http.StatusInternalServerError,
			models.NewErrorMessage("error updating user"),
		)
		return
	}

	userResponse := userToUserResponse(user)
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

	var user models.User
	err = handler.DB.Where("id = ?", id).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
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

	err = handler.DB.Delete(&user).Error
	if err != nil {
		ctx.AbortWithStatusJSON(
			http.StatusInternalServerError,
			models.NewErrorMessage("error deleting user"),
		)
		return
	}

	ctx.Status(http.StatusNoContent)
}

func userRequestToUser(userRequest schemas.UserRequest) models.User {
	return models.User{
		FirstName: userRequest.FirstName,
		LastName:  userRequest.LastName,
		Email:     userRequest.Email,
	}
}

func userToUserResponse(user models.User) schemas.UserResponse {
	return schemas.UserResponse{
		Id:        user.Id,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
	}
}
