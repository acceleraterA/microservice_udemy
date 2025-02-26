package main

import (
	event "broker/events"
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/rpc"
)

type RequestPayload struct {
	Action string      `json:"action"`
	Auth   AuthPayload `json:"auth,omitempty"`
	Log    logPayload  `json:"log,omitempty"`
	Mail   MailPayload `json:"mail,omitempty"`
}

type AuthPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type logPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}
type MailPayload struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

func (app *Config) Broker(w http.ResponseWriter, r *http.Request) {
	payload := jsonResponse{
		Error:   false,
		Message: "Broker service is up and running",
	}
	_ = app.writeJSON(w, http.StatusAccepted, payload)

}

func (app *Config) HandleSubmission(w http.ResponseWriter, r *http.Request) {
	var requestPayload RequestPayload
	// read the json to requestPayload
	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}
	switch requestPayload.Action {
	case "auth":
		app.authenticate(w, requestPayload.Auth)
	case "log":
		app.logViaROC(w, requestPayload.Log)
	case "mail":
		app.sendMail(w, requestPayload.Mail)
	default:
		app.errorJSON(w, errors.New("invalid action"), http.StatusBadRequest)
	}

}
func (app *Config) logItem(w http.ResponseWriter, entry logPayload) {
	// create json we'll send to the log service
	jsonData, _ := json.MarshalIndent(entry, "", "\t")
	logServiceURL := "http://logger-service/log"

	// call the log service
	request, err := http.NewRequest("POST", logServiceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}
	defer response.Body.Close() // close the body when we're done
	//make sure get back correct status code
	if response.StatusCode != http.StatusAccepted {
		app.errorJSON(w, errors.New("error calling log service"))
		return
	}
	// create a variable we'll read response.Body into
	var payload jsonResponse
	payload.Message = "Log written to database"
	payload.Error = false
	// decode the json
	app.writeJSON(w, http.StatusAccepted, payload)
}

func (app *Config) authenticate(w http.ResponseWriter, a AuthPayload) {
	// create json we'll send to the auth service
	jsonData, _ := json.MarshalIndent(a, "", "\t")

	// call the auth service
	request, err := http.NewRequest("POST", "http://authentication-service/authenticate", bytes.NewBuffer(jsonData))
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}
	defer response.Body.Close() // close the body when we're done
	//make sure get back correct status code
	if response.StatusCode == http.StatusUnauthorized {
		app.errorJSON(w, errors.New("invalid credentials"), http.StatusBadRequest)
		return
	} else if response.StatusCode != http.StatusAccepted {
		app.errorJSON(w, errors.New("error calling authentication service"), http.StatusInternalServerError)
		return
	}

	// create a variable we'll read response.Body into
	var jsonFromService jsonResponse
	// decode the json
	err = json.NewDecoder(response.Body).Decode(&jsonFromService)
	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}
	// check if the auth service returned an error
	if jsonFromService.Error {
		app.errorJSON(w, err, http.StatusUnauthorized)
		return
	}
	var payload jsonResponse
	payload.Message = jsonFromService.Message
	payload.Data = jsonFromService.Data
	payload.Error = false

	app.writeJSON(w, http.StatusAccepted, payload)
	// read the json to requestPayload

}

func (app *Config) sendMail(w http.ResponseWriter, m MailPayload) {
	// create json we'll send to the mail service
	jsonData, _ := json.MarshalIndent(m, "", "\t")
	log.Printf("JSON data to send: %s", jsonData)
	// call the mail service
	request, err := http.NewRequest("POST", "http://mail-service/send", bytes.NewBuffer(jsonData))
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	request.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Printf("Error making HTTP request: %v", err)
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}
	defer response.Body.Close() // close the body when we're done
	//make sure get back correct status code
	log.Printf("Response status code: %d", response.StatusCode)
	if response.StatusCode != http.StatusAccepted {
		log.Printf("Error calling mail service: %v", response.Status)
		app.errorJSON(w, errors.New("error calling mail service"))
		return
	}
	// create a variable we'll read response.Body into
	var payload jsonResponse
	payload.Message = "Mail sent to" + m.To
	payload.Error = false
	// decode the json
	app.writeJSON(w, http.StatusAccepted, payload)
}

func (app *Config) logEventViaRabbit(w http.ResponseWriter, l logPayload) {
	err := app.pushToQueue(l.Name, l.Data)
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	var payload jsonResponse
	payload.Message = "Log via rmq"
	payload.Error = false
	app.writeJSON(w, http.StatusAccepted, payload)
}

func (app *Config) pushToQueue(name, msg string) error {
	emitter, err := event.NewEventEmitter(app.Rabbit)
	if err != nil {
		return err
	}
	payload := logPayload{
		Name: name,
		Data: msg,
	}
	j, _ := json.MarshalIndent(&payload, "", "\t")
	err = emitter.Push(string(j), "log.INFO")
	if err != nil {
		return err
	}
	return nil
}

type RPCPayload struct {
	Name string
	Data string
}

func (app *Config) logViaROC(w http.ResponseWriter, l logPayload) {
	// get rpc client
	client, err := rpc.Dial("tcp", "logger-service:5001")
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	defer client.Close()

	// create a payload
	rpcPayload := RPCPayload(l)
	var result string
	// call the RPC server with the payload
	err = client.Call("RPCServer.LogINFO", rpcPayload, &result)
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	var payload jsonResponse = jsonResponse{
		Message: result,
		Error:   false,
	}
	app.writeJSON(w, http.StatusAccepted, payload)
}
