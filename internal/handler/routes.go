package handler

import "net/http"

// RegisterRoutes registers all API routes.
func RegisterRoutes(userHandler *UserHandler) {
	http.HandleFunc("/health", Health)

	http.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			userHandler.GetAll(w, r)
		} else if r.Method == http.MethodPost {
			userHandler.Create(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/user", userHandler.GetByID)
}
