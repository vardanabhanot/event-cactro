package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"eventbooking/jobs"
	mw "eventbooking/middleware"
	"eventbooking/models"

	"github.com/go-chi/chi/v5"
)

type bookEventRequest struct {
	Tickets int `json:"tickets"`
}

// BookEvent handles POST /api/events/{id}/book (customer only)
func BookEvent(dispatcher *jobs.Dispatcher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := mw.GetClaims(r)

		eventID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			jsonError(w, "invalid event id", http.StatusBadRequest)
			return
		}

		var req bookEventRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if req.Tickets <= 0 {
			jsonError(w, "tickets must be greater than 0", http.StatusBadRequest)
			return
		}

		// Fetch event
		event, err := models.GetEventByID(eventID)
		if err != nil {
			jsonError(w, "database error", http.StatusInternalServerError)
			return
		}
		if event == nil {
			jsonError(w, "event not found", http.StatusNotFound)
			return
		}

		// Check available capacity
		booked, err := models.GetBookedTickets(eventID)
		if err != nil {
			jsonError(w, "database error", http.StatusInternalServerError)
			return
		}

		remaining := event.Capacity - booked
		if req.Tickets > remaining {
			jsonError(w, map[string]interface{}{
				"message":   "not enough capacity",
				"requested": req.Tickets,
				"remaining": remaining,
			}, http.StatusConflict)
			return
		}

		// Create the booking
		booking, err := models.CreateBooking(claims.UserID, eventID, req.Tickets)
		if err != nil {
			jsonError(w, "failed to create booking", http.StatusInternalServerError)
			return
		}

		// Fetch customer info for the job
		customer, err := models.FindUserByID(claims.UserID)
		if err != nil || customer == nil {
			jsonError(w, "database error", http.StatusInternalServerError)
			return
		}

		// Dispatch BG Task 1: booking confirmation
		dispatcher.Dispatch(jobs.BookingConfirmJob{
			CustomerName:  customer.Name,
			CustomerEmail: customer.Email,
			EventTitle:    event.Title,
			EventDate:     event.Date,
			Tickets:       booking.Tickets,
		})

		jsonOK(w, map[string]interface{}{"data": booking}, http.StatusCreated)
	}
}

// MyBookings handles GET /api/bookings (customer only)
func MyBookings(w http.ResponseWriter, r *http.Request) {
	claims := mw.GetClaims(r)

	bookings, err := models.GetCustomerBookings(claims.UserID)
	if err != nil {
		jsonError(w, "database error", http.StatusInternalServerError)
		return
	}
	if bookings == nil {
		bookings = []*models.BookingWithEvent{}
	}

	jsonOK(w, map[string]interface{}{"data": bookings, "total": len(bookings)}, http.StatusOK)
}

// CancelBooking handles DELETE /api/bookings/{id} (customer only, must own booking)
func CancelBooking(w http.ResponseWriter, r *http.Request) {
	claims := mw.GetClaims(r)

	bookingID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		jsonError(w, "invalid booking id", http.StatusBadRequest)
		return
	}

	booking, err := models.GetBookingByID(bookingID)
	if err != nil {
		jsonError(w, "database error", http.StatusInternalServerError)
		return
	}
	if booking == nil {
		jsonError(w, "booking not found", http.StatusNotFound)
		return
	}

	// Only the customer who made the booking can cancel it
	if booking.CustomerID != claims.UserID {
		jsonError(w, "forbidden: you do not own this booking", http.StatusForbidden)
		return
	}

	if booking.Status == "cancelled" {
		jsonError(w, "booking is already cancelled", http.StatusBadRequest)
		return
	}

	if err := models.CancelBooking(bookingID); err != nil {
		jsonError(w, "failed to cancel booking", http.StatusInternalServerError)
		return
	}

	jsonOK(w, map[string]interface{}{"message": "booking cancelled"}, http.StatusOK)
}
