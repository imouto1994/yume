package http

import (
	"bytes"
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

type ErrorBody struct {
	Error string `json:"error"`
}

func RespondJSON(writer http.ResponseWriter, statusCode int, body interface{}) {
	writer.Header().Set("Content-Type", "application/json")

	b := new(bytes.Buffer)
	if err := json.NewEncoder(b).Encode(body); err != nil {
		RespondInternalServerError(writer, "Could not encode response to JSON", err)
		return
	} else {
		writer.WriteHeader(statusCode)
	}
	_, _ = writer.Write(b.Bytes())
}

func RespondNotFound(writer http.ResponseWriter, message string, err error) {
	zap.L().Warn(message, zap.Error(err))
	RespondJSON(writer, http.StatusNotFound, &ErrorBody{Error: message})
}

func RespondBadRequest(writer http.ResponseWriter, message string, err error) {
	zap.L().Warn(message, zap.Error(err))
	RespondJSON(writer, http.StatusBadRequest, &ErrorBody{Error: message})
}

func RespondInternalServerError(writer http.ResponseWriter, message string, err error) {
	zap.L().Error(message, zap.Error(err))
	RespondJSON(writer, http.StatusInternalServerError, &ErrorBody{Error: message})
}
