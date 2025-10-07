package core

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

func NewIdentity(db *sql.DB) string {
	id := uuid.NewString()
	_, err := db.Exec("INSERT INTO identities (id) VALUES (?)", id)
	if err != nil {
		fmt.Println("Error creating identity:", err)
	}
	return id
}
