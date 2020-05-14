package manager

import (
	"cache"
	"db"
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"types"
	"utils"

	"github.com/kisielk/sqlstruct"
)

//DeviceManager - device data access object
type DeviceManager struct {
}

//SaveToCache - saves device object to redis cache
func (dm DeviceManager) SaveToCache(device *types.Device, cache *cache.Cache) {
	res, err := json.Marshal(device)
	if err != nil {
		fmt.Println(err)
		return
	}
	if cache.Set(device.ID, string(res)) != nil {
		fmt.Println("Failed saving device to cache")
	}
}

//GetDevice - returns device from a session
func (dm DeviceManager) GetDevice(session *types.Session, db *db.MySQL, cache *cache.Cache) (*types.Device, error) {
	cachedDevice, err := cache.Get(session.Device)
	if err == nil { //Check if cache is available.
		var device *types.Device
		if json.Unmarshal([]byte(cachedDevice), &device) == nil {
			return device, nil
		}
	}
	stmt, err := db.PreparedQuery("SELECT * FROM devices WHERE id = ?")
	if err != nil {
		return nil, err
	}
	rows, err := stmt.Query(session.Device)
	if err != nil {
		return nil, err
	}
	stmt.Close()
	defer rows.Close()
	for rows.Next() {
		device := types.Device{}
		err = sqlstruct.Scan(&device, rows)
		if err != nil {
			return nil, err
		}
		dm.SaveToCache(&device, cache)
		return &device, nil
	}
	return nil, nil
}

//CreateDevice - creates a new device
func (dm DeviceManager) CreateDevice(account *types.Account, db *db.MySQL) (*types.Device, error) {
	device := types.Device{ID: utils.RandomString(), AccountID: account.ID, Created: time.Now(), Active: false, Code: utils.RandomCode()}

	stmt, err := db.PreparedQuery("INSERT INTO devices (id, accountId, created, active, code) VALUES(?,?,?,?,?)")
	if err != nil {
		return nil, err
	}
	rows, err := stmt.Query(device.ID, device.AccountID, device.Created, device.Active, device.Code)
	if err != nil {
		return nil, err
	}
	stmt.Close()
	defer rows.Close()
	return &device, nil
}

//ActivateDevice - activates a device if the code is correct and the account matches the device
func (dm DeviceManager) ActivateDevice(account *types.Account, deviceInfo *types.Device, db *db.MySQL, cache *cache.Cache) error {

	device, err := dm.GetDevice(&types.Session{Device: deviceInfo.ID}, db, cache)
	if err != nil {
		return err
	}
	if device == nil {
		return errors.New("No device was found: " + account.Name)
	}

	if device.Active {
		return errors.New("Device is already active: " + account.Name)
	}

	if device.AccountID != account.ID {
		return errors.New("Device does not belong to the account: " + account.Name)
	}

	if device.Code != deviceInfo.Code {
		return errors.New("Invalid device code: " + account.Name)
	}

	device.Active = true

	stmt, err := db.PreparedQuery("UPDATE devices SET active = 1 WHERE id = ?")
	if err != nil {
		return err
	}
	rows, err := stmt.Query(device.ID)
	if err != nil {
		return err
	}
	stmt.Close()
	defer rows.Close()
	dm.SaveToCache(device, cache)
	return nil
}
