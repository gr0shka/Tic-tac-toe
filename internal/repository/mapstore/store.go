package mapstore

import (
	"sync"

	"github.com/google/uuid"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/game"
)

type Storage struct {
	data sync.Map
}

func NewStorage() *Storage {
	return &Storage{}
}

type repository struct {
	storage *Storage
}

func NewMapRepository(storage *Storage) *repository {
	return &repository{
		storage: storage,
	}
}

func (m *repository) Save(cg *game.CurrentGame) error {
	cgd := DomainToDTO(*cg)
	m.storage.data.Store(cgd.ID, cgd)

	return nil
}

func (m *repository) Get(id uuid.UUID) (*game.CurrentGame, error) {
	cgd, ok := m.storage.data.Load(id)
	if !ok {
		return nil, game.ErrNotFound
	}

	cg := cgd.(CurrentGameDTO)

	return DTOToDomain(cg), nil
}
