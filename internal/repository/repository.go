package repository

import (
	"github.com/google/uuid"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/models"
)

type Repository interface {
	Save(cg *models.CurrentGame) error
	Get(id uuid.UUID) (*models.CurrentGame, error)
}
