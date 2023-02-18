package main

import (
	"os"
	"log"
	"flag"
	"net/http"
    "path/filepath"
	"github.com/kardianos/service"
	"gopkg.in/natefinch/lumberjack.v2"
    "github.com/ltkh/parking/internal/config"
    "github.com/ltkh/parking/internal/api/v1"
    "github.com/ltkh/parking/internal/migration"
)

var logger service.Logger

type program struct{}

func (p *program) Start(s service.Service) error {
	// Start should not block. Do the actual work async.
	go p.run()
	return nil
}

func (p *program) run() {
    // Parking path
    ex, err := os.Executable()
    if err != nil {
        log.Fatalf("[error] %v", err)
    }
    exPath := filepath.Dir(ex)

	// Command-line flag parsing
    cfFile         := flag.String("config.file", exPath+"/parking.yml", "config file")
    webDir         := flag.String("web.dir", exPath+"/web", "web directory")
    lgFile         := flag.String("log.file", exPath+"/parking.log", "log file")
    mdbFile        := flag.String("mdb.file", "", "mdb file")
    flag.Parse()
	
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

    // Loading configuration file
    cfg, err := config.New(*cfFile)
    if err != nil {
        log.Fatalf("[error] %v", err)
    }

    if cfg.Global.WebDir == "" {
        cfg.Global.WebDir = *webDir
    }

    // Start migration
    if *mdbFile != "" {
        if err := migration.Start(*mdbFile, cfg.DB); err != nil {
            log.Fatalf("[error] %v", err)
        }
        os.Exit(0)
    }

    // Creating api
    apiV1, err := v1.New(cfg)
    if err != nil {
        log.Fatalf("[error] %v", err)
    }

    http.HandleFunc("/api/v1/login", apiV1.ApiLogin)
    http.HandleFunc("/api/v1/ws", apiV1.WsEndpoint)
    http.HandleFunc("/api/v1/update", apiV1.ApiUpdate)
    http.HandleFunc("/api/v1/cars", apiV1.ApiObjects)
    http.HandleFunc("/api/v1/owners", apiV1.ApiObjects)
    http.HandleFunc("/api/v1/places", apiV1.ApiObjects)
    http.HandleFunc("/api/v1/prices", apiV1.ApiObjects)
    http.HandleFunc("/api/v1/checks", apiV1.ApiObjects)
    http.HandleFunc("/api/v1/check", apiV1.ApiCheck)
    http.HandleFunc("/api/v1/main", apiV1.ApiObjects)
    http.HandleFunc("/api/v1/parking", apiV1.ApiParking)
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request){
        http.ServeFile(w, r, cfg.Global.WebDir+"/"+r.URL.Path)
    })

    // Enabled listen port
    if err := http.ListenAndServe(cfg.Global.ListenAddr, nil); err != nil {
        log.Fatalf("[error] %v", err)
    }
}

func (p *program) Stop(s service.Service) error {
	// Stop should not block. Return with a few seconds.
	return nil
}

func main() {
	svcConfig := &service.Config{
		Name:        "Parking",
		DisplayName: "Parking",
		Description: "",
	}

	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}
	logger, err = s.Logger(nil)
	if err != nil {
		log.Fatal(err)
	}
	err = s.Run()
	if err != nil {
		logger.Error(err)
	}
}