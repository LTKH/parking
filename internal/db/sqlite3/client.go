package sqlite3

import (
    "fmt"
    "time"
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
            id            varchar(50) primary key,
            idOrg         bigint(20) not null,
            idCar         varchar(50) not null,
            idOwner       varchar(50) not null,
            idCheck       bigint(20) not null,
            idPlace       bigint(20) not null,
            idUser        varchar(100) not null,
            startDate     bigint(20) default 0,
            endDate       bigint(20) default 0,
            status        integer
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
            numOfDays     integer default 0,
            totalCost     float default 0
        );
        create table if not exists checks (
            id            bigint(20) primary key,
            idOrg         bigint(20) not null,
            carNumber     varchar(50) not null,
            carBrand      varchar(100) default '',
            carColor      varchar(50) default '',
            carType       varchar(100),
            placeNumber   integer not null,
            priceType     varchar(100),
            ownerFullName varchar(250) not null,
            checkNumber   bigint(20) default 0,
            writeDate     bigint(20) default 0,
            startDate     bigint(20) default 0,
            endDate       bigint(20) default 0,
            numOfDays     integer default 0,
            totalCost     float default 0,
            userName      bigint(20) default 0
        );
        create table if not exists users (
            id            varchar(100) primary key,
            idOrg         bigint(20) not null,
            password      varchar(100) not null,
            fullName      varchar(250) not null,
            address       varchar(1500) default '',
            telephone     varchar(100) default ''
        );
        create table if not exists main (
            id            bigint(20) primary key,
            idUser        varchar(100) not null,
            name          varchar(100) not null,
            fullName      varchar(250) not null,
            address       varchar(1500) default '',
            telephone     varchar(100) default ''
        );
    `)

    if err != nil {
        return err
    }

    return nil
}

func (db *Client) LoadUsers(values map[string]interface{}) ([]config.User, error) {
    result := []config.User{}
    rows, err := db.client.Query("select * from users order by id")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    for rows.Next() {
        var user config.User
        err := rows.Scan(&user.Id, &user.IdOrg, &user.Password, &user.FullName, &user.Address, &user.Telephone)
        if err != nil { return nil, err }
        result = append(result, user) 
    }

    return result, nil
}

func (db *Client) GetUser(login string) (config.User, error) {
    row := db.client.QueryRow("select id,idOrg,password from users where id = ?", login)
    var user config.User
    err := row.Scan(&user.Id, &user.IdOrg, &user.Password)
    if err != nil {
        return user, err
    }
    return user, nil
}

func (db *Client) SaveUser(user config.User) error {
    ursql := "replace into users (id,idOrg,password,fullName,address,telephone) values (?,?,?,?,?,?)"
    _, err := db.client.Exec(ursql, user.Id, user.IdOrg, &user.Password, &user.FullName, &user.Address, &user.Telephone)
    if err != nil {
        return err
    }
    return nil
}

func (db *Client) LoadCars(values map[string]interface{}) ([]config.Car, error) {
    result := []config.Car{}
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

    return result, nil
}

func (db *Client) SaveCar(object config.Car) error {
    if object.Id == "" {
        object.Id = object.Number
    }
    crsql := "replace into cars (id,number,brand,color,note) values (?,?,?,?,?)"
    _, err := db.client.Exec(crsql, object.Id, object.Number, object.Brand, object.Color, object.Note)
    if err != nil {
        return err
    }
    return nil
}

func (db *Client) LoadOwners(values map[string]interface{}) ([]config.Owner, error) {
    result := []config.Owner{}
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

    return result, nil
}

func (db *Client) SaveOwner(object config.Owner) error {
    if object.Id == "" {
        object.Id = object.Telephone
    }
    owsql := "replace into owners (id,idCar,fullName,telephone,address,document) values (?,?,?,?,?,?)"
    _, err := db.client.Exec(owsql, object.Id, object.IdCar, object.FullName, object.Telephone, object.Address, object.Document)
    if err != nil {
        return err
    }
    return nil
}

func (db *Client) LoadPlaces(values map[string]interface{}) ([]config.Place, error) {
    result := []config.Place{}
    endDate := int64(0)

    rows, err := db.client.Query(`
        select 
            places.id,
            places.idOrg,
            ifnull(parking.id, ''),
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

    return result, nil
}

func (db *Client) SavePlace(object config.Place) error {
    if object.Id == 0 {
        object.Id = time.Now().UTC().Unix()
    }
    plsql := "replace into places (id,idOrg,number,description) values (?,?,?,?)"
    _, err := db.client.Exec(plsql, object.Id, 1, object.Number, object.Description)
    if err != nil {
        return err
    }
    return nil
}

func (db *Client) LoadPrices(values map[string]interface{}) ([]config.Price, error) {
    result := []config.Price{}
    rows, err := db.client.Query(`
        select 
            id,
            idOrg,
            carType,
            priceType,
            ifnull(numOfDays, 0) numOfDays,
            ifnull(totalCost, 0) totalCost
        from prices order by id
    `)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    for rows.Next() {
        var price config.Price
        err := rows.Scan(&price.Id, &price.IdOrg, &price.CarType, &price.PriceType, &price.NumOfDays, &price.TotalCost)
        if err != nil { return nil, err }
        result = append(result, price)
    }

    return result, nil
}

func (db *Client) SavePrice(object config.Price) error {
    if object.Id == 0 {
        object.Id = time.Now().UTC().Unix()
    }
    prsql := "replace into prices (id,idOrg,carType,priceType,numOfDays,totalCost) values (?,?,?,?,?,?)"
    _, err := db.client.Exec(prsql, object.Id, 1, object.CarType, object.PriceType, object.NumOfDays, object.TotalCost)
    if err != nil {
        return err
    }
    return nil
}

func (db *Client) LoadMain(values map[string]interface{}) ([]config.Main, error) {
    result := []config.Main{}
    rows, err := db.client.Query("select * from main order by id")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    for rows.Next() {
        var main config.Main
        err := rows.Scan(&main.Id, &main.IdUser, &main.Name, &main.FullName, &main.Address, &main.Telephone)
        if err != nil { return nil, err }
        result = append(result, main)
    }

    return result, nil
}

func (db *Client) SaveMain(object config.Main, login string) error {
    if object.Id == 0 {
        object.Id = time.Now().UTC().Unix()

        ursql := "update users set idOrg = ? where id = ?"
        _, err := db.client.Exec(ursql, object.Id, login)
        if err != nil {
            return err
        }
    }
    mnsql := "replace into main (id,idUser,name,fullName,telephone,address) values (?,?,?,?,?,?)"
    _, err := db.client.Exec(mnsql, object.Id, login, object.Name, object.FullName, object.Telephone, object.Address)
    if err != nil {
        return err
    }
    
    return nil
}

func (db *Client) DeleteMain(id interface{}, login string) error {
    _, err := db.client.Exec("delete from main where id = ?", id)
    if err != nil {
        return err
    }

    ursql := "update users set idOrg = (select ifnull(max(id), 0) from main) where id = ?"
    _, err = db.client.Exec(ursql, login)
    if err != nil {
        return err
    }
    
    return nil
}

func (db *Client) LoadParking() ([]config.Parking, error) {
    result := []config.Parking{}

    rows, err := db.client.Query(`
        select 
            ifnull(parking.id, '') as id,
            ifnull(parking.idCheck, 0) as idCheck,
            ifnull(cars.number, parking.idCar) as carNumber, 
            ifnull(cars.brand, '') as brand, 
            ifnull(cars.color, '') as color, 
            ifnull(owners.fullName, '') as fullName, 
            ifnull(owners.telephone, '') as telephone, 
            ifnull(places.id, 0) as placeId,
            ifnull(places.number, 0) as placeNumber, 
            ifnull(parking.startDate, 0) as startDate, 
            ifnull(parking.endDate, 0) as endDate,
            ifnull(checks.checkNumber, 0) as checkNumber, 
            ifnull(checks.priceType, '') as priceType,
            ifnull(checks.writeDate, 0) as checkDate,
            ifnull(parking.status, 0) as status
        from parking
        left outer join cars on cars.id = parking.idCar
        left outer join owners on owners.id = parking.idOwner
        left outer join places on places.id = parking.idPlace
        left outer join checks on checks.id = parking.idCheck
        order by places.number
    `)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    for rows.Next() {

        object := config.Parking{}
        var checkDate int64
        var startDate int64
        var endDate int64

        err := rows.Scan(
            &object.Id,
            &object.IdCheck, 
            &object.CarNumber, 
            &object.Brand, 
            &object.Color, 
            &object.FullName, 
            &object.Telephone, 
            &object.IdPlace, 
            &object.PlaceNumber, 
            &startDate, 
            &endDate, 
            &object.CheckNumber, 
            &object.PriceType, 
            &checkDate, 
            &object.Status,
        )
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

    writeDate := time.Now().UTC().Unix()
    startDate := object.StartDate.UTC().Unix()
    endDate := object.EndDate.UTC().Unix()
    idCheck := time.Now().UTC().Unix()

    if object.Id == "" {

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

        prsql := "replace into parking (id,idOrg,idCar,idOwner,idCheck,idPlace,idUser,startDate,endDate,status) values (?,?,?,?,?,?,?,?,?,?)"
        _, err = db.client.Exec(prsql, object.CarNumber, 1, object.CarNumber, object.Telephone, idCheck, object.IdPlace, login, startDate, endDate, 1)
        if err != nil {
            return err
        }

    } else {

        prsql := "update parking set idCheck = ?,idPlace = ?,idUser = ?,startDate = ?,endDate = ?, status = 1 where id = ?"
        _, err := db.client.Exec(prsql, idCheck, object.IdPlace, login, startDate, endDate, object.Id)
        if err != nil {
            return err
        }

    }
 
    chsql := `
        insert into checks (id,idOrg,carNumber,carBrand,carColor,placeNumber,ownerFullName,checkNumber,writeDate,startDate,endDate,priceType,numOfDays,totalCost,userName) 
        values (?,(select max(idOrg) from users where id = ?),?,?,?,(select number from places where id = ?),?,(select max(ifnull(checkNumber, 0))+1 from checks),?,?,?,?,?,?,?)
    `
    _, err := db.client.Exec(chsql, idCheck, login, object.CarNumber, object.Brand, object.Color, object.IdPlace, object.FullName, writeDate, startDate, endDate, object.PriceType, object.Days, object.Cost, login)
    if err != nil {
        return err
    }

    return nil
}

func (db *Client) DeleteParking(id string, login string) error {

    prsql := "delete from parking where id = ?"
    _, err := db.client.Exec(prsql, id)
    if err != nil {
        return err
    }

    return nil
}

func (db *Client) LoadChecks(values map[string]interface{}) ([]config.Check, error) {
    result := []config.Check{}
    writeDate := int64(0)
    startDate := values["startDate"].(int64)
    endDate := values["endDate"].(int64)

    rows, err := db.client.Query(`
        select 
            checks.id as id, 
            checks.carNumber as carNumber, 
            checks.carBrand as carBrand,
            checks.carColor as carColor,
            ifnull(checks.placeNumber, '') as placeNumber,
            ifnull(checks.ownerFullName, '') as ownerFullName,
            ifnull(checks.checkNumber, 0) as checkNumber,
            checks.priceType as priceType,
            checks.writeDate as writeDate,
            checks.totalCost as totalCost,
            ifnull(checks.userName, '') as userName
        from checks
        where checks.writeDate >= ? and checks.writeDate < ?
        order by checks.writeDate
    `, startDate, endDate)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    for rows.Next() {
        var object config.Check
        err := rows.Scan(
            &object.Id, &object.CarNumber, &object.CarBrand, &object.CarColor, &object.PlaceNumber, &object.FullName, 
            &object.CheckNumber, &object.PriceType, &writeDate, &object.TotalCost, &object.UserName,
        )
        if err != nil { return nil, err }
        object.WriteDate = time.Unix(writeDate, 0)
        result = append(result, object) 
    }

    return result, nil
}

func (db *Client) LoadCheck(id int64, login string) (interface{}, error) {
    var writeDate int64
    var startDate int64
    var endDate int64

    row := db.client.QueryRow(`
        select 
            checks.id as id,
            checks.carNumber as carNumber, 
            checks.carBrand as carBrand,
            checks.carColor as carColor,
            ifnull(checks.placeNumber, '') as placeNumber,
            ifnull(checks.ownerFullName, '') as ownerFullName,
            ifnull(checks.checkNumber, 0) as checkNumber,
            checks.priceType as priceType,
            checks.writeDate as writeDate,
            checks.startDate as startDate,
            checks.endDate as endDate,
            ifnull(checks.totalCost, 0) as totalCost,
            ifnull(checks.numOfDays, 0) as numOfDays,
            ifnull(checks.userName, '') as userName,
            ifnull(main.Name, '') as mainName,
            ifnull(main.FullName, '') as mainFullName,
            ifnull(main.Address, '') as mainAddress,
            ifnull(main.Telephone, '') as mainTelephone
        from checks
        left outer join main on main.id = (select idOrg from users where id = ?)
        where checks.id = ?
    `, login, id)
    var object config.Check
    err := row.Scan(
        &object.Id, 
        &object.CarNumber, 
        &object.CarBrand, 
        &object.CarColor, 
        &object.PlaceNumber, 
        &object.FullName, 
        &object.CheckNumber, 
        &object.PriceType, 
        &writeDate, 
        &startDate, 
        &endDate,
        &object.TotalCost, 
        &object.NumOfDays,
        &object.UserName,
        &object.MainName,
        &object.MainFullName,
        &object.MainAddress,
        &object.MainTelephone,
    )
    if err != nil { return object, err }
    object.WriteDate = time.Unix(writeDate, 0)
    object.StartDate = time.Unix(startDate, 0)
    object.EndDate = time.Unix(endDate, 0)

    return object, nil
}

func (db *Client) DeleteOldChecks() (int64, error) {

    stmt, err := db.client.Prepare("delete from checks where writeDate < ?")
    if err != nil {
        return 0, err
    }
    defer stmt.Close()

    res, err := stmt.Exec(time.Now().UTC().Unix() - 86400 * 365)
    if err != nil {
        return 0, err
    }

    cnt, err := res.RowsAffected()
    if err != nil {
        return 0, err
    }

    return cnt, nil

}

func (db *Client) DeleteObject(table string, id interface{}) error {

    _, err := db.client.Exec(fmt.Sprintf("delete from %s where id = ?", table), id)
    if err != nil {
        return err
    }
    return nil

}