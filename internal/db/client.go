package db

import (
    "errors"
    "github.com/ltkh/parking/internal/config"
    "github.com/ltkh/parking/internal/db/sqlite3"    
)

type DbClient interface {
    Close() error
    CreateTables() error
    //Cars
    LoadCars(values map[string]interface{}) ([]config.Car, error)
    SaveCar(object config.Car) error
    //Owners
    LoadOwners(values map[string]interface{}) ([]config.Owner, error)
    SaveOwner(object config.Owner) error
    //Places
    LoadPlaces(values map[string]interface{}) ([]config.Place, error)
    SavePlace(object config.Place) error
    //Prices
    LoadPrices(values map[string]interface{}) ([]config.Price, error)
    SavePrice(object config.Price) error
    //Main
    LoadMain(values map[string]interface{}) ([]config.Main, error)
    SaveMain(object config.Main, login string) error
    DeleteMain(id interface{}, login string) error
    //Parking
    LoadParking() ([]config.Parking, error)
    SaveParking(object config.Parking, login string) error
    DeleteParking(id string, login string) error
    //User
    LoadUsers(values map[string]interface{}) ([]config.User, error)
    SaveUser(user config.User) error
    GetUser(login string) (config.User, error)
    //Check
    LoadChecks(values map[string]interface{}) ([]config.Check, error)
    LoadCheck(id int64, login string) (interface{}, error)
    DeleteOldChecks() (int64, error)
    //Objects
    DeleteObject(table string, id interface{}) error
}

func NewClient(config *config.DB) (DbClient, error) {
    switch config.Client {
        //case "mysql":
        //    return mysql.NewClient(config)
        case "sqlite3":
            return sqlite3.NewClient(config)
    }
    return nil, errors.New("invalid client")
}