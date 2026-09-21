package user

import "github.com/gr0shka/Tic-tac-toe/internal/domain/user"

func DTOtoDomain(u DTO) *user.User {
	return user.New(u.ID, u.Login, u.Password)
}

func DomainToDTO(u *user.User) DTO {
	return DTO{u.ID(), u.Login(), u.Password()}
}
