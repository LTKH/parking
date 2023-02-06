package v1

import (
    "io"
    //"os"
    "log"
    "time"
    "regexp"
    "errors"
    "strconv"
    "net/http"
    "io/ioutil"
    "crypto/sha1"
    "encoding/hex"
    "encoding/json"
    "github.com/gorilla/websocket"
    "github.com/go-git/go-git/v5"
    "github.com/go-git/go-git/v5/storage/memory"
    "github.com/go-git/go-billy/v5"
    "github.com/go-git/go-billy/v5/memfs"
    "github.com/ltkh/parking/internal/db"
    "github.com/ltkh/parking/internal/config"
)

var (
    wsout = make(chan string, 1000)
    upgrader = websocket.Upgrader{
        ReadBufferSize:  1024,
        WriteBufferSize: 1024,
        CheckOrigin:     func(r *http.Request) bool { return true },
    }
)

type Api struct {
    conf         *config.Config
    db           *db.DbClient
}

type Resp struct {
    Status       string                    `json:"status"`
    Error        string                    `json:"error,omitempty"`
    Warnings     []string                  `json:"warnings,omitempty"`
    Data         interface{}               `json:"data,omitempty"`
}

func getHash(text string) string {
    h := sha1.New()
    io.WriteString(h, text)
    return hex.EncodeToString(h.Sum(nil))
}

func authentication(cfg *config.Config, r *http.Request) (string, int, error) {

    login, password, ok := r.BasicAuth()
    if ok {
        if login == cfg.Global.Security.AdminUser { 
            if password == cfg.Global.Security.AdminPassword {
                return login, 204, nil
            }
        }
        return login, 403, errors.New("Forbidden")
    }

    lg, err := r.Cookie("login")
    if err != nil {
        return "", 401, errors.New("Unauthorized")
    }
    tk, err := r.Cookie("token")
    if err != nil {
        return "", 401, errors.New("Unauthorized")
    }
    if lg.Value != "" && tk.Value != "" {
        if lg.Value == cfg.Global.Security.AdminUser { 
            if tk.Value == getHash(cfg.Global.Security.AdminPassword) {
                return login, 204, nil
            }
        }
        return login, 403, errors.New("Forbidden")
    }

    return "", 401, errors.New("Unauthorized")
}

func encodeResp(resp *Resp) []byte {
    jsn, err := json.Marshal(resp)
    if err != nil {
        return encodeResp(&Resp{Status:"error", Error:err.Error()})
    }
    return jsn
}

func New(conf *config.Config) (*Api, error) {
    client, err := db.NewClient(conf.DB)
    if err != nil {
        return &Api{}, err
    }

    if err := client.CreateTables(); err != nil {
        return &Api{}, err
    }

    return &Api{conf: conf, db: &client}, nil
}

func getFiles(fs billy.Filesystem, path string) ([]string, error) {
    var files []string

    fsrd, err := fs.ReadDir(path)
    if err != nil {
        return files, err
    }

    for _, f := range fsrd { 
        if f.IsDir() {
            gtfl, err := getFiles(fs, path+"/"+f.Name())
            if err != nil {
                return files, err
            }
            files = append(files, gtfl...)
        } else {
            files = append(files, path+"/"+f.Name())
        }
    }

    return files, nil
}

func updateFiles() {
    fs := memfs.New()
    //defer fs.Close()

    wsout <- "loading files..."

    _, err := git.Clone(memory.NewStorage(), fs, &git.CloneOptions{
        URL: "https://github.com/ltkh/parking.git",
    })
    if err != nil {
        log.Printf("[error] %v", err)
        wsout <- err.Error()
        return
    }

    wsout <- "saving files..."

    files, err := getFiles(fs, "web")
    if err != nil {
        wsout <- err.Error()
        log.Printf("[error] %v", err)
        return
    }

    for _, f := range files { 
        wsout <- "copy to "+f
        //io.Copy(os.Stdout, changelog)
    }

    wsout <- "finished"

    /*

    files, err := fs.ReadDir("web")
    if err != nil {
        wsout <- err.Error()
        log.Printf("[error] %v", err)
        return
    }

    for _, f := range files { 
        wsout <- "copy to web/"+f.Name()
        //io.Copy(os.Stdout, changelog)
    }
    */

    //ref, err := r.Head()

    //log.Printf("%v", ref)

    //commit, err := r.CommitObject(ref.Hash())

    //log.Printf("%v", commit)

    //history, err := commit.History()

    //log.Printf("%v", history)
}

func (api *Api) ApiLogin(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    err := r.ParseForm()
    if err != nil {
        w.WriteHeader(400)
        w.Write(encodeResp(&Resp{Status:"error", Error:err.Error()}))
        return
    }

    username := r.Form.Get("username")
    password := r.Form.Get("password")

    if username == "" || password == "" {
        w.WriteHeader(403)
        w.Write(encodeResp(&Resp{Status:"error", Error:"Login or password is empty"}))
        return
    }

    if username == api.conf.Global.Security.AdminUser { 
        if password == api.conf.Global.Security.AdminPassword {
            user := &config.User{
                Login: username,
                Token: getHash(password),
            }
            w.WriteHeader(200)
            w.Write(encodeResp(&Resp{Status:"success", Data:user}))
            return
        }
    }

    w.WriteHeader(403)
    w.Write(encodeResp(&Resp{Status:"error", Error:"Login or password is incorrect"}))
}

func (api *Api) WsEndpoint(w http.ResponseWriter, r *http.Request) {

    ws, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Printf("[error] %v", err)
        w.WriteHeader(500)
        return
    }
    defer ws.Close()

    for {
        if len(wsout) > 0 {
            msg, ok := <-wsout
            if ok {
                for i := 0; i < 5; i++ {
                    if err := ws.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
                        if i == 4 {
                            log.Printf("[error] %v", err)
                        } else {
                            time.Sleep(500 * time.Millisecond)
                            continue
                        }
                    }
                    break
                }
            }
        }
        time.Sleep(10 * time.Millisecond)
    }
    
}

func (api *Api) ApiUpdate(w http.ResponseWriter, r *http.Request) {
    updateFiles()
    w.WriteHeader(204)
}

func (api *Api) ApiObjects(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    _, code, err := authentication(api.conf, r)
    if err != nil {
        w.WriteHeader(code)
        w.Write(encodeResp(&Resp{Status:"error", Error:err.Error()}))
        return
    }

    re, _ := regexp.Compile(`^\/api\/v1\/(\w+)`)
    res := re.FindStringSubmatch(r.URL.Path)

    if r.Method == "GET" {

        row := map[string]interface{}{}

        for k, v := range r.URL.Query() {
            switch k {
                case "startDate":
                    i, err := strconv.Atoi(v[0])
                    if err != nil {
                        w.WriteHeader(500)
                        w.Write(encodeResp(&Resp{Status:"error", Error:err.Error()}))
                        return
                    }
                    row["startDate"] = int64(i)
                case "endDate":
                    i, err := strconv.Atoi(v[0])
                    if err != nil {
                        w.WriteHeader(500)
                        w.Write(encodeResp(&Resp{Status:"error", Error:err.Error()}))
                        return
                    }
                    row["endDate"] = int64(i)
            }
        }

        var rows []interface{} 

        rows, err := db.DbClient.LoadObjects(*api.db, res[1], row)
        if err != nil {
            log.Printf("[error] %v - %s", err, r.URL.Path)
            w.WriteHeader(500)
            w.Write(encodeResp(&Resp{Status:"error", Error:err.Error()}))
            return
        }

        w.WriteHeader(200)
        w.Write(encodeResp(&Resp{Status:"success", Data:rows}))
        return
    }

    if r.Method == "POST" {

        body, err := ioutil.ReadAll(r.Body)
        if err != nil {
            log.Printf("[error] %v - %s", err, r.URL.Path)
            w.WriteHeader(400)
            w.Write(encodeResp(&Resp{Status:"error", Error:err.Error()}))
            return
        }

        row := map[string]interface{}{}

        if err := json.Unmarshal(body, &row); err != nil {
            log.Printf("[error] %v - %s", err, r.URL.Path)
            w.WriteHeader(400)
            w.Write(encodeResp(&Resp{Status:"error", Error:err.Error()}))
            return
        }
        
        if err := db.DbClient.SaveObject(*api.db, res[1], row); err != nil {
            log.Printf("[error] %v - %s", err, r.URL.Path)
            w.WriteHeader(500)
            w.Write(encodeResp(&Resp{Status:"error", Error:err.Error()}))
            return
        }

        w.WriteHeader(200)
        w.Write(encodeResp(&Resp{Status:"success"}))
        return
    }

    if r.Method == "DELETE" {

        body, err := ioutil.ReadAll(r.Body)
        if err != nil {
            log.Printf("[error] %v - %s", err, r.URL.Path)
            w.WriteHeader(400)
            w.Write(encodeResp(&Resp{Status:"error", Error:err.Error()}))
            return
        }

        row := map[string]interface{}{}

        if err := json.Unmarshal(body, &row); err != nil {
            log.Printf("[error] %v - %s", err, r.URL.Path)
            w.WriteHeader(400)
            w.Write(encodeResp(&Resp{Status:"error", Error:err.Error()}))
            return
        }
        
        if err := db.DbClient.DeleteObject(*api.db, res[1], row["id"]); err != nil {
            log.Printf("[error] %v - %s", err, r.URL.Path)
            w.WriteHeader(500)
            w.Write(encodeResp(&Resp{Status:"error", Error:err.Error()}))
            return
        }

        w.WriteHeader(200)
        w.Write(encodeResp(&Resp{Status:"success"}))
        return
    }

    w.WriteHeader(405)
    w.Write(encodeResp(&Resp{Status:"error", Error:"Method Not Allowed"}))
}

func (api *Api) ApiParking(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    login, code, err := authentication(api.conf, r)
    if err != nil {
        w.WriteHeader(code)
        w.Write(encodeResp(&Resp{Status:"error", Error:err.Error()}))
        return
    }

    if r.Method == "GET" {
        var rows []config.Parking

        rows, err := db.DbClient.LoadParking(*api.db)
        if err != nil {
            log.Printf("[error] %v - %s", err, r.URL.Path)
            w.WriteHeader(500)
            w.Write(encodeResp(&Resp{Status:"error", Error:err.Error()}))
            return
        }
        w.WriteHeader(200)
        w.Write(encodeResp(&Resp{Status:"success", Data:rows}))
        return
    }

    if r.Method == "POST" {
        body, err := ioutil.ReadAll(r.Body)
        if err != nil {
            log.Printf("[error] %v - %s", err, r.URL.Path)
            w.WriteHeader(400)
            w.Write(encodeResp(&Resp{Status:"error", Error:err.Error()}))
            return
        }

        var row config.Parking

        if err := json.Unmarshal(body, &row); err != nil {
            log.Printf("[error] %v - %s", err, r.URL.Path)
            w.WriteHeader(400)
            w.Write(encodeResp(&Resp{Status:"error", Error:err.Error()}))
            return
        }
        
        if err := db.DbClient.SaveParking(*api.db, row, login); err != nil {
            log.Printf("[error] %v - %s", err, r.URL.Path)
            w.WriteHeader(500)
            w.Write(encodeResp(&Resp{Status:"error", Error:err.Error()}))
            return
        }

        w.WriteHeader(200)
        w.Write(encodeResp(&Resp{Status:"success"}))
        return
    }

    if r.Method == "DELETE" {
        body, err := ioutil.ReadAll(r.Body)
        if err != nil {
            log.Printf("[error] %v - %s", err, r.URL.Path)
            w.WriteHeader(400)
            w.Write(encodeResp(&Resp{Status:"error", Error:err.Error()}))
            return
        }

        var row map[string]int

        if err := json.Unmarshal(body, &row); err != nil {
            log.Printf("[error] %v - %s", err, r.URL.Path)
            w.WriteHeader(400)
            w.Write(encodeResp(&Resp{Status:"error", Error:err.Error()}))
            return
        }
        
        if err := db.DbClient.DeleteParking(*api.db, row["id"], login); err != nil {
            log.Printf("[error] %v - %s", err, r.URL.Path)
            w.WriteHeader(500)
            w.Write(encodeResp(&Resp{Status:"error", Error:err.Error()}))
            return
        }

        w.WriteHeader(200)
        w.Write(encodeResp(&Resp{Status:"success"}))
        return
    }

    w.WriteHeader(405)
    w.Write(encodeResp(&Resp{Status:"error", Error:"Method Not Allowed"}))
}

/*
func (api *Api) ApiCars(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    if r.Method == "GET" {
        rows, err := db.DbClient.LoadCars(*api.db)
        if err != nil {
            log.Printf("[error] %v - %s", err, r.URL.Path)
            w.WriteHeader(500)
            w.Write(encodeResp(&Resp{Status:"error", Error:err.Error()}))
            return
        }
        w.WriteHeader(200)
        w.Write(encodeResp(&Resp{Status:"success", Data:rows}))
        return
    }

    if r.Method == "POST" {
        object := config.Car{}

        body, err := ioutil.ReadAll(r.Body)
        if err != nil {
            log.Printf("[error] %v - %s", err, r.URL.Path)
            w.WriteHeader(400)
            w.Write(encodeResp(&Resp{Status:"error", Error:err.Error()}))
            return
        }

        if err := json.Unmarshal(body, &object); err != nil {
            log.Printf("[error] %v - %s", err, r.URL.Path)
            w.WriteHeader(400)
            w.Write(encodeResp(&Resp{Status:"error", Error:err.Error()}))
            return
        }
        
        if err := db.DbClient.SaveCar(*api.db, object); err != nil {
            log.Printf("[error] %v - %s", err, r.URL.Path)
            w.WriteHeader(500)
            w.Write(encodeResp(&Resp{Status:"error", Error:err.Error()}))
            return
        }

        w.WriteHeader(200)
        w.Write(encodeResp(&Resp{Status:"success"}))
        return
    }

    w.WriteHeader(405)
    w.Write(encodeResp(&Resp{Status:"error", Error:"Method Not Allowed"}))
}

func (api *Api) ApiOwners(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    if r.Method == "GET" {
        owners, err := db.DbClient.LoadOwners(*api.db)
        if err != nil {
            log.Printf("[error] %v - %s", err, r.URL.Path)
            w.WriteHeader(500)
            w.Write(encodeResp(&Resp{Status:"error", Error:err.Error()}))
            return
        }
        w.WriteHeader(200)
        w.Write(encodeResp(&Resp{Status:"success", Data:owners}))
        return
    }

    if r.Method == "POST" {
        owner := config.Owner{}

        body, err := ioutil.ReadAll(r.Body)
        if err != nil {
            log.Printf("[error] %v - %s", err, r.URL.Path)
            w.WriteHeader(400)
            w.Write(encodeResp(&Resp{Status:"error", Error:err.Error()}))
            return
        }

        if err := json.Unmarshal(body, &owner); err != nil {
            log.Printf("[error] %v - %s", err, r.URL.Path)
            w.WriteHeader(400)
            w.Write(encodeResp(&Resp{Status:"error", Error:err.Error()}))
            return
        }
        
        if err := db.DbClient.SaveOwner(*api.db, owner); err != nil {
            log.Printf("[error] %v - %s", err, r.URL.Path)
            w.WriteHeader(500)
            w.Write(encodeResp(&Resp{Status:"error", Error:err.Error()}))
            return
        }

        w.WriteHeader(200)
        w.Write(encodeResp(&Resp{Status:"success"}))
        return
    }

    w.WriteHeader(405)
    w.Write(encodeResp(&Resp{Status:"error", Error:"Method Not Allowed"}))
}

func (api *Api) ApiPlaces(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    if r.Method == "GET" {
        places, err := db.DbClient.LoadPlaces(*api.db)
        if err != nil {
            log.Printf("[error] %v - %s", err, r.URL.Path)
            w.WriteHeader(500)
            w.Write(encodeResp(&Resp{Status:"error", Error:err.Error()}))
            return
        }
        w.WriteHeader(200)
        w.Write(encodeResp(&Resp{Status:"success", Data:places}))
        return
    }

    if r.Method == "POST" {
        place := config.Place{}

        body, err := ioutil.ReadAll(r.Body)
        if err != nil {
            log.Printf("[error] %v - %s", err, r.URL.Path)
            w.WriteHeader(400)
            w.Write(encodeResp(&Resp{Status:"error", Error:err.Error()}))
            return
        }

        if err := json.Unmarshal(body, &place); err != nil {
            log.Printf("[error] %v - %s", err, r.URL.Path)
            w.WriteHeader(400)
            w.Write(encodeResp(&Resp{Status:"error", Error:err.Error()}))
            return
        }
        
        if err := db.DbClient.SavePlace(*api.db, place); err != nil {
            log.Printf("[error] %v - %s", err, r.URL.Path)
            w.WriteHeader(500)
            w.Write(encodeResp(&Resp{Status:"error", Error:err.Error()}))
            return
        }

        w.WriteHeader(200)
        w.Write(encodeResp(&Resp{Status:"success"}))
        return
    }

    w.WriteHeader(405)
    w.Write(encodeResp(&Resp{Status:"error", Error:"Method Not Allowed"}))
}

func (api *Api) ApiPrices(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    if r.Method == "GET" {
        prices, err := db.DbClient.LoadPrices(*api.db)
        if err != nil {
            log.Printf("[error] %v - %s", err, r.URL.Path)
            w.WriteHeader(500)
            w.Write(encodeResp(&Resp{Status:"error", Error:err.Error()}))
            return
        }
        w.WriteHeader(200)
        w.Write(encodeResp(&Resp{Status:"success", Data:prices}))
        return
    }

    if r.Method == "POST" {
        price := config.Price{}

        body, err := ioutil.ReadAll(r.Body)
        if err != nil {
            log.Printf("[error] %v - %s", err, r.URL.Path)
            w.WriteHeader(400)
            w.Write(encodeResp(&Resp{Status:"error", Error:err.Error()}))
            return
        }

        if err := json.Unmarshal(body, &price); err != nil {
            log.Printf("[error] %v - %s", err, r.URL.Path)
            w.WriteHeader(400)
            w.Write(encodeResp(&Resp{Status:"error", Error:err.Error()}))
            return
        }
        
        if err := db.DbClient.SavePrice(*api.db, price); err != nil {
            log.Printf("[error] %v - %s", err, r.URL.Path)
            w.WriteHeader(500)
            w.Write(encodeResp(&Resp{Status:"error", Error:err.Error()}))
            return
        }

        w.WriteHeader(200)
        w.Write(encodeResp(&Resp{Status:"success"}))
        return
    }

    w.WriteHeader(405)
    w.Write(encodeResp(&Resp{Status:"error", Error:"Method Not Allowed"}))
}
*/