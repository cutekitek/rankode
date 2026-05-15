package models

type CreateUserDTO struct {
	Username string `json:"username" validate:"required"`
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"` // Basic validation example
	Roles    int32  `json:"roles"    validate:"omitempty,oneof=0 3"`
}

type AuthUserDTO struct {
	Identifier string `json:"identifier" validate:"required"`
	Password   string `json:"password"   validate:"required"`
}
