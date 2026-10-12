package game

import "errors"

var (
	ErrInvalidMove         = errors.New("invalid move")
	ErrInvalidArgument     = errors.New("invalid argument")
	ErrNotFound            = errors.New("not found")
	ErrGameFull            = errors.New("game is full")
	ErrPlayerAlreadyInGame = errors.New("player already in game")
)
