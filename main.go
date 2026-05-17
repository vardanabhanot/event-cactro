package main

import (
	"fmt"
	"log"
	"net/http"

	"eventbooking/config"
	"eventbooking/db"
	"eventbooking/handlers"
	"eventbooking/jobs"
	"eventbooking/middleware"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
)

func main() {
	// Load config from .env
	config.Load()

	// Connect to SQLite and run migrations
	db.Connect()
	db.Migrate()

	// Start the background job dispatcher (buffer of 100 jobs)
	dispatcher := jobs.NewDispatcher(100)

	// Set up router
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.RealIP)
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			next.ServeHTTP(w, r)
		})
	})

	// API reference UI
	r.Get("/help", handlers.Help)

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"ok"}`))
	})

	// API routes
	r.Route("/api", func(r chi.Router) {

		// --- Public Auth Routes ---
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", handlers.Register)
			r.Post("/login", handlers.Login)
		})

		// --- Public Event Browsing ---
		r.Get("/events", handlers.ListEvents)
		r.Get("/events/{id}", handlers.GetEvent)

		// --- Authenticated Routes ---
		r.Group(func(r chi.Router) {
			r.Use(middleware.Authenticate)

			// Organizer-only event management
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole("organizer"))
				r.Post("/events", handlers.CreateEvent(dispatcher))
				r.Put("/events/{id}", handlers.UpdateEvent(dispatcher))
				r.Delete("/events/{id}", handlers.DeleteEvent)
				r.Get("/events/{id}/bookings", handlers.GetEventBookings)
			})

			// Customer-only booking management
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole("customer"))
				r.Post("/events/{id}/book", handlers.BookEvent(dispatcher))
				r.Get("/bookings", handlers.MyBookings)
				r.Delete("/bookings/{id}", handlers.CancelBooking)
			})
		})
	})

	addr := fmt.Sprintf(":%s", config.App.Port)
	log.Printf("🚀 Event Booking API running on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
