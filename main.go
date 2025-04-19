package main

import (
	"chi-api/initializers"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

var db *sql.DB

type Wrapper struct {
	writer  http.ResponseWriter
	request *http.Request
	data    map[string]any
}

func init() {
	godotenv.Load()
	var err error
	db, err = initializers.ConnectDB()
	if err != nil {
		panic(err)
	}
	fmt.Println(db)
}

func main() {
	port := "3000"
	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	}
	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{os.Getenv("API_URL_PREPROD"), os.Getenv("API_URL_PROD")},
		AllowedMethods:   []string{"GET", "POST", "PATCH", "OPTIONS", "DELETE"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-type", "X-CSRF-Token"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
	Handle(r, http.MethodGet, "/db/flush", FlushDB)
	Handle(r, http.MethodPost, "/user/connect", GetUserConnect)
	Handle(r, http.MethodPost, "/user", CreateUser)
	r.Group(func(r chi.Router) {
		r.Use(CheckAuth)
		Handle(r, http.MethodGet, "/user", GetUser)
		Handle(r, http.MethodGet, "/tasks", GetTasks)
		Handle(r, http.MethodPost, "/tasks", CreateTask)
		Handle(r, http.MethodPatch, "/tasks/{id}", PatchTask)
		Handle(r, http.MethodDelete, "/tasks/{id}", DeleteTask)
	})
	http.ListenAndServe(":"+port, r)
}

func FlushDB(wrapper *Wrapper) {
	err := initializers.ExecFlushDB(db)
	if err != nil {
		wrapper.Error(err.Error())
		return
	}
	wrapper.Render(map[string]any{
		"message": "DB is flushed",
	})
}

func CheckAuth(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" {
			NewWrapper(w, r).Error("Not authorized", http.StatusUnauthorized)
			return
		}
		wrapper := NewWrapper(w, r)
		wrapper.data = make(map[string]any)
		wrapper.data["token"] = auth
		userID, err := GetUserAuth(wrapper)
		if err != nil {
			NewWrapper(w, r).Error("Not authorized : "+err.Error(), http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(wrapper.request.Context(), "user", userID)
		handler.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (wrapper *Wrapper) ReturnUser() int {
	return wrapper.request.Context().Value("user").(int)
}

func Handle(r chi.Router, method string, path string, handler func(w *Wrapper)) {
	r.MethodFunc(method, path, func(w http.ResponseWriter, r *http.Request) {
		wrapper := NewWrapper(w, r)
		if method == http.MethodPost {
			errorMsg, errorCode := wrapper.HandlePOST(wrapper.request)
			if errorMsg != "" {
				wrapper.Error(errorMsg, errorCode)
				return
			}
		}
		handler(wrapper)
	})
}

func (wrapper *Wrapper) HandlePOST(r *http.Request) (errorMSG string, errorCode int) {
	if r.Method != http.MethodPost {
		return "Not authorized", http.StatusMethodNotAllowed
	}
	if err := wrapper.request.ParseMultipartForm(10 >> 20); err != nil {
		return err.Error(), http.StatusBadGateway
	}
	wrapper.data = make(map[string]interface{})
	for key, values := range wrapper.request.MultipartForm.Value {
		if len(values) <= 0 {
			continue
		}
		wrapper.data[key] = values[0]
	}
	if len(wrapper.data) <= 0 {
		return "No data received", http.StatusBadGateway
	}
	return "", 0
}

func HandleGET(r *http.Request) (errorMSG string, errorCode int) {
	if r.Method != http.MethodGet {
		return "Not authorized", http.StatusMethodNotAllowed
	}
	return "", 0
}

func (wrapper Wrapper) Render(data map[string]any, status ...int) {
	wrapper.writer.Header().Set("Content-type", "application/json")
	code := http.StatusOK
	if len(status) > 0 {
		code = status[0]
	}
	var response any
	if payload, ok := data["data"]; ok {
		response = payload
	} else {
		response = data
	}
	wrapper.writer.WriteHeader(code)
	dataJSON, err := json.Marshal(response)
	if err != nil {
		wrapper.Error(err.Error())
		return
	}
	wrapper.writer.Write(dataJSON)
}

func (wrapper Wrapper) Error(error string, code ...int) {
	wrapper.writer.Header().Set("Content-type", "application/json")
	statusCode := 400
	if len(code) > 0 {
		statusCode = code[0]
	}
	dataJSON, _ := json.Marshal(map[string]string{
		"error": error,
	})
	wrapper.writer.WriteHeader(statusCode)
	wrapper.writer.Write(dataJSON)
}

func NewWrapper(w http.ResponseWriter, r *http.Request) *Wrapper {
	return &Wrapper{
		writer:  w,
		request: r,
	}
}

func (wrapper *Wrapper) wrapData(data string) error {
	if wrapper.data[data] == nil || wrapper.data[data] == "" {
		return fmt.Errorf("you have to set a value for %v", data)
	}
	return nil
}

func Index(wrapper *Wrapper) {
	wrapper.Render(map[string]any{
		"message": "Hello world",
	})
}
