package dto

import "github.com/google/uuid"

type SignUpRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type UserInfoResponse struct {
	ID    uuid.UUID `json:"id"`
	Login string    `json:"login"`
}
