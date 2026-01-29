package net

import (
	"encoding/json"
	"net/http"
	"strconv"

	"reservechain/internal/store"
)

// GET /api/slashing/events?epoch=N&subject_type=pop_node&subject_id=node-1&status=pending&limit=200
func (api *HTTPAPI) slashingEventsHandler(w http.ResponseWriter, r *http.Request) {
	if api == nil || api.Store == nil {
		http.Error(w, "store unavailable", http.StatusServiceUnavailable)
		return
	}
	q := r.URL.Query()
	var epochPtr *int64
	if es := q.Get("epoch"); es != "" {
		e, err := strconv.ParseInt(es, 10, 64)
		if err != nil {
			http.Error(w, "invalid epoch", http.StatusBadRequest)
			return
		}
		epochPtr = &e
	}
	subjectType := q.Get("subject_type")
	subjectID := q.Get("subject_id")
	status := q.Get("status")
	limit := 200
	if ls := q.Get("limit"); ls != "" {
		if v, err := strconv.Atoi(ls); err == nil {
			limit = v
		}
	}

	events, err := api.Store.ListSlashingEvents(r.Context(), epochPtr, subjectType, subjectID, status, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(struct {
		Events []store.SlashingEvent `json:"events"`
	}{Events: events})
}
