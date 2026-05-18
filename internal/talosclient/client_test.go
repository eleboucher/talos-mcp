package talosclient

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/siderolabs/talos/pkg/machinery/client"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func newTestClient(build func(ctx context.Context) (*client.Client, error)) *Client {
	return &Client{build: build}
}

func TestRun_Success(t *testing.T) {
	c := newTestClient(func(context.Context) (*client.Client, error) { return nil, nil })

	calls := 0
	err := c.Run(context.Background(), func(_ *client.Client) error {
		calls++
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 1 {
		t.Fatalf("calls = %d, want 1", calls)
	}
}

func TestRun_PermanentErrorDoesNotRetry(t *testing.T) {
	c := newTestClient(func(context.Context) (*client.Client, error) { return nil, nil })

	calls := 0
	want := status.Error(codes.InvalidArgument, "bad request")
	err := c.Run(context.Background(), func(_ *client.Client) error {
		calls++
		return want
	})
	if err == nil || status.Code(err) != codes.InvalidArgument {
		t.Fatalf("got err = %v, want code %v", err, codes.InvalidArgument)
	}
	if calls != 1 {
		t.Fatalf("calls = %d, want 1", calls)
	}
}

func TestRun_TransientErrorRefreshesAndRetries(t *testing.T) {
	var refreshes atomic.Int32
	c := newTestClient(func(context.Context) (*client.Client, error) {
		refreshes.Add(1)
		return nil, nil
	})

	var calls atomic.Int32
	err := c.Run(context.Background(), func(_ *client.Client) error {
		if calls.Add(1) < 3 {
			return status.Error(codes.Unavailable, "flaky")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error after retry: %v", err)
	}
	if got := calls.Load(); got < 3 {
		t.Fatalf("expected at least 3 op invocations, got %d", got)
	}
	if got := refreshes.Load(); got < 2 {
		t.Fatalf("expected at least 2 refreshes, got %d", got)
	}
}

func TestRun_RefreshFailureWrapsBothErrors(t *testing.T) {
	c := newTestClient(func(context.Context) (*client.Client, error) {
		return nil, errors.New("can't rebuild")
	})

	ctx, cancel := context.WithCancel(context.Background())
	go cancel()
	err := c.Run(ctx, func(_ *client.Client) error {
		return status.Error(codes.Unavailable, "flake")
	})
	if err == nil {
		t.Fatalf("expected an error when refresh keeps failing")
	}
}

func TestRun_ConcurrentRefreshSafe(t *testing.T) {
	c := newTestClient(func(context.Context) (*client.Client, error) { return nil, nil })

	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			_ = c.Run(context.Background(), func(_ *client.Client) error { return nil })
		})
	}
	wg.Wait()
}

func TestClient_CloseIsIdempotent(t *testing.T) {
	c := newTestClient(func(context.Context) (*client.Client, error) { return nil, nil })
	if err := c.Close(); err != nil {
		t.Fatalf("first close: %v", err)
	}
	if err := c.Close(); err != nil {
		t.Fatalf("second close: %v", err)
	}
}
