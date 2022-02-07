package main

import (
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/rs/zerolog"
	"github.com/spf13/pflag"
)

func echoIPHandler(w http.ResponseWriter, r *http.Request) {
	remoteIP := ""
	remoteIPs := r.Header.Values("X-Forwarded-For")
	if len(remoteIPs) > 0 {
		remoteIP = remoteIPs[0]
	}
	if remoteIP == "" {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			http.Error(w, "Unexpected remote address", http.StatusBadRequest)
			return
		}
		remoteIP = host
	}
	ip := net.ParseIP(remoteIP)
	if ip == nil {
		http.Error(w, "Invalid IP", http.StatusBadRequest)
		return
	}
	fmt.Fprint(w, ip.String())
}

func main() {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger()
	var addr string
	pflag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of mypubip:\n")
		pflag.PrintDefaults()
	}
	pflag.StringVar(&addr, "addr", "localhost:8000", "Address to listen on")
	pflag.Parse()
	srv := http.Server{}
	srv.Addr = addr
	srv.Handler = http.HandlerFunc(echoIPHandler)
	logger.Info().Msgf("Starting server on %s", addr)
	if err := srv.ListenAndServe(); err != nil {
		logger.Fatal().Err(err).Msg("Failed to start listener.")
	}
}
