package main 

import (
	"log"
	"net/http"
	//"github.com/webview/webview"
	//"github.com/virteman/webview"
	//"github.com/zserge/webview"
	"github.com/NiuStar/webview"
)

func main() {
    go func(){
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request){
			http.ServeFile(w, r, "web"+r.URL.Path)
		})
		// Enabled listen port
		if err := http.ListenAndServe("127.0.0.1:8800", nil); err != nil {
			log.Fatalf("[error] %v", err)
		}
	}()

	//w := webview.New(true)
	//defer w.Destroy()
	//w.ClearCache()
	//w.SetTitle("Parking")
	//w.SetSize(1400, 1000, webview.HintNone)
	//w.Navigate("http://localhost:8800/index.html")
	//w.Run()

	w := webview.New(webview.Settings{
		Width:  800,
		Height: 600,
		Title:  "Simple canvas demo",
		URL:    "http://localhost:8800/index.html",
	})
	defer w.Exit()
	//
	w.Run()
}