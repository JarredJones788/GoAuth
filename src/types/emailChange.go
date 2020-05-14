package types

import "time"

//EmailChangeRequest - struct to request email change
type EmailChangeRequest struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

//EmailChange - email change struct
type EmailChange struct {
	ID        string    `sql:"id"`
	AccountID string    `sql:"accountId"`
	NewEmail  string    `sql:"newEmail"`
	OldEmail  string    `sql:"oldEmail"`
	Created   time.Time `sql:"created"`
}
