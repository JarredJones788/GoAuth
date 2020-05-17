package auth

import (
	"cache"
	"db"
	"errors"
	"manager"
	"types"
	"utils"
)

//Authenticate - Authenticate class
type Authenticate struct {
	DB    *db.MySQL
	Cache *cache.Cache
}

//Init - Start authentication service
func (auth Authenticate) Init(db *db.MySQL, config *types.Config) *Authenticate {
	auth.DB = db
	auth.Cache = cache.Cache{}.Init(config)
	return &auth
}

//getAccountSession - gets an account from a session **DOES NOT CHECK IF DEVICE IS VALID OR NOT
//NOT SECURE FOR VALIDATING SESSION!!! USE CheckAccountSession() INSTEAD!
func (auth Authenticate) getAccountSession(session *types.Session) (*types.Account, error) {

	//Invalid sessionId
	if session.Token == "" {
		return nil, errors.New("Invalid Session Id")
	}

	account, err := manager.AccountManager{}.GetAccountSession(session, auth.DB, auth.Cache)
	if err != nil {
		return nil, err
	}

	return account, nil
}

//Login - Checks if login is valid
func (auth Authenticate) Login(login *types.Login, session *types.Session) (*types.Account, *types.Device, error) {
	am := manager.AccountManager{}
	dm := manager.DeviceManager{}

	//Get account by username or email provided
	account, err := am.GetAccountLoginDetails(login.UserName, auth.DB)

	if err != nil {
		return nil, nil, err
	}

	//Check if password matches hash
	valid := utils.CheckPasswordHash(login.Password, account.Password)
	if !valid {
		return nil, nil, errors.New("Invalid Password Attempt: " + account.Name)
	}

	oldToken := account.Token

	//Set a new session token
	account.Token = utils.RandomString()

	//Get Account Roles
	account = account.GetAccountPermissions()

	//If account is ADMIN or above or 2FA is enabled then make sure device is verified.
	if utils.Contains("ADMIN", account.Roles) || account.TwoFA {
		device, err := dm.GetDevice(session, auth.DB, auth.Cache)
		if err != nil {
			return nil, nil, err
		}

		//No device was found, need to create one.
		if err == nil && device == nil {
			device, err = dm.CreateDevice(account, auth.DB)
			if err != nil {
				return nil, nil, err
			}
		}

		//If the device does not belong to the account.
		//Create a new one for the account.
		if account.ID != device.AccountID {
			device, err = dm.CreateDevice(account, auth.DB)
			if err != nil {
				return nil, nil, err
			}
		}

		//Save updated session to Database and Cache (If cache is enabled)
		err = am.UpdateAccountToken(account, auth.DB)
		if err != nil {
			return nil, nil, err
		}
		auth.Cache.Del(oldToken)
		am.SaveToCache(account, auth.Cache)
		dm.SaveToCache(device, auth.Cache)
		return account, device, nil
	}

	//Save updated session to Database and Cache (If cache is enabled)
	err = am.UpdateAccountToken(account, auth.DB)
	if err != nil {
		return nil, nil, err
	}
	auth.Cache.Del(oldToken)
	am.SaveToCache(account, auth.Cache)
	return account, nil, nil
}

//Logout - removes users session from system
func (auth Authenticate) Logout(session *types.Session) error {
	err := manager.AccountManager{}.RemoveSession(session, auth.Cache)
	if err != nil {
		return err
	}

	return nil
}

//CheckAccountSession - Checks if the session provided is valid
func (auth Authenticate) CheckAccountSession(session *types.Session) (*types.Account, error) {

	//Invalid sessionId
	if session.Token == "" {
		return nil, errors.New("Invalid Session Id")
	}

	account, err := manager.AccountManager{}.GetAccountSession(session, auth.DB, auth.Cache)
	if err != nil {
		return nil, err
	}

	//Get Account Roles
	account = account.GetAccountPermissions()

	//If account is ADMIN or above or 2FA is enabled then make sure device is verified.
	if utils.Contains("ADMIN", account.Roles) || account.TwoFA {
		device, err := manager.DeviceManager{}.GetDevice(session, auth.DB, auth.Cache)
		if err != nil {
			return nil, err
		}
		if device == nil {
			return nil, errors.New("No device found: " + account.Name)
		}
		if !device.Active {
			return nil, errors.New("Device not active - " + device.ID)
		}
	}
	return account, nil
}

//GetAllAccounts - Checks if the session provided is valid
func (auth Authenticate) GetAllAccounts(session *types.Session) (*[]types.Account, error) {
	account, err := auth.CheckAccountSession(session)
	if err != nil {
		return nil, err
	}

	//Get Account Roles
	account = account.GetAccountPermissions()

	//Only Accounts with ADMIN privliges can make this request
	if !utils.Contains("ADMIN", account.Roles) {
		return nil, errors.New("Invalid Privilges: " + account.Name)
	}

	accounts, err := manager.AccountManager{}.GetAllAccounts(auth.DB)
	if err != nil {
		return nil, err
	}

	return accounts, nil
}

//GetAccounts - returns accounts from the role given
func (auth Authenticate) GetAccounts(session *types.Session, roles []int) (*[]types.Account, error) {
	account, err := auth.CheckAccountSession(session)
	if err != nil {
		return nil, err
	}

	//Get Account Roles
	account = account.GetAccountPermissions()

	//Only Accounts with REGIONAL_SUPERVISOR privliges can make this request
	if !utils.Contains("ADMIN", account.Roles) {
		return nil, errors.New("Invalid Privilges: " + account.Name)
	}

	accounts, err := manager.AccountManager{}.GetAccounts(roles, auth.DB)
	if err != nil {
		return nil, err
	}

	return accounts, nil
}

//RegisterAccount - register a new account
func (auth Authenticate) RegisterAccount(session *types.Session, newAccount *types.Account) (string, error) {
	account, err := auth.CheckAccountSession(session)
	if err != nil {
		return "", err
	}

	//Get Account Roles
	account = account.GetAccountPermissions()

	//Only Accounts with ADMIN privliges can make this request
	if !utils.Contains("ADMIN", account.Roles) {
		return "", errors.New("Invalid Privilges: " + account.Name)
	}

	//Get newAccount Roles
	newAccount = newAccount.GetAccountPermissions()

	res, err := manager.AccountManager{}.CreateAccount(newAccount, account, auth.DB)
	if err != nil {
		return "", err
	}

	return res, nil
}

//UpdateAccountSettings - update account settings
func (auth Authenticate) UpdateAccountSettings(updatedAccount *types.Account, session *types.Session) (string, error) {
	account, err := auth.CheckAccountSession(session)
	if err != nil {
		return "", err
	}

	res, err := manager.AccountManager{}.UpdateAccountSettings(updatedAccount, account, auth.DB, auth.Cache)
	if err != nil {
		return "", err
	}

	return res, nil
}

//UpdateOtherAccountSettings - update account settings for another user
func (auth Authenticate) UpdateOtherAccountSettings(updatedAccount *types.Account, session *types.Session) (string, error) {
	account, err := auth.CheckAccountSession(session)
	if err != nil {
		return "", err
	}

	//Get Account Roles
	account = account.GetAccountPermissions()

	//Only Accounts with ADMIN privliges can make this request
	if !utils.Contains("ADMIN", account.Roles) {
		return "", errors.New("Invalid Privilges: " + account.Name)
	}

	accountData, err := manager.AccountManager{}.GetAccountByID(updatedAccount.ID, auth.DB)
	if err != nil {
		return "", err
	}

	//Get roles of the account we are trying to update
	accountData = accountData.GetAccountPermissions()

	accountData.Name = updatedAccount.Name
	accountData.UserName = updatedAccount.UserName
	accountData.Phone = updatedAccount.Phone
	accountData.Email = updatedAccount.Email
	accountData.Role = updatedAccount.Role

	res, err := manager.AccountManager{}.UpdateOtherAccountSettings(accountData, auth.DB, auth.Cache)
	if err != nil {
		return "", err
	}

	return res, nil
}

//DeleteAccount - deletes an account
func (auth Authenticate) DeleteAccount(del *types.DeleteAccountRequest, session *types.Session) (string, error) {
	account, err := auth.CheckAccountSession(session)
	if err != nil {
		return "", err
	}

	//Get Account Roles
	account = account.GetAccountPermissions()

	//Only Accounts with ADMIN privliges can make this request
	if !utils.Contains("ADMIN", account.Roles) {
		return "", errors.New("Invalid Privilges: " + account.Name)
	}

	delAccount, err := manager.AccountManager{}.GetAccountByID(del.ID, auth.DB)
	if err != nil {
		return "", err
	}

	//Get roles of the account we are trying to delete
	delAccount = delAccount.GetAccountPermissions()

	err = manager.AccountManager{}.DeleteAccount(delAccount, auth.DB, auth.Cache)
	if err != nil {
		return "", err
	}

	return "", nil
}

//ActivateDevice - activates a device
func (auth Authenticate) ActivateDevice(session *types.Session, deviceInfo *types.Device) error {
	//Make sure session of the request is valid
	account, err := auth.getAccountSession(session)
	if err != nil {
		return err
	}

	err = manager.DeviceManager{}.ActivateDevice(account, deviceInfo, auth.DB, auth.Cache)
	if err != nil {
		return err
	}

	return nil
}

//RecoverAccount - activates a device
func (auth Authenticate) RecoverAccount(account *types.Account) (*types.Recovery, error) {
	acc, err := manager.AccountManager{}.GetAccountByEmail(account.Email, auth.DB)
	if err != nil {
		return nil, err
	}

	if acc == nil {
		return nil, errors.New("Account not found for recovery: " + account.Email)
	}

	recovery, err := manager.RecoveryManager{}.CreateRecovery(acc, auth.DB)
	if err != nil {
		return nil, err
	}

	return recovery, nil
}

//GetRecovery - returns a recovery
func (auth Authenticate) GetRecovery(recovery *types.Recovery) (*types.Recovery, error) {
	rec, err := manager.RecoveryManager{}.GetRecovery(recovery, auth.DB)
	if err != nil {
		return nil, err
	}
	if rec == nil {
		return nil, errors.New("No recovery was found: " + recovery.ID)
	}
	return rec, nil
}

//FinishRecovery - completes a recovery request
func (auth Authenticate) FinishRecovery(recovery *types.RecoveryRequest) (string, error) {
	rec, err := manager.RecoveryManager{}.GetRecovery(&types.Recovery{ID: recovery.ID}, auth.DB)
	if err != nil {
		return "", err
	}

	if rec == nil {
		return "", errors.New("No recovery was found: " + recovery.ID)
	}

	account, err := manager.AccountManager{}.GetAccountByID(rec.AccountID, auth.DB)
	if err != nil {
		return "", err
	}
	if account == nil {
		return "", errors.New("No account was found: " + rec.AccountID)
	}

	res, err := manager.RecoveryManager{}.FinishRecovery(account, recovery, rec, auth.DB, auth.Cache)
	if err != nil {
		return "", err
	}

	return res, nil
}

//EnableTwoFA - enables two factor authentication for the requesting account
func (auth Authenticate) EnableTwoFA(session *types.Session) error {
	account, err := auth.CheckAccountSession(session)
	if err != nil {
		return err
	}

	err = manager.AccountManager{}.EnableTwoFA(account, auth.DB, auth.Cache)
	if err != nil {
		return err
	}

	return nil
}

//DisableTwoFA - disables two factor authentication for the requesting account
func (auth Authenticate) DisableTwoFA(session *types.Session) error {
	account, err := auth.CheckAccountSession(session)
	if err != nil {
		return err
	}

	//Get Account Roles
	account = account.GetAccountPermissions()

	//ADMIN accounts cannot disable 2FA. Even if they do they will still be required to activate a device
	if utils.Contains("ADMIN", account.Roles) {
		return errors.New("ADMIN account cannot disable TwoFA: " + account.Name)
	}

	err = manager.AccountManager{}.DisableTwoFA(account, auth.DB, auth.Cache)
	if err != nil {
		return err
	}

	return nil
}

//ChangeEmail - sends email change request to the email given
func (auth Authenticate) ChangeEmail(session *types.Session, emailRequest *types.EmailChangeRequest) (string, *types.EmailChange, error) {
	account, err := auth.CheckAccountSession(session)
	if err != nil {
		return "", nil, err
	}

	tmpAccount := &types.Account{Email: emailRequest.Email}

	//Check if is a valid email
	if err := tmpAccount.CheckEmail(); err != nil {
		return err.Error(), nil, nil
	}

	//Make sure email does not already exist
	res, err := manager.AccountManager{}.CheckDuplicates(tmpAccount, auth.DB)
	if err != nil {
		return "", nil, err
	}
	if res != "" {
		return res, nil, nil
	}

	request, err := manager.RecoveryManager{}.RequestEmailChange(account, emailRequest, auth.DB)
	if err != nil {
		return "", nil, err
	}

	return "", request, nil
}

//FinishEmailChange - completes an email change request
func (auth Authenticate) FinishEmailChange(emailRequest *types.EmailChangeRequest) error {

	emailChange, err := manager.RecoveryManager{}.GetEmailChange(emailRequest, auth.DB)
	if err != nil {
		return err
	}
	if emailChange == nil {
		return errors.New("No Email Change Request Found: " + emailRequest.ID)
	}

	account, err := manager.AccountManager{}.GetAccountByID(emailChange.AccountID, auth.DB)
	if err != nil {
		return err
	}
	if account == nil {
		return errors.New("No Account Found: " + emailChange.AccountID)
	}

	err = manager.RecoveryManager{}.FinishEmailChange(account, emailChange, auth.DB, auth.Cache)
	if err != nil {
		return err
	}

	return nil
}
