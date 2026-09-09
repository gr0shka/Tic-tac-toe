package main

import (
	"github.com/gr0shka/Tic-tac-toe/internal/di"
	"go.uber.org/fx"
)

func main() {
	fx.New(di.CreateApp()).Run()
}
