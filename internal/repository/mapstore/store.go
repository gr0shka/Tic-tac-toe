package mapstore

import (
	"sync"

	"github.com/google/uuid"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/game"
)

type Repository struct {
	data sync.Map
}

func NewMapRepository() *Repository {
	return &Repository{}
}

func (m *Repository) Save(cg *game.CurrentGame) error {
	cgd := DomainToDTO(*cg)
	m.data.Store(cgd.ID, cgd)

	return nil
}

func (m *Repository) Get(id uuid.UUID) (*game.CurrentGame, error) {
	cgd, ok := m.data.Load(id)
	if !ok {
		return nil, game.ErrNotFound
	}

	cg := cgd.(CurrentGameDTO)

	return DTOToDomain(cg), nil
}
