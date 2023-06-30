package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
)

// These methods and structs are used only for sending information to the frontend. But not for getting it

type WebSocketConnection struct {
	*websocket.Conn
}

type WebSocketPayload struct {
	Action              string              `json:"action"`
	Message             string              `json:"message"`
	UserName            string              `json:"userName"`
	MessageType         string              `json:"messageType"`
	WebSocketConnection WebSocketConnection `json:"-"`
}

type WebSocketResponse struct {
	Action              string              `json:"action"`
	Message             string              `json:"message"`
	UserID              string              `json:"userId"`
	MessageType         string              `json:"messageType"`
	WebSocketConnection WebSocketConnection `json:"-"`
}

var websocketUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true }, // Right now this method is not used, but if it needed
	// to be connected from browser to the server, it should be implemented and used for the security purposes
}

var clients = make(map[WebSocketConnection]string)

var webSocketChannel = make(chan WebSocketPayload)

// WebSocketEndPoint is a handler which handles connection from the frontend
func (application *application) WebSocketEndPoint(w http.ResponseWriter, r *http.Request) {
	// ws is a pointer to connection with websocket, which comes from the frontend request
	ws, err := websocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		application.errorLog.Println("cannot upgrade websocket connection ", err)
		return
	}

	application.infoLog.Println("client connected to ws")

	var resp WebSocketResponse
	resp.Message = "You are now connected to server via websocket"

	err = ws.WriteJSON(resp)
	if err != nil {
		application.errorLog.Println("cannot upgrade websocket connection ", err)
		return
	}

	conn := WebSocketConnection{Conn: ws}
	clients[conn] = ""

	application.ListenForWebSocket(&conn)
}

func (application application) ListenForWebSocket(conn *WebSocketConnection) {
	defer func() {
		if r := recover(); r != nil {
			application.errorLog.Println("Error:", fmt.Sprintf("%v", r))
		}
	}()

	var payload WebSocketPayload

	// this section listens for payload from frontend, and if it comes, gives this data to channel
	for {
		err := conn.ReadJSON(&payload)
		if err != nil {
			// noting
		} else {
			payload.WebSocketConnection = *conn
			webSocketChannel <- payload
		}
	}
}

// ListenToWebsocketChannel works with webSocketChannel, and when data comes into it, calls broadcastToAll method
// This method starts in main function, when app starts
func (application *application) ListenToWebsocketChannel() {
	var response WebSocketResponse

	for {
		event := <-webSocketChannel

		switch event.Action {
		case "deleteUser":
			response.Action = "logout"
			response.Message = "Your Account has been deleted"

			application.broadcastToAll(response)

		default:
		}
	}
}

func (application *application) broadcastToAll(response WebSocketResponse) {
	for client := range clients {
		err := client.WriteJSON(response)
		if err != nil {
			application.errorLog.Printf("Websocket error on %s, %s", response.Action, err)
			_ = client.Close()
			delete(clients, client)
		}
	}
}
