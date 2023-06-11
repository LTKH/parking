package main

import (
    "os"
    "log"
    //"fmt"
    "flag"
    "time"
    //"runtime"
    //"os/exec"
    "net/http"
    "github.com/pkg/browser"
    //"github.com/webview/webview"
    "gopkg.in/natefinch/lumberjack.v2"
    "github.com/ltkh/parking/internal/config"
    "github.com/ltkh/parking/internal/api/v1"
    "github.com/ltkh/parking/internal/migration"
)

var (
    cfFile  = flag.String("config.file", "parking.yml", "config file")
    lgFile  = flag.String("log.file", "", "log file")
    webDir  = flag.String("web.dir", "web", "web directory")
    mdbFile = flag.String("mdb.file", "", "mdb file (for migration)")
)

/*
func openBrowser(url string) {
	var err error

	switch runtime.GOOS {
        case "linux":
            err = exec.Command("xdg-open", url).Start()
        case "windows":
            err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
        case "darwin":
            err = exec.Command("open", url).Start()
        default:
            err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal("[error] %v", err)
	}
}
*/

func main() {
    // Command-line flag parsing
    flag.Parse()

    // Loading configuration file
    cfg, err := config.New(*cfFile)
    if err != nil {
        log.Fatalf("[error] %v", err)
    }

    // Start migration
    if *mdbFile != "" {
        if err := migration.Start(*mdbFile, cfg.DB); err != nil {
            log.Fatalf("[error] %v", err)
        }
        os.Exit(0)
    }

    // Logging settings
    if *lgFile != "" {
        log.SetOutput(&lumberjack.Logger{
            Filename:   *lgFile,
            MaxSize:    10,      // megabytes after which new file is created
            MaxBackups: 3,       // number of backups
            MaxAge:     10,      // days
            Compress:   true,    // using gzip
        })
    }

    if cfg.Global.WebDir == "" {
        cfg.Global.WebDir = *webDir
    }

    // Creating api
    apiV1, err := v1.New(cfg)
    if err != nil {
        log.Fatalf("[error] %v", err)
    }

    http.HandleFunc("/api/v1/login", apiV1.ApiLogin)
    //http.HandleFunc("/api/v1/ws", apiV1.WsEndpoint)
    //http.HandleFunc("/api/v1/update", apiV1.ApiUpdate)
    http.HandleFunc("/api/v1/cars", apiV1.ApiObjects)
    http.HandleFunc("/api/v1/owners", apiV1.ApiObjects)
    http.HandleFunc("/api/v1/places", apiV1.ApiObjects)
    http.HandleFunc("/api/v1/prices", apiV1.ApiObjects)
    http.HandleFunc("/api/v1/checks", apiV1.ApiObjects)
    http.HandleFunc("/api/v1/check", apiV1.ApiCheck)
    http.HandleFunc("/api/v1/users", apiV1.ApiObjects)
    http.HandleFunc("/api/v1/main", apiV1.ApiObjects)
    http.HandleFunc("/api/v1/parking", apiV1.ApiParking)
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request){
        http.ServeFile(w, r, cfg.Global.WebDir+"/"+r.URL.Path)
    })

    go func() {
        // Enabled listen port
        if err := http.ListenAndServe(cfg.Global.ListenAddr, nil); err != nil {
            log.Fatalf("[error] %v", err)
        }
    }()

    //w := webview.New(false)
    //defer w.Destroy()
    //w.SetTitle("Parking")
    //w.SetSize(cfg.Window.Width, cfg.Window.Height, webview.HintNone)
    //w.Navigate(cfg.Window.Navigate)
    //w.Run()

    //openBrowser(cfg.Window.Navigate)
    browser.OpenURL(cfg.Window.Navigate)

    // Daemon mode
    for {
        time.Sleep(600 * time.Second)
    }

}