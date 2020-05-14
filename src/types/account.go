package types

import (
	"errors"
	"regexp"
	"time"
)

//Account - struct for account class
type Account struct {
	ID       string    `sql:"id" json:"id"`
	UserName string    `sql:"userName" json:"userName"`
	Password string    `sql:"password" json:"password"`
	Name     string    `sql:"name" json:"name"`
	Phone    string    `sql:"phone" json:"phone"`
	Email    string    `sql:"email" json:"email"`
	Role     int       `sql:"role" json:"role"`
	Token    string    `sql:"token" json:"token"`
	TwoFA    bool      `sql:"twoFA" json:"twoFA"`
	Roles    []string  `json:"roles"`
	Created  time.Time `sql:"created" json:"created"`
}

//CheckUserName - verify username is valid.
func (account Account) CheckUserName() error {
	if account.UserName == "" || len(account.UserName) < 6 {
		return errors.New("Invalid UserName: " + account.UserName)
	}
	if regexp.MustCompile(`\s`).MatchString(account.UserName) {
		return errors.New("UserName cannot have spaces: " + account.UserName)
	}
	return nil
}

//CheckPassword - verify password is valid.
func (account Account) CheckPassword() error {
	if len(account.Password) < 7 {
		return errors.New("Password must be 7 or more characters")
	}
	if !regexp.MustCompile(`\d`).MatchString(account.Password) {
		return errors.New("Password must contain a number")
	}
	if !regexp.MustCompile(`.*[a-zA-Z].*`).MatchString(account.Password) {
		return errors.New("Password must contain a letter")
	}
	return nil
}

//CheckEmail - verify email is valid.
func (account Account) CheckEmail() error {
	if !regexp.MustCompile(`^(([^<>()\[\]\\.,;:\s@"]+(\.[^<>()\[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$`).MatchString(account.Email) {
		return errors.New("Invalid email address: " + account.Email)
	}
	return nil
}

//CheckPhone - verify phone is valid.
func (account Account) CheckPhone() error {
	if !regexp.MustCompile(`^[\+]?[(]?[0-9]{3}[)]?[-\s\.]?[0-9]{3}[-\s\.]?[0-9]{4,6}$`).MatchString(account.Phone) {
		return errors.New("Invalid phone number: " + account.Phone)
	}
	return nil
}

//HideImportant - Hides sensative info on the account
func (account Account) HideImportant() *Account {
	account.Password = ""
	account.Token = ""
	return &account
}

//HideInfo - Hides sensative info on the account returns non pointer
func (account Account) HideInfo() Account {
	account.Password = ""
	account.Token = ""
	return account
}

//----------!!! UPDATE BOTH FUNCTIONS AT THE SAME TIME !!!------------\\

//GetAccountPermissions - return the accounts permission. Return a pointer
func (account Account) GetAccountPermissions() *Account {
	account.Roles = GetRoles(account.Role)
	return &account
}

//GetPermissions - return the accounts permission. Returns a value
func (account Account) GetPermissions() Account {
	account.Roles = GetRoles(account.Role)
	return account
}
