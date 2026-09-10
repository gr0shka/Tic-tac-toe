package mapstore

import (
	"errors"
	"sync"

	"github.com/google/uuid"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/models"
)

type MapRepository struct {
	data sync.Map
}

func NewMapRepository() *MapRepository {
	return &MapRepository{}
}

func (m *MapRepository) Save(cg models.CurrentGame) error {
	cgd := DomainToDTO(cg)
	m.data.Store(cgd.ID, cgd)

	return nil
}

func (m *MapRepository) Get(id uuid.UUID) (*models.CurrentGame, error) {
	cgd, ok := m.data.Load(id)
	if !ok {
		return nil, errors.New("not found")
	}

	cg := cgd.(CurrentGameDTO)

	return DTOToDomain(cg), nil
}
