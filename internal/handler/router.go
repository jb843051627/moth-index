package handler

import (
	"net/http"

	"github.com/jb843051627/moth-index/internal/engine"
	"github.com/jb843051627/moth-index/internal/service"
	"github.com/jb843051627/moth-index/internal/web"
)

type Handler struct {
	services *service.Registry
	queue    *engine.Queue
}

func NewRouter(services *service.Registry, queue *engine.Queue) http.Handler {
	h := &Handler{services: services, queue: queue}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", h.home)
	mux.HandleFunc("GET /healthz", h.health)
	mux.HandleFunc("GET /api/stations", h.listStations)
	mux.HandleFunc("POST /api/stations", h.createStation)
	mux.HandleFunc("GET /api/stations/{id}", h.getStation)
	mux.HandleFunc("POST /api/stations/{id}/retire", h.retireStation)
	mux.HandleFunc("POST /api/stations/{id}/reactivate", h.reactivateStation)
	mux.HandleFunc("GET /api/stations/{id}/traps", h.listTraps)
	mux.HandleFunc("POST /api/stations/{id}/traps", h.createTrap)
	mux.HandleFunc("POST /api/traps/{id}/deploy", h.deployTrap)
	mux.HandleFunc("POST /api/traps/{id}/release", h.releaseTrap)
	mux.HandleFunc("POST /api/traps/{id}/break", h.breakTrap)
	mux.HandleFunc("GET /api/batches", h.listBatches)
	mux.HandleFunc("POST /api/batches", h.openBatch)
	mux.HandleFunc("GET /api/batches/{id}", h.getBatch)
	mux.HandleFunc("POST /api/batches/{id}/review", h.reviewBatch)
	mux.HandleFunc("POST /api/batches/{id}/close", h.closeBatch)
	mux.HandleFunc("POST /api/batches/{id}/archive", h.archiveBatch)
	mux.HandleFunc("GET /api/batches/{id}/specimens", h.listSpecimens)
	mux.HandleFunc("POST /api/batches/{id}/specimens", h.captureSpecimen)
	mux.HandleFunc("GET /api/batches/{id}/readings", h.listReadings)
	mux.HandleFunc("POST /api/batches/{id}/readings", h.ingestReading)
	mux.HandleFunc("GET /api/specimens/{id}", h.getSpecimen)
	mux.HandleFunc("POST /api/specimens/{id}/classify", h.classifySpecimen)
	mux.HandleFunc("POST /api/specimens/{id}/approve", h.approveSpecimen)
	mux.HandleFunc("POST /api/specimens/{id}/reject", h.rejectSpecimen)
	mux.HandleFunc("GET /api/taxa", h.searchTaxa)
	mux.HandleFunc("POST /api/taxa", h.registerTaxon)
	mux.HandleFunc("GET /api/batches/{id}/summary", h.batchSummary)
	mux.HandleFunc("POST /api/batches/{id}/evaluate", h.evaluateBatch)
	mux.HandleFunc("GET /api/batches/{id}/analysis", h.analyzeBatch)
	mux.HandleFunc("GET /api/batches/{id}/findings", h.batchFindings)
	mux.HandleFunc("GET /api/batches/{id}/phenology", h.phenologyReport)
	mux.HandleFunc("GET /api/batches/{id}/season-score", h.seasonScore)
	mux.HandleFunc("GET /api/reports/daily", h.dailyReport)
	mux.HandleFunc("GET /api/reports/daily.csv", h.dailyCSV)
	mux.HandleFunc("GET /api/batches/{id}/timeline", h.timeline)
	mux.HandleFunc("GET /api/maintenance/events", h.maintenanceEvents)
	mux.HandleFunc("GET /api/queue/metrics", h.queueMetrics)
	return withHeaders(mux)
}

func withHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		next.ServeHTTP(w, r)
	})
}

func (h *Handler) home(w http.ResponseWriter, r *http.Request) { web.Page().ServeHTTP(w, r) }
