package api

import (
	"net/http"

	"github.com/gorilla/mux"
)

func RegisterRoutes(h *Handler) http.Handler {
	router := mux.NewRouter()

	router.HandleFunc("/api/auth", h.Auth).Methods("POST")

	router.HandleFunc("/api/createUser", h.CreateUser).Methods("POST")

	router.HandleFunc("/api/info", h.Info).Methods("GET")

	router.HandleFunc("/api/sendCoin", h.SendCoin).Methods("POST")

	router.HandleFunc("/api/buy/{item}", h.BuyMerch).Methods("GET")

	return router
}
