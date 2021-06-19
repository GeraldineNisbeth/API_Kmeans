package main

import (
	"API/kmeans/services"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	router := mux.NewRouter().StrictSlash(true)
	Cors(router)

	router.HandleFunc("/", services.HomeRoute)
	router.HandleFunc("/enviarInputs", services.PostInputs).Methods("POST")

	log.Fatal(http.ListenAndServe(":8081", router))
}

func Cors(router *mux.Router) {
	router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
	}).Methods(http.MethodOptions)
	router.Use(habilitarCors)
}

func habilitarCors(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Enconding, X-CSRF-Token, Authorization")
			next.ServeHTTP(w, req)
		})
}
