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
    "html/template"
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

func authentication(api *Api, r *http.Request) (string, int, error) {

    login, password, ok := r.BasicAuth()
    if ok {
        user, err := db.DbClient.GetUser(*api.db, login)
        if err != nil {
            log.Printf("[error] %v", err)
            return "", 401, errors.New("Unauthorized")
        }
        if getHash(password) == user.Password {
            return login, 204, nil
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
        user, err := db.DbClient.GetUser(*api.db, lg.Value)
        if err != nil {
            log.Printf("[error] %v", err)
            return "", 401, errors.New("Unauthorized")
        }
        if tk.Value == user.Password {
            return lg.Value, 204, nil
        }
        return "", 403, errors.New("Forbidden")
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

    // Delete old checks
    go func(client db.DbClient){
        for {
            cnt, err := client.DeleteOldChecks()
            if err != nil {
                log.Printf("[error] %v", err)
            } else {
                log.Printf("[info] deleted old checks (%d)", cnt)
            }
            time.Sleep(24 * time.Hour)
        }
    }(client)

    user, _ := client.GetUser(conf.Global.Security.AdminUser)
    user.Id = conf.Global.Security.AdminUser
    user.Password = getHash(conf.Global.Security.AdminPassword)
    user.FullName = conf.Global.Security.AdminUser

    if err := client.SaveUser(user); err != nil {
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

    user, err := db.DbClient.GetUser(*api.db, username)
    if err != nil {
        w.WriteHeader(403)
        w.Write(encodeResp(&Resp{Status:"error", Error:"Login or password is incorrect"}))
    }
    if getHash(password) == user.Password {
        w.WriteHeader(200)
        w.Write(encodeResp(&Resp{Status:"success", Data:user}))
        return
    }

    /*
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
    */

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

    login, code, err := authentication(api, r)
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

        switch res[1] {
            case "cars":
                rows, err := db.DbClient.LoadCars(*api.db, row)
                if err != nil {
                    log.Printf("[error] %v - %s", err, r.URL.Path)
                    w.WriteHeader(500)
                    w.Write(encodeResp(&Resp{Status:"error", Error:err.Error()}))
                    return
                }

                w.WriteHeader(200)
                w.Write(encodeResp(&Resp{Status:"success", Data:rows}))
                return

            case "owners":
                rows, err := db.DbClient.LoadOwners(*api.db, row)
                if err != nil {
                    log.Printf("[error] %v - %s", err, r.URL.Path)
                    w.WriteHeader(500)
                    w.Write(encodeResp(&Resp{Status:"error", Error:err.Error()}))
                    return
                }

                w.WriteHeader(200)
                w.Write(encodeResp(&Resp{Status:"success", Data:rows}))
                return

            case "places":
                rows, err := db.DbClient.LoadPlaces(*api.db, row)
                if err != nil {
                    log.Printf("[error] %v - %s", err, r.URL.Path)
                    w.WriteHeader(500)
                    w.Write(encodeResp(&Resp{Status:"error", Error:err.Error()}))
                    return
                }

                w.WriteHeader(200)
                w.Write(encodeResp(&Resp{Status:"success", Data:rows}))
                return

            case "prices":
                rows, err := db.DbClient.LoadPrices(*api.db, row)
                if err != nil {
                    log.Printf("[error] %v - %s", err, r.URL.Path)
                    w.WriteHeader(500)
                    w.Write(encodeResp(&Resp{Status:"error", Error:err.Error()}))
                    return
                }

                w.WriteHeader(200)
                w.Write(encodeResp(&Resp{Status:"success", Data:rows}))
                return

            case "main":
                rows, err := db.DbClient.LoadMain(*api.db, row)
                if err != nil {
                    log.Printf("[error] %v - %s", err, r.URL.Path)
                    w.WriteHeader(500)
                    w.Write(encodeResp(&Resp{Status:"error", Error:err.Error()}))
                    return
                }

                w.WriteHeader(200)
                w.Write(encodeResp(&Resp{Status:"success", Data:rows}))
                return

            case "users":
                rows, err := db.DbClient.LoadUsers(*api.db, row)
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

        w.WriteHeader(204)
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

        switch res[1] {
            case "cars":
                object := config.Car{}
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
            case "owners":
                object := config.Owner{}
                if err := json.Unmarshal(body, &object); err != nil {
                    log.Printf("[error] %v - %s", err, r.URL.Path)
                    w.WriteHeader(400)
                    w.Write(encodeResp(&Resp{Status:"error", Error:err.Error()}))
                    return
                }

                if err := db.DbClient.SaveOwner(*api.db, object); err != nil {
                    log.Printf("[error] %v - %s", err, r.URL.Path)
                    w.WriteHeader(500)
                    w.Write(encodeResp(&Resp{Status:"error", Error:err.Error()}))
                    return
                }
            case "places":
                object := config.Place{}
                if err := json.Unmarshal(body, &object); err != nil {
                    log.Printf("[error] %v - %s", err, r.URL.Path)
                    w.WriteHeader(400)
                    w.Write(encodeResp(&Resp{Status:"error", Error:err.Error()}))
                    return
                }

                if err := db.DbClient.SavePlace(*api.db, object); err != nil {
                    log.Printf("[error] %v - %s", err, r.URL.Path)
                    w.WriteHeader(500)
                    w.Write(encodeResp(&Resp{Status:"error", Error:err.Error()}))
                    return
                }
            case "prices":
                object := config.Price{}
                if err := json.Unmarshal(body, &object); err != nil {
                    log.Printf("[error] %v - %s", err, r.URL.Path)
                    w.WriteHeader(400)
                    w.Write(encodeResp(&Resp{Status:"error", Error:err.Error()}))
                    return
                }

                if err := db.DbClient.SavePrice(*api.db, object); err != nil {
                    log.Printf("[error] %v - %s", err, r.URL.Path)
                    w.WriteHeader(500)
                    w.Write(encodeResp(&Resp{Status:"error", Error:err.Error()}))
                    return
                }
            case "main":
                object := config.Main{}
                if err := json.Unmarshal(body, &object); err != nil {
                    log.Printf("[error] %v - %s", err, r.URL.Path)
                    w.WriteHeader(400)
                    w.Write(encodeResp(&Resp{Status:"error", Error:err.Error()}))
                    return
                }

                if err := db.DbClient.SaveMain(*api.db, object, login); err != nil {
                    log.Printf("[error] %v - %s", err, r.URL.Path)
                    w.WriteHeader(500)
                    w.Write(encodeResp(&Resp{Status:"error", Error:err.Error()}))
                    return
                }
            
            case "users":
                user := config.User{}
                if err := json.Unmarshal(body, &user); err != nil {
                    log.Printf("[error] %v - %s", err, r.URL.Path)
                    w.WriteHeader(400)
                    w.Write(encodeResp(&Resp{Status:"error", Error:err.Error()}))
                    return
                }

                if user.Password == "" {
                    usr, err := db.DbClient.GetUser(*api.db, user.Id)
                    if err == nil {
                        user.Password = usr.Password
                    }
                } else {
                    user.Password = getHash(user.Password)
                }

                if err := db.DbClient.SaveUser(*api.db, user); err != nil {
                    log.Printf("[error] %v - %s", err, r.URL.Path)
                    w.WriteHeader(500)
                    w.Write(encodeResp(&Resp{Status:"error", Error:err.Error()}))
                    return
                }
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

        switch res[1] {
            case "cars","owners","prices","places","users":
                if err := db.DbClient.DeleteObject(*api.db, res[1], row["id"]); err != nil {
                    log.Printf("[error] %v - %s", err, r.URL.Path)
                    w.WriteHeader(500)
                    w.Write(encodeResp(&Resp{Status:"error", Error:err.Error()}))
                    return
                }
            case "main":
                if err := db.DbClient.DeleteMain(*api.db, row["id"], login); err != nil {
                    log.Printf("[error] %v - %s", err, r.URL.Path)
                    w.WriteHeader(500)
                    w.Write(encodeResp(&Resp{Status:"error", Error:err.Error()}))
                    return
                }
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

    login, code, err := authentication(api, r)
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

        var row map[string]string

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

func (api *Api) ApiCheck(w http.ResponseWriter, r *http.Request) {
    login, code, err := authentication(api, r)
    if err != nil {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(code)
        w.Write(encodeResp(&Resp{Status:"error", Error:err.Error()}))
        return
    }

    if r.Method == "GET" {

        id := int64(0)

        for k, v := range r.URL.Query() {
            switch k {
                case "id":
                    i, err := strconv.Atoi(v[0])
                    if err != nil {
                        w.Header().Set("Content-Type", "application/json")
                        w.WriteHeader(500)
                        w.Write(encodeResp(&Resp{Status:"error", Error:err.Error()}))
                        return
                    }
                    id = int64(i)
            }
        }

        row, err := db.DbClient.LoadCheck(*api.db, id, login)
        if err != nil {
            log.Printf("[error] %v - %s", err, r.URL.Path)
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(500)
            w.Write(encodeResp(&Resp{Status:"error", Error:err.Error()}))
            return
        }

        tmpl, err := template.ParseFiles(api.conf.Global.WebDir+"/check.tmpl")
        if err != nil {
            log.Printf("[error] %v - %s", err, r.URL.Path)
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(500)
            w.Write(encodeResp(&Resp{Status:"error", Error:err.Error()}))
            return
        }

        err = tmpl.Execute(w, row)
        if err != nil {
            log.Printf("[error] %v - %s", err, r.URL.Path)
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(500)
            w.Write(encodeResp(&Resp{Status:"error", Error:err.Error()}))
            return
        }

        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(405)
    w.Write(encodeResp(&Resp{Status:"error", Error:"Method Not Allowed"}))
}
