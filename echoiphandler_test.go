package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEchoIPHandler(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(echoIPHandler))
	c := http.Client{}

	var expectIP = func(t *testing.T, ip string, req *http.Request) {
		t.Helper()
		resp, err := c.Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		defer resp.Body.Close()
		data, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, ip, string(data))
	}

	t.Run("no headers", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, s.URL, nil)
		expectIP(t, "127.0.0.1", req)
	})
	t.Run("x-forwarded-for", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, s.URL, nil)
		req.Header.Add("X-Forwarded-For", "127.0.0.2")
		expectIP(t, "127.0.0.2", req)
	})
	t.Run("x-forwarded-for-invalid", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, s.URL, nil)
		req.Header.Add("X-Forwarded-For", "abc")
		resp, err := c.Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}
