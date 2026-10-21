package controllers

import (
	"net/http"
	"time"

	"github.com/wibiesana/padi_go_core/realtime"
	"github.com/wibiesana/padi_go_core/response"
	"github.com/wibiesana/padi_go_core/router"
	"github.com/wibiesana/padi_go_core/validator"
)

type ExampleRealtimeController struct{}

func NewExampleRealtimeController() *ExampleRealtimeController {
	return &ExampleRealtimeController{}
}

type ChatBroadcastRequest struct {
	Message  string `json:"message" validate:"required,max=1000"`
	Username string `json:"username" validate:"max=50"`
}

// Subscribe handles client SSE streaming connection
func (c *ExampleRealtimeController) Subscribe(w http.ResponseWriter, r *http.Request) {
	topic := router.QueryParam(r, "topic", "public-chat")
	realtime.SubscribeSSE(topic)(w, r)
}

// Broadcast sends message to SSE subscribers
func (c *ExampleRealtimeController) Broadcast(w http.ResponseWriter, r *http.Request) {
	var req ChatBroadcastRequest
	if errs, err := validator.BindJSON(r, &req); err != nil {
		response.UnprocessableEntity(w, errs, "Validation failed")
		return
	}

	if req.Username == "" {
		req.Username = "Anonymous"
	}

	payload := map[string]interface{}{
		"username": req.Username,
		"message":  req.Message,
		"sent_at":  time.Now().Format(time.RFC3339),
	}

	topic := "public-chat"
	realtime.Publish(topic, payload)

	response.Success(w, map[string]interface{}{
		"topic":   topic,
		"payload": payload,
	}, "Message broadcasted successfully to SSE subscribers")
}
