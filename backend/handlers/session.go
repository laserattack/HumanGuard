package handlers

import (
	"encoding/json"
	"errors"
	"humanguard/storage"
	"net/http"
	"strconv"
)

type SessionHandler struct {
	storage storage.Storage
}

func NewSessionHandler(store storage.Storage) *SessionHandler {
	return &SessionHandler{storage: store}
}

// POST /api/sessions
func (h *SessionHandler) CreateSession(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SiteID    string `json:"site_id"`
		IP        string `json:"ip"`
		UserAgent string `json:"user_agent"`
		Device    string `json:"device"`
		Location  string `json:"location"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.SiteID == "" || req.IP == "" {
		http.Error(w, "site_id and ip are required", http.StatusBadRequest)
		return
	}

	session := &storage.Session{
		SiteID:    &req.SiteID,
		IP:        req.IP,
		UserAgent: req.UserAgent,
		Device:    req.Device,
		Location:  req.Location,
		IsActive:  true,
		RiskScore: 0,
	}

	if err := h.storage.CreateSession(r.Context(), session); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(session)
}

// GET /api/sessions/{id}
func (h *SessionHandler) GetSession(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	session, err := h.storage.GetSession(r.Context(), id)
	if err != nil {
		if errors.Is(err, storage.ErrSessionNotFound) {
			http.Error(w, "Session not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
}

// DELETE /api/sessions/{id}
func (h *SessionHandler) DeactivateSession(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if err := h.storage.DeactivateSession(r.Context(), id); err != nil {
		if errors.Is(err, storage.ErrSessionNotFound) {
			http.Error(w, "Session not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// POST /api/sessions/{id}/block
func (h *SessionHandler) BlockSession(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if err := h.storage.BlockSession(r.Context(), id); err != nil {
		if errors.Is(err, storage.ErrSessionNotFound) {
			http.Error(w, "Session not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// POST /api/sessions/{id}/unblock
func (h *SessionHandler) UnblockSession(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if err := h.storage.UnblockSession(r.Context(), id); err != nil {
		if errors.Is(err, storage.ErrSessionNotFound) {
			http.Error(w, "Session not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// PATCH /api/sessions/{id}/risk
func (h *SessionHandler) UpdateRiskScore(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var req struct {
		RiskScore int `json:"risk_score"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.RiskScore < 0 || req.RiskScore > 100 {
		http.Error(w, "risk_score must be between 0 and 100", http.StatusBadRequest)
		return
	}

	if err := h.storage.UpdateRiskScore(r.Context(), id, req.RiskScore); err != nil {
		if errors.Is(err, storage.ErrSessionNotFound) {
			http.Error(w, "Session not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GET /api/sites/{id}/sessions
func (h *SessionHandler) GetSessionsBySite(w http.ResponseWriter, r *http.Request) {
	siteID := r.PathValue("id")
	limit := 100

	if limitParam := r.URL.Query().Get("limit"); limitParam != "" {
		if l, err := strconv.Atoi(limitParam); err == nil && l > 0 {
			limit = l
		}
	}

	sessions, err := h.storage.GetActiveSessionsBySite(r.Context(), siteID, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if sessions == nil {
		sessions = []*storage.Session{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sessions)
}

// GET /api/sites/{id}/sessions/suspicious
func (h *SessionHandler) GetSuspiciousSessions(w http.ResponseWriter, r *http.Request) {
	siteID := r.PathValue("id")
	minRisk := 60

	if minRiskParam := r.URL.Query().Get("min_risk"); minRiskParam != "" {
		if mr, err := strconv.Atoi(minRiskParam); err == nil && mr >= 0 && mr <= 100 {
			minRisk = mr
		}
	}

	sessions, err := h.storage.GetSuspiciousSessions(r.Context(), siteID, minRisk)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if sessions == nil {
		sessions = []*storage.Session{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sessions)
}

// GET /api/sites/{id}/stats
func (h *SessionHandler) GetSessionStats(w http.ResponseWriter, r *http.Request) {
	siteID := r.PathValue("id")

	stats, err := h.storage.GetSessionStats(r.Context(), siteID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// GET /api/sessions/cookie/{cookie}
func (h *SessionHandler) GetSessionByCookie(w http.ResponseWriter, r *http.Request) {
	cookie := r.PathValue("cookie")

	session, err := h.storage.GetSessionByCookie(r.Context(), cookie)
	if err != nil {
		if errors.Is(err, storage.ErrSessionNotFound) {
			http.Error(w, "Session not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
}

// PUT /api/sessions/{id}
func (h *SessionHandler) UpdateSession(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var req storage.Session
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	req.ID = id
	if err := h.storage.UpdateSession(r.Context(), &req); err != nil {
		if errors.Is(err, storage.ErrSessionNotFound) {
			http.Error(w, "Session not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// POST /api/sessions/{id}/activity
func (h *SessionHandler) UpdateSessionActivity(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if err := h.storage.UpdateSessionActivity(r.Context(), id); err != nil {
		if errors.Is(err, storage.ErrSessionNotFound) {
			http.Error(w, "Session not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// POST /api/sessions/{id}/captcha
func (h *SessionHandler) MarkCaptchaShown(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if err := h.storage.MarkCaptchaShown(r.Context(), id); err != nil {
		if errors.Is(err, storage.ErrSessionNotFound) {
			http.Error(w, "Session not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// POST /api/sessions/cleanup
func (h *SessionHandler) CleanupExpiredSessions(w http.ResponseWriter, r *http.Request) {
	count, err := h.storage.CleanupExpiredSessions(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int64{"deleted": count})
}
