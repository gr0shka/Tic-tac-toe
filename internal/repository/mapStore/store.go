package mapStore

import (
	"errors"
	"sync"

	"github.com/google/uuid"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/models"
	"github.com/gr0shka/Tic-tac-toe/internal/repository/dto"
	"github.com/gr0shka/Tic-tac-toe/internal/repository/mapper"
)

type MapRepository struct {
	data sync.Map
}

func NewMapRepository() *MapRepository {
	return &MapRepository{}
}

func (m *MapRepository) Save(cg models.CurrentGame) error {
	cgd := mapper.DomainToDTO(cg)
	m.data.Store(cgd.ID, cgd)

	return nil
}

func (m *MapRepository) Get(id uuid.UUID) (*models.CurrentGame, error) {
	cgd, ok := m.data.Load(id)
	if !ok {
		return nil, errors.New("not found")
	}

	cg := cgd.(dto.CurrentGameDTO)

	return mapper.DTOToDomain(cg), nil
}
