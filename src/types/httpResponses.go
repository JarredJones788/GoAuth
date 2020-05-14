package types

//GenericResponse - Simple response
type GenericResponse struct {
	Response bool `json:"response"`
}

//AccountResponse - Account response
type AccountResponse struct {
	Response bool     `json:"response"`
	Account  *Account `json:"account"`
}

//GoodLoginResponse - return success with data
type GoodLoginResponse struct {
	Response    bool     `json:"response"`
	Account     *Account `json:"account"`
	DeviceSetup bool     `json:"deviceSetup"`
}

//NewDeviceResponse - new device data.
type NewDeviceResponse struct {
	Response    bool   `json:"response"`
	DeviceID    string `json:"deviceId"`
	DeviceSetup bool   `json:"deviceSetup"`
}

//AllUsersResponse - return success with data
type AllUsersResponse struct {
	Response bool       `json:"response"`
	Data     *[]Account `json:"data"`
}

//RecoveryResponse - return success with data
type RecoveryResponse struct {
	Response bool      `json:"response"`
	Data     *Recovery `json:"data"`
}

//ReasonResponse - return response with a reason
type ReasonResponse struct {
	Response bool   `json:"response"`
	Reason   string `json:"reason"`
}
