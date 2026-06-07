package dto

import (
	"time"
)

const (
	MESSAGE_SUCCESS_GET_LOG_SOCKET    = "success get log socket"
	MESSAGE_FAILED_GET_LOG_SOCKET     = "failed get log socket"
	MESSAGE_SUCCESS_CREATE_LOG_SOCKET = "success create log socket"
	MESSAGE_FAILED_CREATE_LOG_SOCKET  = "failed create log socket"
)

type LogSocketResponse struct {
	LogID     uint      `json:"log_id"`
	Payload   string    `json:"payload"`
	CreatedAt time.Time `json:"created_at"`
}

type LogSocketCreateRequest struct {
	Payload string `json:"payload" binding:"required"`
}
