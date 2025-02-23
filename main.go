package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func main() {

	fmt.Println("Hello world")

	port := "3000"
	if os.Getenv("port") != "" {
		port = os.Getenv("port")
	}
	http.HandleFunc("/", AppRender)
	http.ListenAndServe("0.0.0.0:"+port, nil)

}

func ReturnJSON(w http.ResponseWriter, data map[string]interface{}, statusCode ...int) {
	w.Header().Set("Content-type", "application/json")
	status := http.StatusBadRequest
	if len(statusCode) > 0 {
		status = statusCode[0]
	}
	content, err := json.Marshal(data)
	if err != nil {
		errorContent, _ := json.Marshal(map[string]string{"error": err.Error()})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(errorContent)
		return
	}
	w.WriteHeader(status)
	w.Write(content)
}

func AppRender(w http.ResponseWriter, r *http.Request) {
	ReturnJSON(w, map[string]interface{}{"message": "Hello world"}, 200)
}
