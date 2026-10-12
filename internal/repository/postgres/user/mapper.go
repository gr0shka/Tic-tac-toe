package user

import "github.com/gr0shka/Tic-tac-toe/internal/domain/user"

func DTOtoDomain(u userDTO) *user.User {
	return user.New(u.ID, u.Login, u.Password)
}

func DomainToDTO(u *user.User) userDTO {
	return userDTO{
		ID:       u.ID(),
		Login:    u.Login(),
		Password: u.Password(),
	}
}
