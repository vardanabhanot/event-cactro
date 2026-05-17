package models

import (
	"database/sql"
	"time"

	"eventbooking/db"
)

type Event struct {
	ID          int64     `json:"id"`
	OrganizerID int64     `json:"organizer_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Location    string    `json:"location"`
	Date        time.Time `json:"date"`
	Capacity    int       `json:"capacity"`
	Price       float64   `json:"price"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateEvent inserts a new event.
func CreateEvent(organizerID int64, title, description, location string, date time.Time, capacity int, price float64) (*Event, error) {
	res, err := db.DB.Exec(
		`INSERT INTO events (organizer_id, title, description, location, date, capacity, price) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		organizerID, title, description, location, date, capacity, price,
	)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return GetEventByID(id)
}

// GetEvents returns all events ordered by date ascending.
func GetEvents() ([]*Event, error) {
	rows, err := db.DB.Query(
		`SELECT id, organizer_id, title, description, location, date, capacity, price, created_at, updated_at FROM events ORDER BY date ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*Event
	for rows.Next() {
		e := &Event{}
		if err := rows.Scan(&e.ID, &e.OrganizerID, &e.Title, &e.Description, &e.Location, &e.Date, &e.Capacity, &e.Price, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, nil
}

// GetEventByID fetches a single event by ID.
func GetEventByID(id int64) (*Event, error) {
	e := &Event{}
	err := db.DB.QueryRow(
		`SELECT id, organizer_id, title, description, location, date, capacity, price, created_at, updated_at FROM events WHERE id = ?`,
		id,
	).Scan(&e.ID, &e.OrganizerID, &e.Title, &e.Description, &e.Location, &e.Date, &e.Capacity, &e.Price, &e.CreatedAt, &e.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return e, err
}

// UpdateEvent updates event fields and bumps updated_at.
func UpdateEvent(id int64, title, description, location string, date time.Time, capacity int, price float64) (*Event, error) {
	_, err := db.DB.Exec(
		`UPDATE events SET title=?, description=?, location=?, date=?, capacity=?, price=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`,
		title, description, location, date, capacity, price, id,
	)
	if err != nil {
		return nil, err
	}
	return GetEventByID(id)
}

// DeleteEvent removes an event by ID.
func DeleteEvent(id int64) error {
	_, err := db.DB.Exec(`DELETE FROM events WHERE id = ?`, id)
	return err
}

// GetBookedTickets returns the total confirmed tickets booked for an event.
func GetBookedTickets(eventID int64) (int, error) {
	var total int
	err := db.DB.QueryRow(
		`SELECT COALESCE(SUM(tickets), 0) FROM bookings WHERE event_id = ? AND status = 'confirmed'`,
		eventID,
	).Scan(&total)
	return total, err
}
