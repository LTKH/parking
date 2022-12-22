package main 

import (
	"log"
	"net/http"
	"github.com/webview/webview"
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

	w := webview.New(true)
	defer w.Destroy()
	w.SetTitle("Parking")
	w.SetSize(800, 600, webview.HintNone)
	w.Navigate("http://localhost:8800/index.html")
	w.Run()
}