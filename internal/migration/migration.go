package migration

import (
    "io"
    "os"
    //"log"
    "fmt"
	"time"
	"regexp"
    //"strings"
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
    "github.com/arch-mage/mdb"
    "github.com/ltkh/parking/internal/config"
    "golang.org/x/text/encoding/charmap"
)

func DecodeWindows1251(ba interface{}) string {
    if ba == nil {
        return ""
    }
    dec := charmap.Windows1251.NewDecoder()
    out, _ := dec.Bytes([]byte(ba.(string)))
    return string(out)
}

func Start(mdbfile string, conf *config.DB) error {
    // Open sqlite3
    db, err := sql.Open("sqlite3", conf.ConnString)
    if err != nil {
        return err
    }
    defer db.Close()

    // Open mdb
    file, err := os.Open(mdbfile)
    if err != nil {
        return err
    }
    defer file.Close()


    tables, err := mdb.Tables(file)
    if err != nil {
        return err
    }

    //tx, err := db.Begin()
    //if err != nil {
    //    return err
    //}

	db.Exec("delete from places")
	db.Exec("delete from checks")
	db.Exec("delete from owners")
	db.Exec("delete from prices")
	db.Exec("delete from parking")
    db.Exec("delete from main")

	re, _ := regexp.Compile(`.*~(\d+)$`)

    for _, table := range tables {
        if table.Sys { // skip system table
            continue
        }

		fmt.Printf("TABLE - %v\n", table)

        rows, err := mdb.Rows(file, table.Name)
        if err != nil {
            return err
        }

        for {
            fields, err := rows.Next()
            if err == io.EOF {
                break
            }
            if err != nil {
                return err
            }

            if table.Name == "SAvto" {
                _, err = db.Exec(
                    "replace into cars (id,number,brand,color,note) values (?,?,?,?,?)", 
                    DecodeWindows1251(fields[0]), 
                    DecodeWindows1251(fields[0]), 
                    DecodeWindows1251(fields[1]), 
                    DecodeWindows1251(fields[2]), 
                    DecodeWindows1251(fields[3]),
                )
                if err != nil {
                    return err
                }
            }

			if table.Name == "SVlad" {

                idOwner := re.FindStringSubmatch(DecodeWindows1251(fields[0]))

                _, err = db.Exec(
                    "replace into owners (id,idCar,fullName,telephone,address,document) values (?,?,?,?,?,?)", 
					idOwner[1],
                    DecodeWindows1251(fields[4]),
                    DecodeWindows1251(fields[1]),
					DecodeWindows1251(fields[3]),
                    DecodeWindows1251(fields[2]),
                    DecodeWindows1251(fields[5]), 
                )
                if err != nil {
                    return err
                }
				//fmt.Printf("idOwner - %v\n", DecodeWindows1251(fields[0]))
            }

            if table.Name == "SMesta" {
                _, err = db.Exec(
                    "replace into places (id,idOrg,number,description) values (?,?,?,?)", 
                    fields[0], 
                    0,
                    fields[0], 
                    DecodeWindows1251(fields[1]),
                )
                if err != nil {
                    return err
                }
            }
            
            if table.Name == "SPrise" {
                _, err = db.Exec(
                    "replace into prices (id,idOrg,carType,priceType,numOfDays,pricePerDay) values (?,?,?,?,?,?)", 
					fields[0],
                    0,
                    DecodeWindows1251(fields[1]),
                    DecodeWindows1251(fields[2]),
                    fields[3],
                    fields[4], 
                )
                if err != nil {
                    return err
                }
            }

            if table.Name == "Chek" {

				idCheck := re.FindStringSubmatch(DecodeWindows1251(fields[0]))

				startDate := int64(0)
				if fields[3] != nil {
					startDate = fields[3].(time.Time).UTC().Unix()
				}
                writeDate := int64(0)
				if fields[4] != nil {
					writeDate = fields[4].(time.Time).UTC().Unix()
				}

                //fmt.Printf("Type - %v\n", DecodeWindows1251(fields[8]))

                _, err = db.Exec(
                    "replace into checks (id,checkNumber,idOrg,carNumber,ownerFullName,startDate,writeDate,totalCost,numOfDays,priceType,placeNumber,userName) values (?,?,?,?,?,?,?,?,?,?,?,?)",
                    idCheck[1],                   //ID
                    fields[1],                    //Nomer
					0,
                    DecodeWindows1251(fields[2]), //AvtoNomer
                    "",
                    startDate,
                    writeDate,
                    fields[5],                    //SUMMA
                    fields[6],                    //DAYSE
                    //DecodeWindows1251(fields[7]), //TIPAVTO
                    DecodeWindows1251(fields[8]), //TIPPRISE
                    fields[9],                    //Mesto
                    "admin",
                )
                if err != nil {
                    return err
                }
            }

            if table.Name == "Stojanka" {

                idCheck := re.FindStringSubmatch(DecodeWindows1251(fields[3]))
                idOwner := re.FindStringSubmatch(DecodeWindows1251(fields[2]))

                if len(idCheck) < 2 {
                    idCheck = []string{"0","0"}
                }

                if len(idOwner) < 2 {
                    idOwner = []string{"0","0"}
                }

                startDate := int64(0)
                endDate := int64(0)

                if fields[4] != nil {
                    startDate = fields[4].(time.Time).UTC().Unix()
                }

                if fields[5] != nil {
                    endDate = fields[5].(time.Time).UTC().Unix()
                }

                //fmt.Printf("id - %v\n", fields[6])
                fmt.Printf("idCar - %v\n", DecodeWindows1251(fields[1]))
                //fmt.Printf("idOwner - %v\n", DecodeWindows1251(fields[2]))
                //fmt.Printf("idCheck - %v\n", idCheck[1])
                //fmt.Printf("idPlace - %v\n", fields[6])
                //fmt.Printf("startDate - %v\n", startDate)
                //fmt.Printf("endDate - %v\n", endDate)

                _, err = db.Exec(
                    "replace into parking (id,idOrg,idCar,idOwner,idCheck,idPlace,idUser,startDate,endDate,status) values (?,?,?,?,?,?,?,?,?,?)", 
                    DecodeWindows1251(fields[1]),
                    0,
                    DecodeWindows1251(fields[1]),
                    idOwner[1],
                    idCheck[1],
                    fields[6],
                    "admin",
                    startDate,
                    endDate,
                    1,
                )
                if err != nil {
                    return err
                }
            }

            if table.Name == "Main" {

                //fmt.Printf("UNP - %v\n", DecodeWindows1251(fields[0]))
                //fmt.Printf("NamePl - %v\n", DecodeWindows1251(fields[2]))
                //fmt.Printf("Ryk - %v\n", DecodeWindows1251(fields[4]))
                //fmt.Printf("LastNomer - %v\n", fields[6])
                //fmt.Printf("Chek - %v\n", fields[9])

                _, err = db.Exec(
                    "insert into main (id,idUser,name,fullName,address,telephone) values (?,?,?,?,?,?)",
                    0,
                    "admin",
					DecodeWindows1251(fields[1]),
                    DecodeWindows1251(fields[4]),
                    DecodeWindows1251(fields[3]),
                    DecodeWindows1251(fields[5]),
                )
                if err != nil {
                    return err
                }
            }
        }

	}

    return nil
}