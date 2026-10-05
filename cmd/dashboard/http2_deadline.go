package main

import (
	"context"
	"net"
	"net/http"
	"time"
)

type dashboardConnKey struct{}

func dashboardConnectionContext(ctx context.Context, conn net.Conn) context.Context {
	return context.WithValue(ctx, dashboardConnKey{}, conn)
}

// Go's native h2c handoff can retain the HTTP/1 header deadline when
// ReadHeaderTimeout is set without ReadTimeout. Clear it only after complete
// HTTP/2 headers arrive, so long-lived agent streams survive the header limit.
func clearHTTP2HeaderDeadline(r *http.Request) {
	if r.ProtoMajor != 2 {
		return
	}
	if conn, ok := r.Context().Value(dashboardConnKey{}).(net.Conn); ok {
		_ = conn.SetReadDeadline(time.Time{})
	}
}
