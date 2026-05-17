package models

import (
	"database/sql"
	"time"

	"eventbooking/db"
)

type User struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // never serialized
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateUser inserts a new user and returns the created record.
func CreateUser(name, email, hashedPassword, role string) (*User, error) {
	res, err := db.DB.Exec(
		`INSERT INTO users (name, email, password, role) VALUES (?, ?, ?, ?)`,
		name, email, hashedPassword, role,
	)
	if err != nil {
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	return FindUserByID(id)
}

// FindUserByEmail fetches a user by email (used during login).
func FindUserByEmail(email string) (*User, error) {
	u := &User{}
	err := db.DB.QueryRow(
		`SELECT id, name, email, password, role, created_at FROM users WHERE email = ?`,
		email,
	).Scan(&u.ID, &u.Name, &u.Email, &u.Password, &u.Role, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return u, err
}

// FindUserByID fetches a user by primary key.
func FindUserByID(id int64) (*User, error) {
	u := &User{}
	err := db.DB.QueryRow(
		`SELECT id, name, email, password, role, created_at FROM users WHERE id = ?`,
		id,
	).Scan(&u.ID, &u.Name, &u.Email, &u.Password, &u.Role, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return u, err
}
