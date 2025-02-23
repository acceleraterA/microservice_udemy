package main

import (
	"log"
	"net/http"
)

func (app *Config) SendMail(w http.ResponseWriter, r *http.Request) {
	type mailMessage struct {
		FROM    string `json:"from"`
		TO      string `json:"to"`
		Subject string `json:"subject"`
		Message string `json:"message"`
	}
	// read the json to requestPayload
	var requestPayload mailMessage
	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		log.Println(err)
		app.errorJSON(w, err)
		return
	}

	msg := Message{
		From:    requestPayload.FROM,
		To:      requestPayload.TO,
		Subject: requestPayload.Subject,
		Data:    requestPayload.Message,
	}
	// send the mail
	err = app.Mailer.SendSMTPMessage(msg)
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	log.Println(err)
	// send response
	payload := jsonResponse{
		Error:   false,
		Message: "Sent to" + requestPayload.TO,
	}
	app.writeJSON(w, http.StatusAccepted, payload)
}
