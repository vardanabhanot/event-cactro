package models

import (
	"database/sql"
	"time"

	"eventbooking/db"
)

type Booking struct {
	ID         int64     `json:"id"`
	CustomerID int64     `json:"customer_id"`
	EventID    int64     `json:"event_id"`
	Tickets    int       `json:"tickets"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}

// BookingWithCustomer extends Booking with customer info for organizer views.
type BookingWithCustomer struct {
	Booking
	CustomerName  string `json:"customer_name"`
	CustomerEmail string `json:"customer_email"`
}

// BookingWithEvent extends Booking with event info for customer views.
type BookingWithEvent struct {
	Booking
	EventTitle string    `json:"event_title"`
	EventDate  time.Time `json:"event_date"`
}

// CreateBooking inserts a new booking record.
func CreateBooking(customerID, eventID int64, tickets int) (*Booking, error) {
	res, err := db.DB.Exec(
		`INSERT INTO bookings (customer_id, event_id, tickets) VALUES (?, ?, ?)`,
		customerID, eventID, tickets,
	)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return GetBookingByID(id)
}

// GetBookingByID fetches a single booking.
func GetBookingByID(id int64) (*Booking, error) {
	b := &Booking{}
	err := db.DB.QueryRow(
		`SELECT id, customer_id, event_id, tickets, status, created_at FROM bookings WHERE id = ?`, id,
	).Scan(&b.ID, &b.CustomerID, &b.EventID, &b.Tickets, &b.Status, &b.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return b, err
}

// GetCustomerBookings returns all bookings for a customer with event info.
func GetCustomerBookings(customerID int64) ([]*BookingWithEvent, error) {
	rows, err := db.DB.Query(`
		SELECT b.id, b.customer_id, b.event_id, b.tickets, b.status, b.created_at,
		       e.title, e.date
		FROM bookings b
		JOIN events e ON e.id = b.event_id
		WHERE b.customer_id = ?
		ORDER BY b.created_at DESC
	`, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookings []*BookingWithEvent
	for rows.Next() {
		bwe := &BookingWithEvent{}
		if err := rows.Scan(
			&bwe.ID, &bwe.CustomerID, &bwe.EventID, &bwe.Tickets, &bwe.Status, &bwe.CreatedAt,
			&bwe.EventTitle, &bwe.EventDate,
		); err != nil {
			return nil, err
		}
		bookings = append(bookings, bwe)
	}
	return bookings, nil
}

// GetBookingsForEvent returns all confirmed bookings for an event with customer info.
func GetBookingsForEvent(eventID int64) ([]*BookingWithCustomer, error) {
	rows, err := db.DB.Query(`
		SELECT b.id, b.customer_id, b.event_id, b.tickets, b.status, b.created_at,
		       u.name, u.email
		FROM bookings b
		JOIN users u ON u.id = b.customer_id
		WHERE b.event_id = ? AND b.status = 'confirmed'
		ORDER BY b.created_at ASC
	`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookings []*BookingWithCustomer
	for rows.Next() {
		bwc := &BookingWithCustomer{}
		if err := rows.Scan(
			&bwc.ID, &bwc.CustomerID, &bwc.EventID, &bwc.Tickets, &bwc.Status, &bwc.CreatedAt,
			&bwc.CustomerName, &bwc.CustomerEmail,
		); err != nil {
			return nil, err
		}
		bookings = append(bookings, bwc)
	}
	return bookings, nil
}

// CancelBooking sets a booking's status to cancelled.
func CancelBooking(id int64) error {
	_, err := db.DB.Exec(`UPDATE bookings SET status = 'cancelled' WHERE id = ?`, id)
	return err
}
