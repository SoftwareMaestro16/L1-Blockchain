package observability

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"
)

const MetricsPath = "/metrics"

func Handler(registry *Registry) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc(MetricsPath, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		if err := registry.WritePrometheus(w); err != nil {
			http.Error(w, "failed to render metrics", http.StatusInternalServerError)
		}
	})
	return mux
}

func StartMetricsServer(ctx context.Context, addr string, registry *Registry) error {
	if addr == "" {
		return nil
	}
	if registry == nil {
		registry = DefaultRegistry
	}
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	server := &http.Server{
		Handler:		Handler(registry),
		ReadHeaderTimeout:	3 * time.Second,
	}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 3*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()
	go func() {
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			_ = listener.Close()
		}
	}()
	return nil
}
