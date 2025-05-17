package initializers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Wrapper struct {
	Writer  http.ResponseWriter
	Request *http.Request
	Data    map[string]any
}

const DateFormat = "2006-01-02 15:04:05"

var locParis = time.FixedZone("CEST", 2*60*60)

func StringPtr(s string) *string {
	return &s
}

func (wrapper *Wrapper) Render(data map[string]any, status ...int) {
	wrapper.Writer.Header().Set("Content-type", "application/json")
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
	wrapper.Writer.WriteHeader(code)
	dataJSON, err := json.Marshal(response)
	if err != nil {
		wrapper.Error(err.Error())
		return
	}
	wrapper.Writer.Write(dataJSON)
}

func (wrapper *Wrapper) Error(error string, code ...int) {
	wrapper.Writer.Header().Set("Content-type", "application/json")
	statusCode := 400
	if len(code) > 0 {
		statusCode = code[0]
	}
	dataJSON, _ := json.Marshal(map[string]string{
		"error": error,
	})
	wrapper.Writer.WriteHeader(statusCode)
	wrapper.Writer.Write(dataJSON)
}

func (wrapper *Wrapper) WrapData(data string) error {
	if wrapper.Data[data] == nil || wrapper.Data[data] == "" {
		return fmt.Errorf("you have to set a value for the key 	'%v'", data)
	}
	return nil
}

func (wrapper *Wrapper) HandlePOST(r *http.Request) (errorMSG string, errorCode int) {
	if r.Method != http.MethodPost {
		return "Not authorized", http.StatusMethodNotAllowed
	}
	if err := wrapper.Request.ParseMultipartForm(10 >> 20); err != nil {
		return err.Error(), http.StatusBadGateway
	}
	wrapper.Data = make(map[string]interface{})
	for key, values := range wrapper.Request.MultipartForm.Value {
		if len(values) <= 0 {
			continue
		}
		wrapper.Data[key] = values[0]
	}
	if len(wrapper.Data) <= 0 {
		return "No data received", http.StatusBadGateway
	}
	return "", 0
}

func WrapFormat(dateStr *string) (string, error) {
	parsed, err := time.ParseInLocation(DateFormat, *dateStr, time.UTC)
	if err != nil {
		return "", err
	}
	return parsed.In(locParis).Format(DateFormat), nil
}

func NewWrapper(w http.ResponseWriter, r *http.Request) *Wrapper {
	return &Wrapper{
		Writer:  w,
		Request: r,
	}
}

func (wrapper *Wrapper) ReturnUser() int {
	return wrapper.Request.Context().Value("user").(int)
}
