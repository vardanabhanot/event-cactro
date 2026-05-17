package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"eventbooking/jobs"
	mw "eventbooking/middleware"
	"eventbooking/models"

	"github.com/go-chi/chi/v5"
)

type eventRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Location    string `json:"location"`
	Date        string `json:"date"` // RFC3339 e.g. "2026-12-01T18:00:00Z"
	Capacity    int    `json:"capacity"`
	Price       float64 `json:"price"`
}

// ListEvents handles GET /api/events (public)
func ListEvents(w http.ResponseWriter, r *http.Request) {
	events, err := models.GetEvents()
	if err != nil {
		jsonError(w, "failed to fetch events", http.StatusInternalServerError)
		return
	}
	if events == nil {
		events = []*models.Event{}
	}
	jsonOK(w, map[string]interface{}{"data": events, "total": len(events)}, http.StatusOK)
}

// GetEvent handles GET /api/events/{id} (public)
func GetEvent(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		jsonError(w, "invalid event id", http.StatusBadRequest)
		return
	}

	event, err := models.GetEventByID(id)
	if err != nil {
		jsonError(w, "database error", http.StatusInternalServerError)
		return
	}
	if event == nil {
		jsonError(w, "event not found", http.StatusNotFound)
		return
	}

	jsonOK(w, map[string]interface{}{"data": event}, http.StatusOK)
}

// CreateEvent handles POST /api/events (organizer only)
func CreateEvent(dispatcher *jobs.Dispatcher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := mw.GetClaims(r)
		var req eventRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if err := validateEventRequest(w, &req); err != nil {
			return
		}

		date, err := time.Parse(time.RFC3339, req.Date)
		if err != nil {
			jsonError(w, "invalid date format, use RFC3339 e.g. 2026-12-01T18:00:00Z", http.StatusBadRequest)
			return
		}

		event, err := models.CreateEvent(claims.UserID, strings.TrimSpace(req.Title), req.Description, req.Location, date, req.Capacity, req.Price)
		if err != nil {
			jsonError(w, "failed to create event", http.StatusInternalServerError)
			return
		}

		jsonOK(w, map[string]interface{}{"data": event}, http.StatusCreated)
	}
}

// UpdateEvent handles PUT /api/events/{id} (organizer owner only)
func UpdateEvent(dispatcher *jobs.Dispatcher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := mw.GetClaims(r)
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			jsonError(w, "invalid event id", http.StatusBadRequest)
			return
		}

		event, err := models.GetEventByID(id)
		if err != nil || event == nil {
			jsonError(w, "event not found", http.StatusNotFound)
			return
		}

		// Only the organizer who created the event can update it
		if event.OrganizerID != claims.UserID {
			jsonError(w, "forbidden: you do not own this event", http.StatusForbidden)
			return
		}

		var req eventRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if err := validateEventRequest(w, &req); err != nil {
			return
		}

		date, err := time.Parse(time.RFC3339, req.Date)
		if err != nil {
			jsonError(w, "invalid date format, use RFC3339 e.g. 2026-12-01T18:00:00Z", http.StatusBadRequest)
			return
		}

		// Fetch affected bookings BEFORE updating (for notifications)
		rawBookings, err := models.GetBookingsForEvent(id)
		if err != nil {
			jsonError(w, "database error", http.StatusInternalServerError)
			return
		}

		updated, err := models.UpdateEvent(id, strings.TrimSpace(req.Title), req.Description, req.Location, date, req.Capacity, req.Price)
		if err != nil {
			jsonError(w, "failed to update event", http.StatusInternalServerError)
			return
		}

		// Dispatch BG Task 2: notify all customers with confirmed bookings
		if len(rawBookings) > 0 {
			recipients := make([]jobs.Recipient, len(rawBookings))
			for i, b := range rawBookings {
				recipients[i] = jobs.Recipient{Name: b.CustomerName, Email: b.CustomerEmail}
			}
			dispatcher.Dispatch(jobs.EventNotifyJob{
				EventTitle: updated.Title,
				Recipients: recipients,
			})
		}

		jsonOK(w, map[string]interface{}{"data": updated}, http.StatusOK)
	}
}

// DeleteEvent handles DELETE /api/events/{id} (organizer owner only)
func DeleteEvent(w http.ResponseWriter, r *http.Request) {
	claims := mw.GetClaims(r)
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		jsonError(w, "invalid event id", http.StatusBadRequest)
		return
	}

	event, err := models.GetEventByID(id)
	if err != nil || event == nil {
		jsonError(w, "event not found", http.StatusNotFound)
		return
	}

	if event.OrganizerID != claims.UserID {
		jsonError(w, "forbidden: you do not own this event", http.StatusForbidden)
		return
	}

	if err := models.DeleteEvent(id); err != nil {
		jsonError(w, "failed to delete event", http.StatusInternalServerError)
		return
	}

	jsonOK(w, map[string]interface{}{"message": "event deleted"}, http.StatusOK)
}

// GetEventBookings handles GET /api/events/{id}/bookings (organizer owner only)
func GetEventBookings(w http.ResponseWriter, r *http.Request) {
	claims := mw.GetClaims(r)
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		jsonError(w, "invalid event id", http.StatusBadRequest)
		return
	}

	event, err := models.GetEventByID(id)
	if err != nil || event == nil {
		jsonError(w, "event not found", http.StatusNotFound)
		return
	}

	if event.OrganizerID != claims.UserID {
		jsonError(w, "forbidden: you do not own this event", http.StatusForbidden)
		return
	}

	bookings, err := models.GetBookingsForEvent(id)
	if err != nil {
		jsonError(w, "database error", http.StatusInternalServerError)
		return
	}
	if bookings == nil {
		bookings = []*models.BookingWithCustomer{}
	}

	jsonOK(w, map[string]interface{}{"data": bookings, "total": len(bookings)}, http.StatusOK)
}

// validateEventRequest checks required fields and writes error if invalid.
// Returns non-nil error if validation failed (response already written).
func validateEventRequest(w http.ResponseWriter, req *eventRequest) error {
	if strings.TrimSpace(req.Title) == "" {
		jsonError(w, "title is required", http.StatusBadRequest)
		return http.ErrBodyNotAllowed
	}
	if req.Capacity <= 0 {
		jsonError(w, "capacity must be greater than 0", http.StatusBadRequest)
		return http.ErrBodyNotAllowed
	}
	if req.Date == "" {
		jsonError(w, "date is required", http.StatusBadRequest)
		return http.ErrBodyNotAllowed
	}
	return nil
}
