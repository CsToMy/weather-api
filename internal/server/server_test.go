package server

import (
	"context"
	"errors"
	"net"
	"net/http"
	"testing"
	"time"
)

func newListener(t *testing.T) net.Listener {
	t.Helper()

	ln, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", "127.0.0.1:0")

	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	return ln
}

func get(ctx context.Context, url string) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)

	if err != nil {
		return 0, err
	}

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return 0, err
	}

	defer func() { _ = resp.Body.Close() }()

	return resp.StatusCode, nil
}

func startRun(ctx context.Context, srv *http.Server, ln net.Listener, timeout time.Duration) <-chan error {
	done := make(chan error, 1)

	go func() { done <- Run(ctx, srv, ln, timeout) }()

	return done
}

func waitFor(t *testing.T, ch <-chan struct{}, what string) {
	t.Helper()

	select {
	case <-ch:
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for %s", what)
	}
}

func TestRunStopsOnContextCancel(t *testing.T) {
	ln := newListener(t)

	url := "http://" + ln.Addr().String()

	srv := &http.Server{
		Handler:           http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
		ReadHeaderTimeout: time.Second,
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := startRun(ctx, srv, ln, time.Second)

	status, err := get(context.Background(), url)

	if err != nil || status != http.StatusOK {
		t.Fatalf("before shutdown: status=%d err=%v", status, err)
	}

	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run returned an error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return after the context was cancelled")
	}

	if _, err := get(context.Background(), url); err == nil {
		t.Fatal("server still accepts requests after shutdown")
	}
}

func TestRunWaitsForInFlightRequest(t *testing.T) {
	ln := newListener(t)

	url := "http://" + ln.Addr().String()

	started := make(chan struct{})
	release := make(chan struct{})

	srv := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			close(started)
			<-release
		}),
		ReadHeaderTimeout: time.Second,
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := startRun(ctx, srv, ln, 5*time.Second)

	type result struct {
		status int
		err    error
	}

	results := make(chan result, 1)

	go func() {
		status, err := get(context.Background(), url)
		results <- result{status, err}
	}()

	waitFor(t, started, "the request to reach the handler")
	cancel()

	select {
	case <-done:
		t.Fatal("Run returned while a request was still in flight")
	case <-time.After(100 * time.Millisecond):
	}

	close(release)

	res := <-results
	if res.err != nil || res.status != http.StatusOK {
		t.Fatalf("in-flight request: status=%d err=%v", res.status, res.err)
	}

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run returned an error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return after the request finished")
	}
}

func TestRunForcesCloseAfterTimeout(t *testing.T) {
	ln := newListener(t)

	url := "http://" + ln.Addr().String()

	started := make(chan struct{})
	release := make(chan struct{})

	t.Cleanup(func() { close(release) })

	srv := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			close(started)
			select {
			case <-release:
			case <-r.Context().Done():
			}
		}),
		ReadHeaderTimeout: time.Second,
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := startRun(ctx, srv, ln, 50*time.Millisecond)

	go func() { _, _ = get(context.Background(), url) }()

	waitFor(t, started, "the request to reach the handler")
	cancel()

	select {
	case err := <-done:
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("error = %v, want context deadline exceeded", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not give up after the shutdown timeout")
	}
}

func TestRunReturnsServeError(t *testing.T) {
	ln := newListener(t)
	_ = ln.Close()

	srv := &http.Server{ReadHeaderTimeout: time.Second}

	err := Run(context.Background(), srv, ln, time.Second)

	if err == nil {
		t.Fatal("expected an error from a closed listener")
	}
}
