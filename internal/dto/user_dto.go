package dto

type UserResponse struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	LastName string `json:"last_name"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

type UpdateUserRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=Student Teacher Admin"`
}

type UpdateUserRequest struct {
	Name     *string `json:"name"`
	LastName *string `json:"last_name"`
	Role     *string `json:"role"`
}

type UsersListResponse struct {
	Users []UserResponse `json:"users"`
	Total int            `json:"total"`
}
