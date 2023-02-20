package main 

import (
	"log"
	//"os/exec"
	"github.com/webview/webview"
	"github.com/ltkh/parking/internal/config"
)

func main() {
	// Loading configuration file
    cfg, err := config.New("parking.yml")
    if err != nil {
        log.Fatalf("[error] %v", err)
    }

	//cmnd := exec.Command("parking", "")
	//cmnd.Run()

	//go func(){
	//cmnd := exec.Command("./parking")
	//cmnd.Start()
	//cmnd.Wait()
	//}()

	/*
    go func(){
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request){
			http.ServeFile(w, r, "web"+r.URL.Path)
		})
		// Enabled listen port
		if err := http.ListenAndServe("127.0.0.1:8800", nil); err != nil {
			log.Fatalf("[error] %v", err)
		}
	}()
	*/

	w := webview.New(true)
	defer w.Destroy()
	//w.ClearCache()
	w.SetTitle("Parking")
	w.SetSize(cfg.Window.Width, cfg.Window.Height, webview.HintNone)
	w.Navigate(cfg.Window.Navigate)
	w.Run()
}