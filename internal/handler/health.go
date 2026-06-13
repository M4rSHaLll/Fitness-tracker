package handler

import "net/http"

func NewHealthHandler(storageDriver string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{
			"status":  "ok",
			"storage": storageDriver,
		})
	}
}
