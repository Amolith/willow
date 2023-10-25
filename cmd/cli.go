// SPDX-FileCopyrightText: Amolith <amolith@secluded.site>
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"database/sql"
	"fmt"
	"os"
	"syscall"

	"git.sr.ht/~amolith/willow/users"
	"golang.org/x/term"
)

// createUser is a CLI that creates a new user with the specified username
func createUser(dbConn *sql.DB, username string) {
	fmt.Println("Creating user", username)

	fmt.Print("Enter password: ")
	password, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		fmt.Println("Error reading password:", err)
		os.Exit(1)
	}
	fmt.Println()

	fmt.Print("Confirm password: ")
	passwordConfirmation, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		fmt.Println("Error reading password confirmation:", err)
		os.Exit(1)
	}
	fmt.Println()

	if string(password) != string(passwordConfirmation) {
		fmt.Println("Passwords do not match")
		os.Exit(1)
	}
	err = users.Register(dbConn, username, string(password))
	if err != nil {
		fmt.Println("Error creating user:", err)
		os.Exit(1)
	}

	fmt.Println("\nUser", username, "created successfully")
	os.Exit(0)
}

// deleteUser is a CLI that deletes a user with the specified username
func deleteUser(dbConn *sql.DB, username string) {
	fmt.Println("Deleting user", username)
	err := users.Delete(dbConn, username)
	if err != nil {
		fmt.Println("Error deleting user:", err)
		os.Exit(1)
	}

	fmt.Printf("User %s deleted successfully\n", username)
	os.Exit(0)
}

// listUsers is a CLI that lists all users in the database
func listUsers(dbConn *sql.DB) {
	fmt.Println("Listing all users")

	dbUsers, err := users.GetUsers(dbConn)
	if err != nil {
		fmt.Println("Error retrieving users from the database:", err)
		os.Exit(1)
	}

	if len(dbUsers) == 0 {
		fmt.Println("- No users found")
	} else {
		for _, u := range dbUsers {
			fmt.Println("-", u)
		}
	}
	os.Exit(0)
}

// checkAuthorised is a CLI that checks whether the provided user/password
// combo is authorised.
func checkAuthorised(dbConn *sql.DB, username string) {
	fmt.Printf("Checking whether password for user %s is correct\n", username)

	fmt.Print("Enter password: ")
	password, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		fmt.Println("Error reading password:", err)
		os.Exit(1)
	}
	fmt.Println()

	authorised, err := users.Authorised(dbConn, username, string(password))
	if err != nil {
		fmt.Println("Error checking authorisation:", err)
		os.Exit(1)
	}

	if authorised {
		fmt.Println("User is authorised")
	} else {
		fmt.Println("User is not authorised")
	}
	os.Exit(0)
}
