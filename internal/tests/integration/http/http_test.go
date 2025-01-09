package http_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"server/internal/db"
	"server/internal/server"
)

func TestCliHandler(t *testing.T) {
	router := gin.Default()
	diceClient := &db.DiceDB{}
	httpServer := server.NewHTTPServer(router, nil, diceClient, 0, 0)

	router.POST("/cli", gin.WrapF(httpServer.CliHandler))

	// Test case: valid command (currently failing)
	reqBody := `{"Cmd": "set", "Args": {"k1": "v1"}}`
	req, err := http.NewRequest("POST", "/cli", strings.NewReader(reqBody))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	expectedBody := `{"data":"result of the command"}`
	assert.JSONEq(t, expectedBody, w.Body.String())

	// Test case: invalid command
	reqBody = `{"Cmd": "invalid", "Args": {}}`
	req, err = http.NewRequest("POST", "/cli", strings.NewReader(reqBody))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	expectedBody = `{"error":"json: cannot unmarshal object into Go value of type []interface {}"}`
	assert.JSONEq(t, expectedBody, w.Body.String())

	// Test case: malformed JSON
	reqBody = `{"Cmd": "invalid", "Args": {"k1": 6`
	req, err = http.NewRequest("POST", "/cli", strings.NewReader(reqBody))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	expectedBody = `{"error":"unexpected end of JSON input"}`
	assert.JSONEq(t, expectedBody, w.Body.String())
}
