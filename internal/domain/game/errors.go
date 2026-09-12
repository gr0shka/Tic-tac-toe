package game

import "errors"

var (
	ErrInvalidTurn             = errors.New("invalid turn")
	ErrWrongNumberOfTurns      = errors.New("wrong number of turns")
	ErrAlteredPreviousMoves    = errors.New("altered previous moves")
	ErrMultipleMoves           = errors.New("multiple moves")
	ErrNoMoveMade              = errors.New("no move made")
	ErrWrongPlayerMove         = errors.New("wrong player move")
	ErrNotFound                = errors.New("not found")
	ErrFailedCalculateNextTurn = errors.New("failed to calculate next turn")
)
