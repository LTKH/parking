package config

import (
    "time"
    "io/ioutil"
    "gopkg.in/yaml.v2"
)

type Config struct {
    Global           *Global                 `yaml:"global"`
    DB               *DB                     `yaml:"db"`
    Window           *Window                 `yaml:"window"`
}

type Global struct {
    ListenAddr       string                  `yaml:"listen_address"`
    CertFile         string                  `yaml:"cert_file"`
    CertKey          string                  `yaml:"cert_key"`
    WebDir           string                  `yaml:"web_dir"`
    Security         *Security               `yaml:"security"`
}

type Security struct {
    AdminUser        string                  `yaml:"admin_user"`
    AdminPassword    string                  `yaml:"admin_password"`
}

type DB struct {
    Client           string                  `yaml:"client"`
    ConnString       string                  `yaml:"conn_string"`
    HistoryDays      int                     `yaml:"history_days"`
}

type Window struct {
    Enabled          bool                    `yaml:"enabled"`
    Width            int                     `yaml:"width"`
    Height           int                     `yaml:"height"`
    Navigate         string                  `yaml:"navigate"`
}

type Car struct {
    Id               string                  `json:"id"`
    Number           string                  `json:"number"`
    Brand            string                  `json:"brand"`
    Color            string                  `json:"color"`
    Note             string                  `json:"note"`
}

type Owner struct {
    Id               string                  `json:"id"`
    IdCar            string                  `json:"idCar"`
    FullName         string                  `json:"fullName"`
    Address          string                  `json:"address"`
    Telephone        string                  `json:"telephone"`
    Document         string                  `json:"document"`
}

type Place struct {
    Id               int64                   `json:"id"`
    IdOrg            int64                   `json:"idOrg"`
    IdPark           string                  `json:"idPark"`
    EndDate          time.Time               `json:"endDate"`
    Number           int                     `json:"number"`
    Description      string                  `json:"description"`
}

type Price struct {
    Id               int64                   `json:"id"`
    IdOrg            int64                   `json:"idOrg"`
    CarType          string                  `json:"carType"`
    PriceType        string                  `json:"priceType"`
    NumOfDays        int                     `json:"numOfDays"`
    TotalCost        float32                 `json:"totalCost"`
}

type Parking struct {
    Id               string                  `json:"id"`
    IdCheck          int64                   `json:"idCheck"`
    CarNumber        string                  `json:"carNumber"`
    Brand            string                  `json:"brand"`
    Color            string                  `json:"color"`
    Note             string                  `json:"note"`
    FullName         string                  `json:"fullName"`
    Address          string                  `json:"address"`
    Telephone        string                  `json:"telephone"`
    Document         string                  `json:"document"`
    StartDate        time.Time               `json:"startDate"`
    EndDate          time.Time               `json:"endDate"`
    Days             int                     `json:"days"`
    IdPlace          int64                   `json:"idPlace"`
    PlaceNumber      int                     `json:"placeNumber"`
    Price            string                  `json:"price"`
    PriceType        string                  `json:"priceType"`
    Cost             float32                 `json:"cost"`
    Debtor           int                     `json:"debtor"`
    CheckNumber      int                     `json:"checkNumber"`
    CheckDate        time.Time               `json:"checkDate"`
    Status           int                     `json:"status"`
}

type User struct {
    Id               string                  `json:"login"`
    IdOrg            int64                   `json:"idOrg"`
    Password         string                  `json:"token"`
    FullName         string                  `json:"fullName"`
    Address          string                  `json:"address"`
    Telephone        string                  `json:"telephone"`
}

type Check struct {
    Id               int64                   `json:"id"`
    CarNumber        string                  `json:"carNumber"`
    CarBrand         string                  `json:"carBrand"`
    CarColor         string                  `json:"carColor"`
    CarType          string                  `json:"carType"`
    PriceType        string                  `json:"priceType"`
    FullName         string                  `json:"fullName"`
    PlaceNumber      int                     `json:"placeNumber"`
    CheckNumber      int                     `json:"checkNumber"`
    WriteDate        time.Time               `json:"writeDate"`
    StartDate        time.Time               `json:"startDate"`
    EndDate          time.Time               `json:"endDate"`
    TotalCost        float32                 `json:"totalCost"`
    NumOfDays        int                     `json:"numOfDays"`
    UserName         string                  `json:"userName"`
    MainName         string                  `json:"mainName,omitempty"`
    MainFullName     string                  `json:"mainFullName,omitempty"`
    MainAddress      string                  `json:"mainAddress,omitempty"`
    MainTelephone    string                  `json:"mainTelephone,omitempty"`
}

type Main struct {
    Id               int64                   `json:"id"`
    IdUser           string                  `json:"idUser"`
    Name             string                  `json:"name"`
    FullName         string                  `json:"fullName"`
    Address          string                  `json:"address"`
    Telephone        string                  `json:"telephone"`
}

func New(filename string) (*Config, error) {

    cfg := &Config{}

    content, err := ioutil.ReadFile(filename)
    if err != nil {
       return cfg, err
    }

    if err := yaml.UnmarshalStrict(content, cfg); err != nil {
        return cfg, err
    }
    
    return cfg, nil
}