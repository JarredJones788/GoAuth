package router

import (
	"auth"
	"emailer"
	"encoding/json"
	"fmt"
	"logw"
	"net/http"
	"time"
	"types"
	"utils"

	"github.com/gorilla/mux"
	"golang.org/x/time/rate"
)

//Router type
type Router struct {
	Auth        *auth.Authenticate
	Emailer     *emailer.Emailer
	Log         *logw.Log
	Host        string
	HardLimiter *rate.Limiter
	MedLimiter  *rate.Limiter
}

//Init - inits all routes.
func (router Router) Init(auth *auth.Authenticate, config *types.Config) {

	//Setup Limiters
	router.HardLimiter = rate.NewLimiter(rate.Every(5*time.Minute), 20)
	router.MedLimiter = rate.NewLimiter(rate.Every(1*time.Minute), 30)

	//Setup Helpers
	router.Auth = auth
	router.Emailer = emailer.Emailer{}.Init(config)
	router.Log = logw.Log{}.Init(config)
	router.Host = config.Host

	//Setup mux router
	r := mux.NewRouter()
	router.setUpRoutes(r)
	fmt.Println("Server Started")
	go router.Log.LogEvent(logw.Event{Message: "Server started"})
	http.ListenAndServe(config.ServerPort, r)
}

//setUpRoutes - sets up all endpoints for the service
func (router Router) setUpRoutes(r *mux.Router) {
	r.HandleFunc("/api/auth/login", router.login)
	r.HandleFunc("/api/auth/logout", router.logout)
	r.HandleFunc("/api/auth/checkSession", router.checkSession)
	r.HandleFunc("/api/auth/register", router.registerAccount)
	r.HandleFunc("/api/auth/delete", router.deleteAccount)
	r.HandleFunc("/api/auth/getAllAccounts", router.getAllAccounts)
	r.HandleFunc("/api/auth/getAccounts", router.getAccounts)
	r.HandleFunc("/api/auth/updateSettings", router.updateSettings)
	r.HandleFunc("/api/auth/updateAccountSettings", router.updateAccountSettings)
	r.HandleFunc("/api/auth/activateDevice", router.activateDevice)
	r.HandleFunc("/api/auth/recoverAccount", router.recoverAccount)
	r.HandleFunc("/api/auth/getRecovery", router.getRecovery)
	r.HandleFunc("/api/auth/finishRecovery", router.finishRecovery)
	r.HandleFunc("/api/auth/enableTwoFA", router.enableTwoFA)
	r.HandleFunc("/api/auth/disableTwoFA", router.disableTwoFA)
	r.HandleFunc("/api/auth/changeEmail", router.changeEmail)
	r.HandleFunc("/api/auth/finishEmailChange", router.finishEmailChange)
}

//---------------HELPERS BELOW-------------------\\

//badRequest - returns a generic bad response
func (router Router) badRequest(w http.ResponseWriter) {
	failed, err := json.Marshal(types.GenericResponse{Response: false})
	if err != nil {
		w.Write([]byte("BACKEND ERROR"))
		return
	}
	w.Write(failed)
}

//goodRequest - returns a generic good response
func (router Router) goodRequest(w http.ResponseWriter) {
	good, err := json.Marshal(types.GenericResponse{Response: true})
	if err != nil {
		w.Write([]byte("BACKEND ERROR"))
		return
	}
	w.Write(good)
}

//tooManyRequests - returns too many requests
func (router Router) tooManyRequests(w http.ResponseWriter) {
	router.reasonRequest(w, false, "Too Many Requests!")
}

//reasonRequest - returns a response with a reason
func (router Router) reasonRequest(w http.ResponseWriter, response bool, reason string) {
	good, err := json.Marshal(types.ReasonResponse{Response: response, Reason: reason})
	if err != nil {
		w.Write([]byte("BACKEND ERROR"))
		return
	}
	w.Write(good)
}

//addCookie - adds a cookie to a response
func (router Router) addCookie(w http.ResponseWriter, name string, value string) {
	expire := time.Now().AddDate(1, 0, 0)
	cookie := http.Cookie{
		Name:    name,
		Value:   value,
		Expires: expire,
		Path:    "/",
	}
	http.SetCookie(w, &cookie)
}

//setUpHeaders - sets the desired headers for an http response
func (router Router) setUpHeaders(w http.ResponseWriter, r *http.Request) bool {
	w.Header().Set("Access-Control-Allow-Origin", router.Host)
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Max-Age", "120")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Access-Control-Allow-Headers, Authorization, X-Requested-With")
	if r.Method == http.MethodOptions {
		w.WriteHeader(200)
		return false
	}
	return true
}

//getSessionID - returns sessionId from request cookies
func (router Router) getSessionID(r *http.Request) string {
	for _, cookie := range r.Cookies() {
		if cookie.Name == "sessionId" {
			return cookie.Value
		}
	}
	return ""
}

//getSessionID - returns deviceId from request cookies
func (router Router) getDeviceID(r *http.Request) string {
	for _, cookie := range r.Cookies() {
		if cookie.Name == "deviceId" {
			return cookie.Value
		}
	}
	return ""
}

//getIP - return the ip from the request
func (router Router) getIP(r *http.Request) string {
	IPAddress := r.Header.Get("X-Real-Ip")
	if IPAddress == "" {
		IPAddress = r.Header.Get("X-Forwarded-For")
	}
	if IPAddress == "" {
		IPAddress = r.RemoteAddr
	}
	return IPAddress
}

func (router Router) getSession(r *http.Request) *types.Session {
	return &types.Session{Token: router.getSessionID(r), Device: router.getDeviceID(r)}
}

//---------------ROUTES BELOW-------------------\\

//login - endpoint to login
func (router Router) login(w http.ResponseWriter, r *http.Request) {
	//Medium limiter is set on this request
	if !router.MedLimiter.Allow() {
		router.tooManyRequests(w)
		return
	}

	if !router.setUpHeaders(w, r) {
		return //request was an OPTIONS which was handled.
	}

	var login types.Login
	if err := json.NewDecoder(r.Body).Decode(&login); err != nil {
		go router.Log.LogError(logw.Error{Message: err.Error()})
		router.badRequest(w)
		return
	}

	account, device, err := router.Auth.Login(&login, router.getSession(r))
	if err != nil {
		go router.Log.LogError(logw.Error{Message: err.Error()})
		router.badRequest(w)
		return
	}

	if account != nil {
		router.addCookie(w, "sessionId", account.Token)
		account = account.HideImportant()
		account.GetAccountPermissions()

		if utils.Contains("ADMIN", account.Roles) || account.TwoFA {
			//DEVICE was not found.
			if device == nil {
				router.badRequest(w)
				return
			}
			//Device needs activation.
			if !device.Active {
				//Send New Device Email
				if err = router.Emailer.NewDeviceEmail(account, device); err != nil {
					go router.Log.LogError(logw.Error{Message: err.Error()})
					router.badRequest(w)
					return
				}

				//Email was sent, create response.
				data, err := json.Marshal(types.NewDeviceResponse{Response: true, DeviceID: device.ID, DeviceSetup: true})
				if err != nil {
					go router.Log.LogError(logw.Error{Message: err.Error()})
					router.badRequest(w)
					return
				}

				//Device needs setup. Log and send response to client
				go router.Log.LogEvent(logw.Event{Message: "New device email sent: " + account.Email})
				router.addCookie(w, "deviceId", device.ID)
				w.Write(data)
				return
			}
		}

		//Login is good
		data, err := json.Marshal(types.GoodLoginResponse{Response: true, Account: account, DeviceSetup: false})
		if err == nil {
			w.Write(data)
			return
		}
	}

	router.badRequest(w)
}

//logout - endpoint to logout
func (router Router) logout(w http.ResponseWriter, r *http.Request) {
	if !router.setUpHeaders(w, r) {
		return //request was an OPTIONS which was handled.
	}

	err := router.Auth.Logout(router.getSession(r))
	if err != nil {
		go router.Log.LogError(logw.Error{Message: err.Error()})
		router.badRequest(w)
		return
	}

	router.addCookie(w, "sessionId", "")

	router.goodRequest(w)
}

//checkSession - endpoint to check session
func (router Router) checkSession(w http.ResponseWriter, r *http.Request) {
	if !router.setUpHeaders(w, r) {
		return //request was an OPTIONS which was handled.
	}

	if router.getSessionID(r) == "" {
		router.badRequest(w)
		return
	}

	acc, err := router.Auth.CheckAccountSession(router.getSession(r))
	if err != nil {
		go router.Log.LogError(logw.Error{Message: err.Error()})
		router.badRequest(w)
		return
	}

	res, err := json.Marshal(types.AccountResponse{Response: true, Account: acc.HideImportant()})
	if err != nil {
		go router.Log.LogError(logw.Error{Message: err.Error()})
		router.badRequest(w)
		return
	}
	w.Write(res)
}

//getAllAccounts - endpoint to get all accounts (ADMINS ONLY)
func (router Router) getAllAccounts(w http.ResponseWriter, r *http.Request) {
	if !router.setUpHeaders(w, r) {
		return //request was an OPTIONS which was handled.
	}

	accounts, err := router.Auth.GetAllAccounts(router.getSession(r))
	if err != nil {
		go router.Log.LogError(logw.Error{Message: err.Error()})
		router.badRequest(w)
		return
	}

	data, err := json.Marshal(types.AllUsersResponse{Response: true, Data: accounts})
	if err != nil {
		go router.Log.LogError(logw.Error{Message: err.Error()})
		router.badRequest(w)
		return
	}

	w.Write(data)

}

//getAccounts - endpoint to get all accounts that match the role provided
func (router Router) getAccounts(w http.ResponseWriter, r *http.Request) {
	if !router.setUpHeaders(w, r) {
		return //request was an OPTIONS which was handled.
	}

	var request types.GetAccountsRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		go router.Log.LogError(logw.Error{Message: err.Error()})
		router.badRequest(w)
		return
	}

	accounts, err := router.Auth.GetAccounts(router.getSession(r), request.Roles)
	if err != nil {
		go router.Log.LogError(logw.Error{Message: err.Error()})
		router.badRequest(w)
		return
	}

	data, err := json.Marshal(types.AllUsersResponse{Response: true, Data: accounts})
	if err != nil {
		go router.Log.LogError(logw.Error{Message: err.Error()})
		router.badRequest(w)
		return
	}

	w.Write(data)

}

//registerAccount - endpoint to register a new account
func (router Router) registerAccount(w http.ResponseWriter, r *http.Request) {
	//Hard limiter is set on this request
	// if !router.HardLimiter.Allow() {
	// 	router.tooManyRequests(w)
	// 	return
	// }

	if !router.setUpHeaders(w, r) {
		return //request was an OPTIONS which was handled.
	}
	var account types.Account
	if err := json.NewDecoder(r.Body).Decode(&account); err != nil {
		go router.Log.LogError(logw.Error{Message: err.Error()})
		router.badRequest(w)
		return
	}

	res, err := router.Auth.RegisterAccount(router.getSession(r), &account)
	//Some error occured while trying to create the account
	if err != nil {
		go router.Log.LogError(logw.Error{Message: err.Error()})
		router.badRequest(w)
		return
	}

	//Return a bad response with the reason
	if res != "" {
		router.reasonRequest(w, false, res)
		return
	}

	//Account created
	router.goodRequest(w)
}

//deleteAccount - endpoint to delete an account
func (router Router) deleteAccount(w http.ResponseWriter, r *http.Request) {

	if !router.setUpHeaders(w, r) {
		return //request was an OPTIONS which was handled.
	}
	var del types.DeleteAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&del); err != nil {
		go router.Log.LogError(logw.Error{Message: err.Error()})
		router.badRequest(w)
		return
	}

	res, err := router.Auth.DeleteAccount(&del, router.getSession(r))
	//Some error occured while trying to delete the account
	if err != nil {
		go router.Log.LogError(logw.Error{Message: err.Error()})
		router.badRequest(w)
		return
	}

	//Return a bad response with the reason
	if res != "" {
		router.reasonRequest(w, false, res)
		return
	}

	//Account deleted
	router.goodRequest(w)
}

//updateAccountSettings - endpoint to update another users account settings
func (router Router) updateAccountSettings(w http.ResponseWriter, r *http.Request) {
	if !router.setUpHeaders(w, r) {
		return //request was an OPTIONS which was handled.
	}
	var account types.Account
	if err := json.NewDecoder(r.Body).Decode(&account); err != nil {
		go router.Log.LogError(logw.Error{Message: err.Error()})
		router.badRequest(w)
		return
	}

	res, err := router.Auth.UpdateOtherAccountSettings(&account, router.getSession(r))
	//Some error occured while trying to create the account
	if err != nil {
		go router.Log.LogError(logw.Error{Message: err.Error()})
		router.badRequest(w)
		return
	}

	//Return a bad response with the reason
	if res != "" {
		router.reasonRequest(w, false, res)
		return
	}

	//Account Updated
	router.goodRequest(w)
}

//updateSettings - endpoint to update account settings
func (router Router) updateSettings(w http.ResponseWriter, r *http.Request) {
	if !router.setUpHeaders(w, r) {
		return //request was an OPTIONS which was handled.
	}
	var account types.Account
	if err := json.NewDecoder(r.Body).Decode(&account); err != nil {
		go router.Log.LogError(logw.Error{Message: err.Error()})
		router.badRequest(w)
		return
	}

	res, err := router.Auth.UpdateAccountSettings(&account, router.getSession(r))
	//Some error occured while trying to create the account
	if err != nil {
		go router.Log.LogError(logw.Error{Message: err.Error()})
		router.badRequest(w)
		return
	}

	//Return a bad response with the reason
	if res != "" {
		router.reasonRequest(w, false, res)
		return
	}

	//Account Updated
	router.goodRequest(w)
}

//activateDevice - endpoint to activate a device
func (router Router) activateDevice(w http.ResponseWriter, r *http.Request) {
	//Limit amount of times can be attempted
	if !router.HardLimiter.Allow() {
		router.tooManyRequests(w)
		return
	}

	if !router.setUpHeaders(w, r) {
		return //request was an OPTIONS which was handled.
	}

	var device types.Device
	if err := json.NewDecoder(r.Body).Decode(&device); err != nil {
		go router.Log.LogError(logw.Error{Message: err.Error()})
		router.badRequest(w)
		return
	}

	//Set the device id from the cookie
	device.ID = router.getDeviceID(r)

	//Attempt to activate device with info given
	err := router.Auth.ActivateDevice(router.getSession(r), &device)
	if err != nil {
		go router.Log.LogError(logw.Error{Message: err.Error()})
		router.badRequest(w)
		return
	}

	//Device Activated
	router.goodRequest(w)
}

//recoverAccount - endpoint to recover account by email
func (router Router) recoverAccount(w http.ResponseWriter, r *http.Request) {
	//Hard limiter is set on this request
	if !router.HardLimiter.Allow() {
		router.tooManyRequests(w)
		return
	}

	if !router.setUpHeaders(w, r) {
		return //request was an OPTIONS which was handled.
	}

	var account types.Account
	if err := json.NewDecoder(r.Body).Decode(&account); err != nil {
		go router.Log.LogError(logw.Error{Message: err.Error()})
		router.badRequest(w)
		return
	}

	recovery, err := router.Auth.RecoverAccount(&account)
	if err != nil {
		go router.Log.LogError(logw.Error{Message: err.Error()})
		router.badRequest(w)
		return
	}

	if err = router.Emailer.RecoverAccountEmail(recovery); err != nil {
		go router.Log.LogError(logw.Error{Message: err.Error()})
		router.badRequest(w)
		return
	}

	go router.Log.LogEvent(logw.Event{Message: "Recovery email sent: " + recovery.Email})
	router.goodRequest(w)
}

//getRecover - endpoint to get an account recovery
func (router Router) getRecovery(w http.ResponseWriter, r *http.Request) {
	//Hard limiter is set on this request
	if !router.MedLimiter.Allow() {
		router.tooManyRequests(w)
		return
	}

	if !router.setUpHeaders(w, r) {
		return //request was an OPTIONS which was handled.
	}

	var recovery types.Recovery
	if err := json.NewDecoder(r.Body).Decode(&recovery); err != nil {
		go router.Log.LogError(logw.Error{Message: err.Error()})
		router.badRequest(w)
		return
	}

	rec, err := router.Auth.GetRecovery(&recovery)
	if err != nil {
		go router.Log.LogError(logw.Error{Message: err.Error()})
		router.badRequest(w)
		return
	}

	data, err := json.Marshal(types.RecoveryResponse{Response: true, Data: rec})
	if err != nil {
		go router.Log.LogError(logw.Error{Message: err.Error()})
		router.badRequest(w)
		return
	}

	w.Write(data)
}

//finishRecovery - endpoint to complete an account recovery
func (router Router) finishRecovery(w http.ResponseWriter, r *http.Request) {

	if !router.setUpHeaders(w, r) {
		return //request was an OPTIONS which was handled.
	}

	var recovery types.RecoveryRequest
	if err := json.NewDecoder(r.Body).Decode(&recovery); err != nil {
		go router.Log.LogError(logw.Error{Message: err.Error()})
		router.badRequest(w)
		return
	}

	//Attempt to finish recovery
	res, err := router.Auth.FinishRecovery(&recovery)
	if err != nil {
		go router.Log.LogError(logw.Error{Message: err.Error()})
		router.badRequest(w)
		return
	}

	//Return a bad response with the reason
	if res != "" {
		router.reasonRequest(w, false, res)
		return
	}

	router.goodRequest(w)
}

//enableTwoFA - endpoint to enable TwoFA for an account
func (router Router) enableTwoFA(w http.ResponseWriter, r *http.Request) {
	if !router.setUpHeaders(w, r) {
		return //request was an OPTIONS which was handled.
	}

	err := router.Auth.EnableTwoFA(router.getSession(r))
	if err != nil {
		go router.Log.LogError(logw.Error{Message: err.Error()})
		router.badRequest(w)
		return
	}

	//Log the user out after enabling
	router.logout(w, r)
}

//disableTwoFA - endpoint to disable TwoFA for an account
func (router Router) disableTwoFA(w http.ResponseWriter, r *http.Request) {
	if !router.setUpHeaders(w, r) {
		return //request was an OPTIONS which was handled.
	}

	err := router.Auth.DisableTwoFA(router.getSession(r))
	if err != nil {
		go router.Log.LogError(logw.Error{Message: err.Error()})
		router.badRequest(w)
		return
	}
	//Log the user out after disabling
	router.logout(w, r)
}

//changeEmail - endpoint to change an account email
func (router Router) changeEmail(w http.ResponseWriter, r *http.Request) {
	if !router.setUpHeaders(w, r) {
		return //request was an OPTIONS which was handled.
	}

	var request types.EmailChangeRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		go router.Log.LogError(logw.Error{Message: err.Error()})
		router.badRequest(w)
		return
	}

	res, req, err := router.Auth.ChangeEmail(router.getSession(r), &request)
	if err != nil {
		go router.Log.LogError(logw.Error{Message: err.Error()})
		router.badRequest(w)
		return
	}

	//Return a bad response with the reason
	if res != "" {
		router.reasonRequest(w, false, res)
		return
	}

	if err = router.Emailer.ChangeEmail(req); err != nil {
		go router.Log.LogError(logw.Error{Message: err.Error()})
		router.badRequest(w)
		return
	}

	router.goodRequest(w)
}

//finishEmailChange - completes email change request
func (router Router) finishEmailChange(w http.ResponseWriter, r *http.Request) {
	if !router.setUpHeaders(w, r) {
		return //request was an OPTIONS which was handled.
	}

	var request types.EmailChangeRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		go router.Log.LogError(logw.Error{Message: err.Error()})
		router.badRequest(w)
		return
	}

	err := router.Auth.FinishEmailChange(&request)
	if err != nil {
		go router.Log.LogError(logw.Error{Message: err.Error()})
		router.badRequest(w)
		return
	}

	router.goodRequest(w)
}
