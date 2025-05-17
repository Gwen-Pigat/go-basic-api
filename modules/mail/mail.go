package mail

import (
	"net/http"
	"oristack/initializers"
	"os"

	"gopkg.in/gomail.v2"
)

func SendMail(wrapper *initializers.Wrapper) {
	keys := []string{"from", "to", "subject", "content"}
	data := make(map[string]any)
	for _, key := range keys {
		if err := wrapper.WrapData(key); err != nil {
			wrapper.Error(err.Error(), http.StatusBadRequest)
			return
		}
		data[key] = wrapper.Data[key]
	}
	m := gomail.NewMessage()
	m.SetHeader("From", wrapper.Data["from"].(string))
	m.SetHeader("To", wrapper.Data["to"].(string))
	m.SetHeader("Subject", wrapper.Data["subject"].(string))
	m.SetBody("text/html", wrapper.Data["content"].(string))

	d := gomail.NewDialer(os.Getenv("MAIL_HOST"), 465, os.Getenv("MAIL_USERNAME"), os.Getenv("MAIL_PASSWORD"))
	if err := d.DialAndSend(m); err != nil {
		wrapper.Error(err.Error(), http.StatusBadGateway)
		return
	}
	wrapper.Render(data)
}
