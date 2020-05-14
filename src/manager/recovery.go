package manager

import (
	"cache"
	"db"
	"errors"
	"time"
	"types"
	"utils"

	"github.com/kisielk/sqlstruct"
)

//RecoveryManager - recovery data access object
type RecoveryManager struct {
}

//CreateRecovery - creates a new recovery
func (rm RecoveryManager) CreateRecovery(account *types.Account, db *db.MySQL) (*types.Recovery, error) {

	recovery := types.Recovery{ID: utils.RandomString(), AccountID: account.ID, Created: time.Now(), Email: account.Email, UserName: account.UserName}

	stmt, err := db.PreparedQuery("INSERT INTO recover (id, accountId, created, email, userName) VALUES(?,?,?,?,?)")
	if err != nil {
		return nil, err
	}
	rows, err := stmt.Query(recovery.ID, recovery.AccountID, recovery.Created, recovery.Email, recovery.UserName)
	if err != nil {
		return nil, err
	}
	stmt.Close()
	defer rows.Close()
	return &recovery, nil

}

//GetRecovery - returns a recovery from db
func (rm RecoveryManager) GetRecovery(recovery *types.Recovery, db *db.MySQL) (*types.Recovery, error) {
	stmt, err := db.PreparedQuery("SELECT * FROM recover WHERE id = ?")
	if err != nil {
		return nil, err
	}
	rows, err := stmt.Query(recovery.ID)
	if err != nil {
		return nil, err
	}
	stmt.Close()
	defer rows.Close()
	for rows.Next() {
		rec := types.Recovery{}
		err = sqlstruct.Scan(&rec, rows)
		if err != nil {
			return nil, err
		}
		return &rec, nil
	}
	return nil, nil
}

//FinishRecovery - completes a account recovery process
func (rm RecoveryManager) FinishRecovery(account *types.Account, recoveryRequest *types.RecoveryRequest, recovery *types.Recovery, db *db.MySQL, cache *cache.Cache) (string, error) {

	//Validate that the password is correct
	if err := account.CheckPassword(); err != nil {
		return err.Error(), nil
	}

	//Hash the password
	hash, err := utils.HashPassword(recoveryRequest.Password)
	if err != nil {
		return "", err
	}

	//Set password to  account object
	account.Password = hash

	stmt, err := db.PreparedQuery("UPDATE users SET password = ? WHERE id = ?")
	if err != nil {
		return "", err
	}
	rows, err := stmt.Query(account.Password, account.ID)
	if err != nil {
		return "", err
	}
	stmt.Close()
	defer rows.Close()
	//Save updated account to cache
	AccountManager{}.SaveToCache(account, cache)

	//If this fails it will expire within the HOUR. The request is already completed.
	_, _ = db.SimpleQuery("DELETE FROM recover WHERE id = '" + recovery.ID + "'")

	return "", nil
}

//RequestEmailChange - creates a email change request
func (rm RecoveryManager) RequestEmailChange(account *types.Account, emailChange *types.EmailChangeRequest, db *db.MySQL) (*types.EmailChange, error) {

	emailRequest := types.EmailChange{ID: utils.RandomString(), AccountID: account.ID, OldEmail: account.Email, NewEmail: emailChange.Email, Created: time.Now()}

	stmt, err := db.PreparedQuery("INSERT INTO emailChange VALUES(?,?,?,?,?)")
	if err != nil {
		return nil, err
	}
	rows, err := stmt.Query(emailRequest.ID, emailRequest.AccountID, emailRequest.OldEmail, emailRequest.NewEmail, emailRequest.Created)
	if err != nil {
		return nil, err
	}
	stmt.Close()
	defer rows.Close()

	return &emailRequest, nil
}

//GetEmailChange - returns a email change request from db
func (rm RecoveryManager) GetEmailChange(changeRequest *types.EmailChangeRequest, db *db.MySQL) (*types.EmailChange, error) {
	stmt, err := db.PreparedQuery("SELECT * FROM emailChange WHERE id = ?")
	if err != nil {
		return nil, err
	}
	rows, err := stmt.Query(changeRequest.ID)
	if err != nil {
		return nil, err
	}
	stmt.Close()
	defer rows.Close()
	for rows.Next() {
		emailChange := types.EmailChange{}
		err = sqlstruct.Scan(&emailChange, rows)
		if err != nil {
			return nil, err
		}
		return &emailChange, nil
	}
	return nil, nil
}

//FinishEmailChange - finishes changing account email
func (rm RecoveryManager) FinishEmailChange(account *types.Account, emailChange *types.EmailChange, db *db.MySQL, cache *cache.Cache) error {

	account.Email = emailChange.NewEmail

	am := AccountManager{}

	email, err := am.GetAccountByEmail(account.Email, db)

	//Double check and make sure the new email is not taken
	if email != nil {
		//Remove email change request
		_, _ = db.SimpleQuery("DELETE FROM emailChange WHERE id = '" + emailChange.ID + "'")
		return errors.New("Email is taken: " + account.Email)
	}

	//Set new email to account
	stmt, err := db.PreparedQuery("UPDATE users SET email = ? WHERE id = ?")
	if err != nil {
		return err
	}
	rows, err := stmt.Query(account.Email, account.ID)
	if err != nil {
		return err
	}
	stmt.Close()
	defer rows.Close()

	//Save updated account to cache
	am.SaveToCache(account, cache)

	//If this fails it will expire within the HOUR. The request is already completed.
	_, _ = db.SimpleQuery("DELETE FROM emailChange WHERE id = '" + emailChange.ID + "'")

	return nil
}
