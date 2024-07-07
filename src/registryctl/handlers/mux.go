package handlers

import "github.com/gorilla/mux"

func NewMuxRouter() *mux.Router {
	return mux.NewRouter()
}
