package main

import (
	"net/http"
)

func (aCfg *apiConfig) handleReset(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	err := aCfg.db.ResetDatabase(r.Context())
	if err != nil {
		respondWithError(w, 500, "failed to reset database")
		return

	}
	w.WriteHeader(http.StatusOK)
	aCfg.fileServerHits.Store(0)
	w.Write([]byte("hits reset to 0"))
}
func (af *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		af.fileServerHits.Add(1)
		next.ServeHTTP(w, r)
	})
}
