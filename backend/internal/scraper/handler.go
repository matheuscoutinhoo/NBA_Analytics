package scraper

import (
	"net/http"
	"time"

	"github.com/better/backend/pkg/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// TriggerDailyScrape - Admin endpoint to manually trigger daily scrape
func (h *Handler) TriggerDailyScrape(w http.ResponseWriter, r *http.Request) {
	go func() {
		h.service.RunDailyScrape(r.Context())
	}()

	response.JSON(w, http.StatusAccepted, map[string]interface{}{
		"message":   "daily scrape triggered",
		"timestamp": time.Now().UTC(),
	})
}

// TriggerResultsUpdate - Admin endpoint for results update
func (h *Handler) TriggerResultsUpdate(w http.ResponseWriter, r *http.Request) {
	go func() {
		h.service.RunResultsUpdate(r.Context())
	}()

	response.JSON(w, http.StatusAccepted, map[string]interface{}{
		"message":   "results update triggered",
		"timestamp": time.Now().UTC(),
	})
}

// GetScrapingJobs - Admin endpoint listing scraping job history
func (h *Handler) GetScrapingJobs(w http.ResponseWriter, r *http.Request) {
	rows, err := h.service.db.Query(r.Context(),
		`SELECT id, job_type, status, idempotency_key, records_processed, error_message, started_at, completed_at, created_at
		FROM scraping_jobs ORDER BY created_at DESC LIMIT 50`,
	)
	if err != nil {
		response.InternalError(w)
		return
	}
	defer rows.Close()

	var jobs []map[string]interface{}
	for rows.Next() {
		var id, jobType, status, key string
		var recordsProcessed *int
		var errMsg *string
		var startedAt, completedAt *time.Time
		var createdAt time.Time

		if err := rows.Scan(&id, &jobType, &status, &key, &recordsProcessed, &errMsg, &startedAt, &completedAt, &createdAt); err != nil {
			continue
		}

		jobs = append(jobs, map[string]interface{}{
			"id":                id,
			"job_type":          jobType,
			"status":            status,
			"idempotency_key":   key,
			"records_processed": recordsProcessed,
			"error_message":     errMsg,
			"started_at":        startedAt,
			"completed_at":      completedAt,
			"created_at":        createdAt,
		})
	}

	response.JSON(w, http.StatusOK, jobs)
}
