package auth

type SignUpRequest struct {
	Login    string
	Password string
}

type JwtResponse struct {
	AccessToken  string
	RefreshToken string
}
