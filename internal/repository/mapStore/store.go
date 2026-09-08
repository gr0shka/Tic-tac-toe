package mapStore

import (
	"errors"
	"sync"

	"github.com/google/uuid"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/models"
	"github.com/gr0shka/Tic-tac-toe/internal/repository/DTO"
	"github.com/gr0shka/Tic-tac-toe/internal/repository/mapper"
)

type MapRepository struct {
	data sync.Map
}

func (m MapRepository) Save(cg models.CurrentGame) error {
	cgd := mapper.DomainToDTO(cg)
	m.data.Store(cgd.ID, cgd)

	return nil
}

func (m MapRepository) Get(id uuid.UUID) (models.CurrentGame, error) {
	cgd, ok := m.data.Load(id)
	if !ok {
		return models.CurrentGame{}, errors.New("not found")
	}

	return mapper.DTOToDomain(cgd.(DTO.CurrentGameDTO)), nil
}
