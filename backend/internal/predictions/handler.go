package predictions

import (
	"encoding/json"
	"net/http"

	"github.com/matheuscoutinhoo/better/internal/models"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) GetPredictions(w http.ResponseWriter, r *http.Request) {
	predictions, err := h.svc.GetPredictions()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to fetch predictions"})
		return
	}
	if predictions == nil {
		predictions = []models.AIPrediction{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(predictions)
}
