package model

type RegisterRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
	Role     string `json:"role" validate:"required,oneof=student teacher"`
}

type LoginRequest struct {
	Username    string `json:"username" validate:"required"`
	Password    string `json:"password" validate:"required"`
	DurationMin int    `json:"duration_min" validate:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
}
