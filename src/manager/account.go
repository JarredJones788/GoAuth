package manager

import (
	"cache"
	"db"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"
	"types"
	"utils"

	"github.com/kisielk/sqlstruct"
)

//AccountManager - Account data access object
type AccountManager struct {
}

//SaveToCache - saves account object to redis cache
func (am AccountManager) SaveToCache(account *types.Account, cache *cache.Cache) {
	res, err := json.Marshal(account)
	if err != nil {
		fmt.Println(err)
		return
	}
	if cache.Set(account.Token, string(res)) != nil {
		fmt.Println("Failed saving account to cache")
	}
}

//CheckDuplicates - checks if account info already exists.
//Returns empty string and no error if no duplicates are found
//Returns string with an error message if duplicates are found
func (am AccountManager) CheckDuplicates(account *types.Account, db *db.MySQL) (string, error) {
	stmt, err := db.PreparedQuery("SELECT * FROM users WHERE (userName = ? AND id <> ?) OR (email = ? AND id <> ?)")
	if err != nil {
		return "", err
	}
	rows, err := stmt.Query(account.UserName, account.ID, account.Email, account.ID)
	if err != nil {
		return "", err
	}
	stmt.Close()
	defer rows.Close()
	for rows.Next() {
		duplicate := types.Account{}
		err = sqlstruct.Scan(&duplicate, rows)
		if err != nil {
			return "", err
		}
		if duplicate.UserName == account.UserName {
			return "UserName is taken: " + account.UserName, nil
		} else if duplicate.Email == account.Email {
			return "Email is taken: " + account.Email, nil
		}
		return "", nil
	}

	return "", nil
}

//RemoveSession - removes the session from cache
func (am AccountManager) RemoveSession(session *types.Session, cache *cache.Cache) error {
	cache.Del(session.Token)
	cache.Del(session.Device)
	return nil
}

//GetAllAccounts - returns all account from db
func (am AccountManager) GetAllAccounts(db *db.MySQL) (*[]types.Account, error) {
	rows, err := db.SimpleQuery("SELECT * FROM users ORDER BY name ASC")
	if err != nil {
		return nil, err
	}
	accounts := []types.Account{}
	defer rows.Close()
	for rows.Next() {
		account := types.Account{}
		err := sqlstruct.Scan(&account, rows)
		if err != nil {
			//fmt.Println(err) -> fails to convert NULL to datatype
		}
		account = account.HideInfo()
		account = account.GetPermissions()
		accounts = append(accounts, account)
	}
	return &accounts, nil
}

//GetAccounts - returns all accounts with the role given
func (am AccountManager) GetAccounts(roles []int, db *db.MySQL) (*[]types.Account, error) {

	if len(roles) <= 0 {
		return nil, errors.New("Roles array is empty")
	}

	query := "SELECT * FROM users WHERE role = '" + strconv.Itoa(roles[0]) + "'"

	for i, r := range roles {
		if i == 0 {
			continue
		}
		query += " OR role = '" + strconv.Itoa(r) + "'"
	}

	rows, err := db.SimpleQuery(query + " ORDER BY name ASC")
	if err != nil {
		return nil, err
	}
	accounts := []types.Account{}
	defer rows.Close()
	for rows.Next() {
		account := types.Account{}
		err := sqlstruct.Scan(&account, rows)
		if err != nil {
			//fmt.Println(err) -> fails to convert NULL to datatype
		}
		account = account.HideInfo()
		account = account.GetPermissions()
		accounts = append(accounts, account)
	}
	return &accounts, nil
}

//GetAccountSession - returns account from a session
func (am AccountManager) GetAccountSession(session *types.Session, db *db.MySQL, cache *cache.Cache) (*types.Account, error) {
	cachedAccount, err := cache.Get(session.Token)
	if err == nil { //Check if cache is available.
		var account *types.Account
		if json.Unmarshal([]byte(cachedAccount), &account) == nil {
			return account, nil
		}
	}
	stmt, err := db.PreparedQuery("SELECT * FROM users WHERE token = ?")
	if err != nil {
		return nil, err
	}
	rows, err := stmt.Query(session.Token)
	if err != nil {
		return nil, err
	}
	stmt.Close()
	defer rows.Close()
	for rows.Next() {
		account := types.Account{}
		err = sqlstruct.Scan(&account, rows)
		if err != nil {
			return nil, err
		}
		am.SaveToCache(&account, cache)
		return &account, nil
	}
	return nil, errors.New("No session found")
}

//UpdateAccountToken - returns account from an id **DOES NOT USE CACHE
func (am AccountManager) UpdateAccountToken(account *types.Account, db *db.MySQL) error {
	stmt, err := db.PreparedQuery("UPDATE users SET token = ? WHERE id = ?")
	if err != nil {
		return err
	}
	_, err = stmt.Query(account.Token, account.ID)
	if err != nil {
		return err
	}
	stmt.Close()
	return nil
}

//CreateAccount - verifies and creates a new account
func (am AccountManager) CreateAccount(account *types.Account, authedAccount *types.Account, db *db.MySQL) (string, error) {

	if err := account.CheckUserName(); err != nil {
		return err.Error(), nil
	}
	if err := account.CheckPassword(); err != nil {
		return err.Error(), nil
	}
	if err := account.CheckEmail(); err != nil {
		return err.Error(), nil
	}
	if err := account.CheckPhone(); err != nil {
		return err.Error(), nil
	}
	//Check if account details already exist with another account
	isDuplicate, err := am.CheckDuplicates(account, db)
	if err != nil {
		return "", err
	}
	if isDuplicate != "" {
		return isDuplicate, nil
	}

	//Setup account details
	account.ID = utils.RandomString()
	account.Token = utils.RandomString()
	account.Created = time.Now()

	//Hash password
	account.Password, err = utils.HashPassword(account.Password)
	if err != nil {
		return "", err
	}
	//Insert into database
	stmt, err := db.PreparedQuery("INSERT INTO users (id, userName, password, token, role, name, phone, email, created) VALUES(?,?,?,?,?,?,?,?,?)")
	if err != nil {
		return "", err
	}
	_, err = stmt.Query(account.ID, account.UserName, account.Password, account.Token, account.Role, account.Name, account.Phone, account.Email, account.Created)
	if err != nil {
		return "", err
	}
	stmt.Close()

	return "", nil
}

//DeleteAccount - deletes account from DB and cache.
func (am AccountManager) DeleteAccount(account *types.Account, db *db.MySQL, cache *cache.Cache) error {

	stmt, err := db.PreparedQuery("DELETE FROM users WHERE id = ?")
	if err != nil {
		return err
	}
	_, err = stmt.Query(account.ID)
	if err != nil {
		return err
	}

	cache.Del(account.Token)

	stmt.Close()
	return nil

}

//UpdateAccountSettings - updates the given account settings
func (am AccountManager) UpdateAccountSettings(updatedAccount *types.Account, account *types.Account, db *db.MySQL, cache *cache.Cache) (string, error) {

	if err := updatedAccount.CheckPhone(); err != nil {
		return err.Error(), nil
	}
	account.Name = updatedAccount.Name
	account.Phone = updatedAccount.Phone

	stmt, err := db.PreparedQuery("UPDATE users SET name = ?, phone = ? WHERE id = ?")
	if err != nil {
		return "", err
	}
	_, err = stmt.Query(account.Name, account.Phone, account.ID)
	if err != nil {
		return "", err
	}
	stmt.Close()

	am.SaveToCache(account, cache)

	return "", nil
}

//UpdateOtherAccountSettings - updates the another users account settings (ADMINS ONLY)
func (am AccountManager) UpdateOtherAccountSettings(updatedAccount *types.Account, db *db.MySQL, cache *cache.Cache) (string, error) {

	if err := updatedAccount.CheckPhone(); err != nil {
		return err.Error(), nil
	}
	if err := updatedAccount.CheckEmail(); err != nil {
		return err.Error(), nil
	}

	//Check if account details already exist with another account
	isDuplicate, err := am.CheckDuplicates(updatedAccount, db)
	if err != nil {
		return "", err
	}
	if isDuplicate != "" {
		return isDuplicate, nil
	}

	stmt, err := db.PreparedQuery("UPDATE users SET name = ?, userName = ?, email = ?, phone = ?, role = ? WHERE id = ?")
	if err != nil {
		return "", err
	}
	_, err = stmt.Query(updatedAccount.Name, updatedAccount.UserName, updatedAccount.Email, updatedAccount.Phone, updatedAccount.Role, updatedAccount.ID)
	if err != nil {
		return "", err
	}
	stmt.Close()

	am.SaveToCache(updatedAccount, cache)

	return "", nil
}

//GetAccountByEmail - returns an account by email
func (am AccountManager) GetAccountByEmail(email string, db *db.MySQL) (*types.Account, error) {
	stmt, err := db.PreparedQuery("SELECT * FROM users WHERE email = ?")
	if err != nil {
		return nil, err
	}
	rows, err := stmt.Query(email)
	if err != nil {
		return nil, err
	}
	stmt.Close()
	defer rows.Close()
	for rows.Next() {
		account := types.Account{}
		err = sqlstruct.Scan(&account, rows)
		if err != nil {
			return nil, err
		}
		return &account, nil
	}
	return nil, nil
}

//GetAccountLoginDetails - returns the account by checking if username or email matches what the user inputed.
func (am AccountManager) GetAccountLoginDetails(login string, db *db.MySQL) (*types.Account, error) {
	stmt, err := db.PreparedQuery("SELECT * FROM users WHERE email = ? OR username = ?")
	if err != nil {
		return nil, err
	}
	rows, err := stmt.Query(login, login)
	if err != nil {
		return nil, err
	}
	stmt.Close()
	defer rows.Close()
	for rows.Next() {
		account := types.Account{}
		err = sqlstruct.Scan(&account, rows)
		if err != nil {
			return nil, err
		}
		return &account, nil
	}
	return nil, errors.New("Invalid Username Or Email: " + login)
}

//GetAccountByID - returns an account by id
func (am AccountManager) GetAccountByID(id string, db *db.MySQL) (*types.Account, error) {
	stmt, err := db.PreparedQuery("SELECT * FROM users WHERE id = ?")
	if err != nil {
		return nil, err
	}
	rows, err := stmt.Query(id)
	if err != nil {
		return nil, err
	}
	stmt.Close()
	defer rows.Close()
	for rows.Next() {
		account := types.Account{}
		err = sqlstruct.Scan(&account, rows)
		if err != nil {
			return nil, err
		}
		return &account, nil
	}
	return nil, nil
}

//GetAccountFromUserName - returns account from an id **DOES NOT USE CACHE
func (am AccountManager) GetAccountFromUserName(userName string, db *db.MySQL) (*types.Account, error) {
	stmt, err := db.PreparedQuery("SELECT * FROM users WHERE username = ?")
	if err != nil {
		return nil, err
	}
	rows, err := stmt.Query(userName)
	if err != nil {
		return nil, err
	}
	stmt.Close()
	defer rows.Close()
	for rows.Next() {
		account := types.Account{}
		err = sqlstruct.Scan(&account, rows)
		if err != nil {
			return nil, err
		}
		return &account, nil
	}
	return nil, errors.New("No session found")
}

//EnableTwoFA - enables two factor authentication for the given account
func (am AccountManager) EnableTwoFA(account *types.Account, db *db.MySQL, cache *cache.Cache) error {
	stmt, err := db.PreparedQuery("UPDATE users SET twoFA = 1 WHERE id = ?")
	if err != nil {
		return err
	}
	rows, err := stmt.Query(account.ID)
	if err != nil {
		return err
	}
	stmt.Close()
	defer rows.Close()
	account.TwoFA = true

	am.SaveToCache(account, cache)

	return nil
}

//DisableTwoFA - disables two factor authentication for the given account
func (am AccountManager) DisableTwoFA(account *types.Account, db *db.MySQL, cache *cache.Cache) error {
	stmt, err := db.PreparedQuery("UPDATE users SET twoFA = 0 WHERE id = ?")
	if err != nil {
		return err
	}
	rows, err := stmt.Query(account.ID)
	if err != nil {
		return err
	}
	stmt.Close()
	defer rows.Close()
	account.TwoFA = false

	am.SaveToCache(account, cache)

	//Deletes all devices connected with this account
	_, _ = db.SimpleQuery("DELETE FROM devices WHERE accountId = '" + account.ID + "'")

	return nil
}
