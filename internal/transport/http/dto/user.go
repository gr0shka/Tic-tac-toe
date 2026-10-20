package dto

import "github.com/google/uuid"

type UserInfoResponse struct {
	ID    uuid.UUID `json:"id"`
	Login string    `json:"login"`
}
