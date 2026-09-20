package handlers

import (
	"chesscourse/internal/apperror"
	"chesscourse/internal/dto"
	"chesscourse/internal/service/auth"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService auth.AuthService
}

func NewAuthHandler(authService auth.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register godoc
//
//	@Summary		Регистрация нового пользователя
//	@Description	Создаёт нового пользователя с email, именем и паролем. Возвращает 201 при успехе.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body	dto.RegisterRequest	true	"Данные для регистрации"
//	@Success		201
//	@Failure		400	{object}	map[string]string	"Некорректный запрос"
//	@Failure		409	{object}	map[string]string	"Email уже зарегистрирован"
//	@Failure		500	{object}	map[string]string	"Внутренняя ошибка сервера"
//	@Router			/api/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	err := h.authService.Register(req.Email, req.Name, req.Password)
	if err != nil {
		if errors.Is(err, apperror.ErrEmailExists) {
			c.JSON(http.StatusConflict, gin.H{"error": "email already exists"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "registration failed"})
			return
		}
	}

	c.Status(http.StatusCreated)
}

// Login godoc
//
//	@Summary		Вход пользователя
//	@Description	Проверяет email и пароль, возвращает JWT access token.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.LoginRequest		true	"Данные для входа"
//	@Success		200		{object}	map[string]string		"JWT токен"
//	@Failure		400		{object}	map[string]string		"Некорректный запрос"
//	@Failure		401		{object}	map[string]string		"Неверный email или пароль"
//	@Failure		500		{object}	map[string]string		"Внутренняя ошибка сервера"
//	@Router			/api/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	token, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		if errors.Is(err, apperror.ErrBadCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid password or email"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "login failed"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

// Me godoc
//
//	@Summary		Текущий пользователь
//	@Description	Возвращает данные авторизованного пользователя по JWT токену.
//	@Tags			auth
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	dto.MeResponse
//	@Failure		401	{object}	map[string]string	"Не авторизован"
//	@Router			/api/auth/me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	user, err := h.authService.GetUserByID(userID.(int))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, dto.MeResponse{
		ID:    user.ID,
		Email: user.Email,
		Name:  user.Name,
		Role:  user.Role,
	})
}
