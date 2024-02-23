// SPDX-FileCopyrightText: Amolith <amolith@secluded.site>
//
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"database/sql"
	"time"
)

// DeleteUser deletes specific user from the database and returns an error if it
// fails
func DeleteUser(db *sql.DB, user string) error {
	mutex.Lock()
	defer mutex.Unlock()
	_, err := db.Exec("DELETE FROM users WHERE username = ?", user)
	return err
}

// CreateUser creates a new user in the database and returns an error if it fails
func CreateUser(db *sql.DB, username, hash, salt string) error {
	mutex.Lock()
	defer mutex.Unlock()
	_, err := db.Exec("INSERT INTO users (username, hash, salt) VALUES (?, ?, ?)", username, hash, salt)
	return err
}

// GetUser returns a user's hash and salt from the database as strings and
// returns an error if it fails
func GetUser(db *sql.DB, username string) (string, string, error) {
	var hash, salt string
	err := db.QueryRow("SELECT hash, salt FROM users WHERE username = ?", username).Scan(&hash, &salt)
	return hash, salt, err
}

// GetUsers returns a list of all users in the database as a slice of strings
// and returns an error if it fails
func GetUsers(db *sql.DB) ([]string, error) {
	rows, err := db.Query("SELECT username FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []string
	for rows.Next() {
		var user string
		err = rows.Scan(&user)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

// GetSession accepts a session ID and returns the username associated with it
// and an error
func GetSession(db *sql.DB, session string) (string, time.Time, error) {
	var username string
	var expiresString string
	err := db.QueryRow("SELECT username, expires FROM sessions WHERE token = ?", session).Scan(&username, &expiresString)
	if err != nil {
		return "", time.Time{}, err
	}

	expires, err := time.Parse(time.RFC3339, expiresString)
	if err != nil {
		return "", time.Time{}, err
	}
	return username, expires, nil
}

// InvalidateSession invalidates a session by setting the expiration date to the
// provided time.
func InvalidateSession(db *sql.DB, session string, expiry time.Time) error {
	mutex.Lock()
	defer mutex.Unlock()
	_, err := db.Exec("UPDATE sessions SET expires = ? WHERE token = ?", expiry.Format(time.RFC3339), session)
	return err
}

// CreateSession creates a new session in the database and returns an error if
// it fails
func CreateSession(db *sql.DB, username, token string, expiry time.Time) error {
	mutex.Lock()
	defer mutex.Unlock()
	_, err := db.Exec("INSERT INTO sessions (token, username, expires) VALUES (?, ?, ?)", token, username, expiry.Format(time.RFC3339))
	return err
}
