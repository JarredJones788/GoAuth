package types

import "time"

//Recovery - struct for recovery class
type Recovery struct {
	ID        string    `sql:"id"`
	AccountID string    `sql:"accountId"`
	Email     string    `sql:"email"`
	UserName  string    `sql:"userName"`
	Created   time.Time `sql:"created"`
}

//RecoveryRequest - struct for finishing a recovery
type RecoveryRequest struct {
	ID       string `sql:"id"`
	Password string `sql:"password"`
}
