package types

import "time"

//Device - device struct
type Device struct {
	ID        string    `sql:"id" json:"id"`
	AccountID string    `sql:"accountId" json:"accountId"`
	Created   time.Time `sql:"created" json:"created"`
	Active    bool      `sql:"active" json:"active"`
	Code      string    `sql:"code" json:"code"`
}
