package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/rs/zerolog"
	"github.com/spf13/pflag"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
)

func echoIPHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tracer := otel.Tracer(tracerName)
	_, span := tracer.Start(ctx, "echoIPHandler")
	defer span.End()
	remoteIP := ""
	remoteIPs := r.Header.Values("X-Forwarded-For")
	if len(remoteIPs) > 0 {
		remoteIP = remoteIPs[0]
	}
	if remoteIP == "" {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "hostname splittint failed")
			http.Error(w, "Unexpected remote address", http.StatusBadRequest)
			return
		}
		remoteIP = host
	}
	ip := net.ParseIP(remoteIP)
	if ip == nil {
		span.RecordError(fmt.Errorf("invalid IP address"))
		span.SetStatus(codes.Error, "IP could not be parsed")
		http.Error(w, "Invalid IP", http.StatusBadRequest)
		return
	}
	fmt.Fprint(w, ip.String())
}

const tracerName = "mypubip"

func main() {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger()
	var addr string
	var otlpTracing bool
	pflag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of mypubip:\n")
		pflag.PrintDefaults()
	}
	pflag.StringVar(&addr, "addr", "localhost:8000", "Address to listen on")
	pflag.BoolVar(&otlpTracing, "telemetry", false, "Enable telemetry submission via OTLP")
	pflag.Parse()

	if otlpTracing {
		logger.Info().Msg("Telemetry enabled")
		client := otlptracehttp.NewClient(otlptracehttp.WithInsecure())
		exporter, err := otlptrace.New(context.Background(), client)
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to create tracing exporter")
		}
		res, err := resource.Merge(
			resource.Default(),
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String("mypubip"),
			),
		)
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to build telemetry resource")
		}
		tp := trace.NewTracerProvider(trace.WithBatcher(exporter), trace.WithResource(res))
		otel.SetTracerProvider(tp)
	}

	srv := http.Server{}
	srv.Addr = addr
	srv.Handler = http.HandlerFunc(echoIPHandler)
	logger.Info().Msgf("Starting server on %s", addr)
	if err := srv.ListenAndServe(); err != nil {
		logger.Fatal().Err(err).Msg("Failed to start listener.")
	}
}
