package mapstore

import (
	"sync"

	"github.com/google/uuid"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/models"
)

type Repository struct {
	data sync.Map
}

func NewMapRepository() *Repository {
	return &Repository{}
}

func (m *Repository) Save(cg *models.CurrentGame) error {
	cgd := DomainToDTO(*cg)
	m.data.Store(cgd.ID, cgd)

	return nil
}

func (m *Repository) Get(id uuid.UUID) (*models.CurrentGame, error) {
	cgd, ok := m.data.Load(id)
	if !ok {
		return nil, models.ErrNotFound
	}

	cg := cgd.(CurrentGameDTO)

	return DTOToDomain(cg), nil
}
