package handlers

import (
	"chesscourse/internal/apperror"
	"chesscourse/internal/dto"
	"chesscourse/internal/service/user"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserHandler struct {
	userService user.UserService
}

func NewUserHandler(userService user.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// GET /api/admin/users - получить всех пользователей (только админ)
func (h *UserHandler) GetUsers(c *gin.Context) {
	users, err := h.userService.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get users"})
		return
	}

	response := make([]dto.UserResponse, len(users))
	for i, u := range users {
		response[i] = dto.UserResponse{
			ID:       u.ID,
			Name:     u.Name,
			LastName: u.LastName,
			Email:    u.Email,
			Role:     u.Role,
		}
	}

	c.JSON(http.StatusOK, response)
}

// PUT /api/admin/users/:id/role - обновить роль пользователя (только админ)
func (h *UserHandler) UpdateUserRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	var req dto.UpdateUserRoleRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	err = h.userService.UpdateUserRole(id, req.Role)
	if errors.Is(err, apperror.ErrUserNotFound) || errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	if errors.Is(err, apperror.ErrInvalidRole) || errors.Is(err, user.ErrInvalidProfile) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user role"})
		return
	}

	c.Status(http.StatusNoContent)
}

func userError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, user.ErrSelfDelete):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, user.ErrInvalidProfile):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, gorm.ErrRecordNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user operation failed"})
	}
}

// UpdateUser godoc
// @Summary Изменить имя, фамилию или роль пользователя
// @Tags users
// @Security BearerAuth
// @Accept json
// @Param id path int true "ID пользователя"
// @Param request body dto.UpdateUserRequest true "Поля пользователя"
// @Success 204
// @Router /api/admin/users/{id} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if err := h.userService.UpdateUser(c.Request.Context(), id, req); err != nil {
		userError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// DeleteUser godoc
// @Summary Удалить пользователя и освободить места
// @Tags users
// @Security BearerAuth
// @Param id path int true "ID пользователя"
// @Success 204
// @Router /api/admin/users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	if err := h.userService.DeleteUser(c.Request.Context(), c.GetInt("userID"), id); err != nil {
		userError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
