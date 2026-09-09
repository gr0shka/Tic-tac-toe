package main

import (
	"net/http"

	appService "github.com/gr0shka/Tic-tac-toe/internal/application/service"
	domainService "github.com/gr0shka/Tic-tac-toe/internal/domain/service"
	"github.com/gr0shka/Tic-tac-toe/internal/repository/mapStore"
	transportHttp "github.com/gr0shka/Tic-tac-toe/internal/transport/http"
)

func main() {
	repo := mapStore.NewMapRepository()
	gs := domainService.NewGameService()
	as := appService.NewAppService(gs, repo)
	handler := transportHttp.NewHandler(as)

	mux := http.NewServeMux()
	mux.HandleFunc("/game/{uuid}", handler.NextTurn)
	mux.HandleFunc("/create", handler.NewGame)
	mux.HandleFunc("/get/{uuid}", handler.GetGame)
	http.ListenAndServe(":8080", mux)
}
