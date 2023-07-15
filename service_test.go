package slcansvc

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-kit/log"
	"github.com/stretchr/testify/assert"
)

func TestHTTP(t *testing.T) {
	svc := NewService()
	mux := MakeHTTPHandler(svc, log.NewNopLogger())
	srv := httptest.NewServer(mux)
	defer srv.Close()

	msg := Message{
		ID:   123,
		Data: "200rpm",
	}
	jsonMsg, _ := json.Marshal(msg)
	req, _ := http.NewRequest("POST", srv.URL+"/slcan", bytes.NewBuffer(jsonMsg))
	resp, _ := http.DefaultClient.Do(req)
	body, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "{}", strings.TrimSpace(string(body)))

	var want getMessageResponse
	req, _ = http.NewRequest("GET", srv.URL+"/slcan/123", nil)
	resp, _ = http.DefaultClient.Do(req)
	body, _ = ioutil.ReadAll(resp.Body)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	err := json.Unmarshal(body, &want)
	assert.NoError(t, err)
	assert.Equal(t, msg, want.Msg)

	msg.Data = "201rpm"
	jsonMsg, _ = json.Marshal(msg)
	req, _ = http.NewRequest("PUT", srv.URL+"/slcan/123", bytes.NewBuffer(jsonMsg))
	resp, _ = http.DefaultClient.Do(req)
	body, _ = ioutil.ReadAll(resp.Body)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "{}", strings.TrimSpace(string(body)))

	req, _ = http.NewRequest("GET", srv.URL+"/slcan/123", nil)
	resp, _ = http.DefaultClient.Do(req)
	body, _ = ioutil.ReadAll(resp.Body)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	err = json.Unmarshal(body, &want)
	assert.NoError(t, err)
	assert.Equal(t, msg, want.Msg)

	jsonMsg, _ = json.Marshal(msg)
	req, _ = http.NewRequest("DELETE", srv.URL+"/slcan/123", bytes.NewBuffer(jsonMsg))
	resp, _ = http.DefaultClient.Do(req)
	body, _ = ioutil.ReadAll(resp.Body)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "{}", strings.TrimSpace(string(body)))

	// req, _ = http.NewRequest("GET", srv.URL+"/slcan/123", nil)
	// resp, _ = http.DefaultClient.Do(req)
	// body, _ = ioutil.ReadAll(resp.Body)
	// assert.Equal(t, http.StatusOK, resp.StatusCode)
	// err = json.Unmarshal(body, &want)
	// assert.NoError(t, err)
	// assert.Equal(t, Message{}, want.Msg)
}
