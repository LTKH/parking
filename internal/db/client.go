package db

import (
    "errors"
    "github.com/ltkh/parking/internal/config"
    "github.com/ltkh/parking/internal/db/sqlite3"    
)

type DbClient interface {
    Close() error
    CreateTables() error
    //Objects
    LoadObjects(table string, values map[string]interface{}) ([]interface{}, error)
    SaveObject(table string, object map[string]interface{}) error
    DeleteObject(table string, id interface{}) error
    //Parking
    LoadParking() ([]config.Parking, error)
    SaveParking(object config.Parking, login string) error
    DeleteParking(id interface{}, login string) error
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