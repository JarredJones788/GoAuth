package types

//GetAccountsRequest - type of account wanted
type GetAccountsRequest struct {
	Role int `json:"role"`
}

//DeleteAccountRequest - Id of the account being deletes
type DeleteAccountRequest struct {
	ID string `json:"id"`
}
