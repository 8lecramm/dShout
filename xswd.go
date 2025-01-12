package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/gorilla/websocket"
)

type XSWD struct {
	connection *websocket.Conn
	active     bool
	address    url.URL
	AppInfo    *AppicationInfo
	response   chan []byte
}

type XSWD_Auth_Response struct {
	Accepted bool   `json:"accepted"`
	Message  string `json:"message"`
}

type AppicationInfo struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Url         string `json:"url"`
}

var xswd *XSWD

func XSWD_Init() *XSWD {

	xswd := XSWD{
		connection: nil,
		AppInfo:    nil,
		address: url.URL{
			Scheme: "ws",
			Host:   "localhost:44326",
			Path:   "/xswd",
		},
		response: make(chan []byte),
	}

	return &xswd
}

func (x *XSWD) XSWD_SetServer(server string) {
	x.address.Scheme = "ws"
	x.address.Host = fmt.Sprintf("%s", server)
	x.address.Path = "/xswd"
}

func (x *XSWD) XSWD_Connect() error {
	log_xswd.Println("> Connect")
	if c, _, err := websocket.DefaultDialer.Dial(x.address.String(), nil); err != nil {
		return err
	} else {
		x.connection = c
	}

	if err := x.xswd_authorize(); err != nil {
		return err
	}
	go x.xswd_read_loop()

	return nil
}

func (x *XSWD) XSWD_Exit() {
	x.active = false
	x.connection.Close()
	log_xswd.Println("> Shutdown")
}

func (x *XSWD) xswd_authorize() error {

	log_xswd.Println("> Authorization")

	hash := sha256.Sum256([]byte(x.AppInfo.Name))
	x.AppInfo.Id = hex.EncodeToString(hash[:])

	data, err := json.Marshal(x.AppInfo)
	if err != nil {
		return err
	}

	if err = x.connection.WriteMessage(websocket.TextMessage, data); err != nil {
		return err
	}

	_, buffer, err := x.connection.ReadMessage()
	if err != nil {
		return err
	}

	var auth_response XSWD_Auth_Response
	if err = json.Unmarshal(buffer, &auth_response); err != nil {
		return err
	}
	if !auth_response.Accepted {
		return fmt.Errorf("authorization failed")
	}
	x.active = true

	log_xswd.Println(auth_response.Message)

	return nil
}

func (x *XSWD) xswd_read_loop() {

	for x.active {
		msg_type, buffer, err := x.connection.ReadMessage()
		if err != nil {
			continue
		}
		if msg_type != websocket.TextMessage {
			continue
		}
		x.response <- buffer
	}
}

func (x *XSWD) xswd_send(data []byte) bool {
	if err := x.connection.WriteMessage(websocket.TextMessage, data); err != nil {
		return false
	}

	return true
}

func xswd_response(b []byte, t any) error {

	var temp JSONRPCResponse

	if err := json.Unmarshal(b, &temp); err != nil {
		return err
	}
	if temp.Result == nil {
		return fmt.Errorf("no permission or invalid result")
	}
	data, err := json.Marshal(temp.Result)
	if err = json.Unmarshal(data, &t); err != nil {
		return err
	}

	return nil
}
