package handler

import (
	"log/slog"
	"net/http"

	. "todos/internal/adapter/http/helper"
	. "todos/internal/adapter/http/validation"
	"todos/internal/core/domain"
	"todos/internal/core/model/request"
	"todos/internal/core/model/response"
	"todos/internal/core/port"
	"todos/internal/core/util"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	svc port.UserService
}

func NewUserHandler(svc port.UserService) *UserHandler {
	return &UserHandler{
		svc: svc,
	}
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	ctx := c.Request.Context()

	var params request.UserRequest

	if err := c.ShouldBindJSON(&params); err != nil {
		SendBadRequestError(c, "request", "Invalid request parameters")
		return
	}

	if err := Validator.Struct(params); err != nil {
		SendValidationError(c, err)
		return
	}

	encrypted, err := util.GenerateEncrypt(params.Password)

	if err != nil {
		SendBadRequestError(c, "request", "Invalid request parameters")
		return
	}

	user := domain.User{
		Name:              params.Name,
		Email:             params.Email,
		EncryptedPassword: encrypted,
	}

	savedUser, err := h.svc.Create(ctx, user)

	if err != nil {
		slog.Error("Error creating user", "error", err)
		SendBadRequestError(c, "request", "Invalid request parameters")
		return
	}

	response := response.UserResponse{
		UUID:      savedUser.UUID.String(),
		Name:      savedUser.Name,
		Email:     savedUser.Email,
		CreatedAt: savedUser.CreatedAt,
		UpdatedAt: savedUser.UpdatedAt,
	}

	SendSuccess(c, http.StatusCreated, response)
}

func (h *UserHandler) DeleteByUUID(c *gin.Context) {
	ctx := c.Request.Context()
	err := h.svc.DeleteByUUID(ctx, c.Param("uuid"))

	if err != nil {
		slog.Error("Error deleting user", "error", err)
		SendInternalError(c, err.Error())
		return
	}

	response := map[string]string{
		"message": "User deleted successfully",
	}

	SendSuccess(c, http.StatusOK, response)
}
