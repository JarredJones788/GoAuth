package types

//GetAccountsRequest - type of account wanted
type GetAccountsRequest struct {
	Roles []int `json:"roles"`
}

//DeleteAccountRequest - Id of the account being deletes
type DeleteAccountRequest struct {
	ID string `json:"id"`
}
