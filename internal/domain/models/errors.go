package models

import "errors"

var (
	ErrInvalidTurn          = errors.New("invalid turn")
	ErrWrongNumberOfTurns   = errors.New("wrong number of turns")
	ErrAlteredPreviousMoves = errors.New("altered previous moves")
	ErrMultipleMoves        = errors.New("multiple moves")
	ErrNoMoveMade           = errors.New("no move made")
	ErrWrongPlayerMove      = errors.New("wrong player move")
)
