package apperror

import "errors"

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrContextNotEmpty = errors.New("cannot delete non-empty context")
	ErrUnauthorized    = errors.New("unauthorized")
	ErrBadCredentials  = errors.New("invalid password or email")
	ErrTaskIsOverdue   = errors.New("task deadline has passed")
	ErrToken           = errors.New("couldn't generate token")
	ErrEmailExists     = errors.New("email already exists")
	ErrInvalidToken    = errors.New("invalid token")

	ErrInvalidRole = errors.New("invalid role")
	ErrForbidden   = errors.New("access denied")

	ErrCategoryNotFound   = errors.New("category not found")
	ErrCategoryExists     = errors.New("category already exists")
	ErrCategoryHasCourses = errors.New("category has associated courses")
)
