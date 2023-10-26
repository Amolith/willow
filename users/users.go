// SPDX-FileCopyrightText: Amolith <amolith@secluded.site>
//
// SPDX-License-Identifier: Apache-2.0

package users

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"time"

	"git.sr.ht/~amolith/willow/db"
	"golang.org/x/crypto/argon2"
)

// argonHash accepts two strings for the user's password and a random salt,
// hashes the password using the salt, and returns the hash as a base64-encoded
// string.
func argonHash(password, salt string) (string, error) {
	decodedSalt, err := base64.StdEncoding.DecodeString(salt)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(argon2.IDKey([]byte(password), decodedSalt, 2, 64*1024, 4, 64)), nil
}

// generateSalt generates a random salt and returns it as a base64-encoded
// string.
func generateSalt() (string, error) {
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(salt), nil
}

// Register accepts a username and password, hashes the password and stores the
// hash and salt in the database.
func Register(dbConn *sql.DB, username, password string) error {
	salt, err := generateSalt()
	if err != nil {
		return err
	}

	hash, err := argonHash(password, salt)
	if err != nil {
		return err
	}

	return db.CreateUser(dbConn, username, hash, salt)
}

// Delete removes a user from the database.
func Delete(dbConn *sql.DB, username string) error { return db.DeleteUser(dbConn, username) }

// UserAuthorised accepts a username string, a token string, and returns true if the
// user is authorised, false if not, and an error if one is encountered.
func UserAuthorised(dbConn *sql.DB, username, token string) (bool, error) {
	dbHash, dbSalt, err := db.GetUser(dbConn, username)
	if err != nil {
		return false, err
	}

	providedHash, err := argonHash(token, dbSalt)
	if err != nil {
		return false, err
	}

	return dbHash == providedHash, nil
}

// SessionAuthorised accepts a session string and returns true if the session is
// valid and false if not.
func SessionAuthorised(dbConn *sql.DB, session string) (bool, error) {
	dbResult, expiry, err := db.GetSession(dbConn, session)
	if dbResult == "" || expiry.Before(time.Now()) || err != nil {
		return false, err
	}

	return true, nil
}

// InvalidateSession invalidates a session by setting the expiration date to now.
func InvalidateSession(dbConn *sql.DB, session string) error {
	return db.InvalidateSession(dbConn, session, time.Now())
}

// CreateSession accepts a username, generates a token, stores it in the
// database, and returns it
func CreateSession(dbConn *sql.DB, username string) (string, time.Time, error) {
	token, err := generateSalt()
	if err != nil {
		return "", time.Time{}, err
	}

	expiry := time.Now().Add(7 * 24 * time.Hour)

	err = db.CreateSession(dbConn, username, token, expiry)
	if err != nil {
		return "", time.Time{}, err
	}

	return token, expiry, nil
}

// GetUsers returns a list of all users in the database as a slice of strings.
func GetUsers(dbConn *sql.DB) ([]string, error) { return db.GetUsers(dbConn) }
