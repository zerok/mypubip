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
	resp, err := c.Get(s.URL)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, "127.0.0.1", string(data))
}
