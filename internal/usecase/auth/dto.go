package auth

import "time"

type SignUpRequest struct {
	Login    string
	Password string
}

type JwtResponse struct {
	Type                  string
	AccessToken           string
	RefreshToken          string
	RefreshTokenExpiresAt time.Time
}
