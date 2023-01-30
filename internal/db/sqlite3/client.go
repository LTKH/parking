package sqlite3

import (
    "fmt"
    "time"
    "strings"
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
    "github.com/ltkh/parking/internal/config"
)

type Client struct {
    client *sql.DB
    config *config.DB
}

func NewClient(conf *config.DB) (*Client, error) {
    conn, err := sql.Open("sqlite3", conf.ConnString)
    if err != nil {
        return nil, err
    }
    return &Client{ client: conn, config: conf }, nil
}

func (db *Client) Close() error {
    db.client.Close()

    return nil
}

func (db *Client) CreateTables() error {
    _, err := db.client.Exec(`
        create table if not exists cars (
            id            varchar(50) primary key,
            number        varchar(50) not null,
            brand         varchar(100) default '',
            color         varchar(50) default '',
            note          varchar(250) default ''
        );
        create table if not exists owners (
            id            varchar(50) primary key,
            idCar         varchar(50) not null,
            fullName      varchar(250) not null,
            telephone     varchar(50) not null,
            address       varchar(1500) default '',
            document      varchar(150) default ''
        );
        create table if not exists parking (
            id            bigint(20) primary key,
            idOrg         bigint(20) not null,
            idCar         varchar(50) not null,
            idOwner       varchar(50) not null,
            idCheck       bigint(20) not null,
            idPlace       bigint(20) not null,
            idUser        varchar(100) not null,
            startDate     bigint(20) default 0,
            endDate       bigint(20) default 0
        );
        create table if not exists places (
            id            bigint(20) primary key,
            idOrg         bigint(20) not null,
            number        integer not null,
            description   varchar(1500) default ''
        );
        create table if not exists prices (
            id            bigint(20) primary key,
            idOrg         bigint(20) not null,
            carType       varchar(100),
            priceType     varchar(100),
            numOfDays     integer,
            pricePerDay   float
        );
        create table if not exists users (
            id            bigint(20) primary key,
            idOrg         bigint(20) not null,
            password      varchar(100) not null,
            fullName      varchar(250) not null,
            address       varchar(1500) default '',
            telephone     varchar(100) default ''
        );
        create table if not exists checks (
            id            bigint(20) primary key,
            idOrg         bigint(20) not null,
            idCar         varchar(50) not null,
            idPlace       bigint(20) not null,
            idUser        varchar(100) not null,
            number        bigint(20) default 0,
            writeDate     bigint(20) default 0,
            startDate     bigint(20) default 0,
            carType       varchar(100),
            priceType     varchar(100),
            numOfDays     integer,
            totalCost     float
        );
        create table if not exists main (
            id            bigint(20) primary key,
            idUser        varchar(100) not null,
            name          varchar(100) not null,
            address       varchar(1500) default '',
            telephone     varchar(100) default ''
        );
    `)

    if err != nil {
        return err
    }

    return nil
}

func (db *Client) LoadObjects(table string) ([]interface{}, error) {
    result := []interface{}{}

    if table == "cars" {
        rows, err := db.client.Query("select * from cars order by id")
        if err != nil {
            return nil, err
        }
        defer rows.Close()

        for rows.Next() {
            var car config.Car
            err := rows.Scan(&car.Id, &car.Number, &car.Brand, &car.Color, &car.Note)
            if err != nil { return nil, err }
            result = append(result, car) 
        }
    }

    if table == "owners" {
        rows, err := db.client.Query("select * from owners order by id")
        if err != nil {
            return nil, err
        }
        defer rows.Close()

        for rows.Next() {
            var owner config.Owner
            err := rows.Scan(&owner.Id, &owner.IdCar, &owner.FullName, &owner.Telephone, &owner.Address, &owner.Document)
            if err != nil { return nil, err }
            result = append(result, owner) 
        }
    }

    if table == "prices" {
        rows, err := db.client.Query("select * from prices order by id")
        if err != nil {
            return nil, err
        }
        defer rows.Close()

        for rows.Next() {
            var price config.Price
            err := rows.Scan(&price.Id, &price.IdOrg, &price.CarType, &price.PriceType, &price.NumOfDays, &price.PricePerDay)
            if err != nil { return nil, err }
            result = append(result, price)
        }
    }

    if table == "places" {
        endDate := int64(0)

        rows, err := db.client.Query(`
            select 
                places.id,
                places.idOrg,
                ifnull(parking.id, 0),
                ifnull(parking.endDate, 0),
                places.number,
                places.description 
            from places 
            left outer join parking on parking.idPlace = places.id
            order by places.number
        `)
        if err != nil {
            return nil, err
        }
        defer rows.Close()

        for rows.Next() {
            var object config.Place
            err := rows.Scan(&object.Id, &object.IdOrg, &object.IdPark, &endDate, &object.Number, &object.Description)
            if err != nil { return nil, err }
            object.EndDate = time.Unix(endDate, 0)
            result = append(result, object) 
        }
    }

    if table == "main" {
        rows, err := db.client.Query("select * from main order by id")
        if err != nil {
            return nil, err
        }
        defer rows.Close()

        for rows.Next() {
            var main config.Main
            err := rows.Scan(&main.Id, &main.IdUser, &main.Name, &main.Address, &main.Telephone)
            if err != nil { return nil, err }
            result = append(result, main)
        }
    }

    return result, nil
    
}

func (db *Client) SaveObject(table string, object map[string]interface{}) error {
    
    fields := []string{}
    values := []interface{}{}
    count  := []string{}
    for k, v := range object {
        fields = append(fields, k)
        values = append(values, v)
        count  = append(count, "?")
    }

    stmt, err := db.client.Prepare(fmt.Sprintf("replace into %s (%s) values (%s)", table, strings.Join(fields, ","), strings.Join(count, ",")))
    if err != nil { 
        return err 
    }
    defer stmt.Close()

    _, err = stmt.Exec(values...)
    if err != nil { 
        return err 
    }

    return nil
}

func (db *Client) DeleteObject(table string, id interface{}) error {
    stmt, err := db.client.Prepare(fmt.Sprintf("delete from %s where id = ?", table))
    if err != nil {
        return err
    }
    defer stmt.Close()

    _, err = stmt.Exec(id)
    if err != nil {
        return err
    }

    return nil
}

func (db *Client) LoadParking() ([]config.Parking, error) {
    result := []config.Parking{}

    rows, err := db.client.Query(`
        select 
            ifnull(parking.id,0) as id,
            ifnull(cars.number,'') as carNumber, 
            ifnull(cars.brand,'') as brand, 
            ifnull(owners.fullName, '') as fullName, 
            ifnull(owners.telephone, '') as telephone, 
            ifnull(places.number, 0) as placeNumber, 
            ifnull(parking.startDate, 0) as startDate, 
            ifnull(parking.endDate, 0) as endDate,
            ifnull(checks.number, 0) as checkNumber, 
            ifnull(checks.writeDate, 0) as checkDate
        from parking
        left outer join cars on cars.id = parking.idCar
        left outer join owners on owners.id = parking.idOwner
        left outer join places on places.id = parking.idPlace
        left outer join checks on checks.id = parking.idCheck
        order by parking.id
    `)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    for rows.Next() {

        object := config.Parking{}
        var startDate int64
        var endDate int64
        var checkDate int64

        err := rows.Scan(&object.Id, &object.CarNumber, &object.Brand, &object.FullName, &object.Telephone, &object.Place, &startDate, &endDate, &object.CheckNumber, &checkDate)
        if err != nil {
            return nil, err
        }

        if endDate < time.Now().UTC().Unix() {
            object.Debtor = 1
        }
        object.StartDate = time.Unix(startDate, 0)
        object.EndDate = time.Unix(endDate, 0)
        object.CheckDate = time.Unix(checkDate, 0)
        result = append(result, object) 
    }

    return result, nil
}

func (db *Client) SaveParking(object config.Parking, login string) error {

    idCheck := time.Now().UTC().UnixMicro()
    writeDate := time.Now().UTC().Unix()
    startDate := object.StartDate.UTC().Unix()
    endDate := object.EndDate.UTC().Unix()

    crsql := "replace into cars (id,number,brand,color,note) values (?,?,?,?,?)"
    _, err := db.client.Exec(crsql, object.CarNumber, object.CarNumber, object.Brand, object.Color, object.Note)
    if err != nil {
        return err
    }

    owsql := "replace into owners (id,idCar,fullName,telephone,address,document) values (?,?,?,?,?,?)"
    _, err = db.client.Exec(owsql, object.Telephone, object.CarNumber, object.FullName, object.Telephone, object.Address, object.Document)
    if err != nil {
        return err
    }
    
    chsql := "insert into checks (id,idOrg,number,idCar,writeDate,startDate,numOfDays,idPlace,idUser,totalCost) values (?,?,(select max(number)+1 from checks where idOrg = ?),?,?,?,?,?,?,?)"
    _, err = db.client.Exec(chsql, idCheck, 0, 0, object.CarNumber, writeDate, startDate, object.Days, object.Place, login, object.Cost)
    if err != nil {
        return err
    }

    prsql := "replace into parking (id,idOrg,idCar,idOwner,idCheck,idPlace,idUser,startDate,endDate) values (?,?,?,?,?,?,?,?,?)"
    _, err = db.client.Exec(prsql, object.Place, 0, object.CarNumber, object.Telephone, idCheck, object.Place, login, startDate, endDate)
    if err != nil {
        return err
    }

    return nil
}

func (db *Client) DeleteParking(id interface{}, login string) error {

    prsql := "update parking set idPlace = 0, endDate = 0, idUser = ? where id = ?"
    _, err := db.client.Exec(prsql, login, id)
    if err != nil {
        return err
    }

    return nil
}