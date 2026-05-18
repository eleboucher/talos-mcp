package talosclient

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/siderolabs/go-retry/retry"
	"github.com/siderolabs/talos/pkg/machinery/client"
)

type Factory struct{}

func NewFactory() (*Factory, error) {
	return &Factory{}, nil
}

type Override struct {
	Context   string
	Endpoints []string
	Nodes     []string
}

type Client struct {
	mu    sync.Mutex
	inner *client.Client
	build func(ctx context.Context) (*client.Client, error)
}

type CallCtx = context.Context

func (f *Factory) New(ctx context.Context, ov Override) (*Client, CallCtx, error) {
	inner, err := newClient(ctx, ov)
	if err != nil {
		return nil, nil, err
	}
	c := &Client{
		inner: inner,
		build: func(ctx context.Context) (*client.Client, error) { return newClient(ctx, ov) },
	}
	callCtx := ctx
	if len(ov.Nodes) > 0 {
		callCtx = client.WithNodes(ctx, ov.Nodes...)
	}
	return c, callCtx, nil
}

// NewSingleNode targets exactly one node via WithNode. COSI calls don't
// support one-to-many proxying so they must use this constructor.
func (f *Factory) NewSingleNode(ctx context.Context, ov Override) (*Client, CallCtx, error) {
	inner, err := newClient(ctx, ov)
	if err != nil {
		return nil, nil, err
	}
	c := &Client{
		inner: inner,
		build: func(ctx context.Context) (*client.Client, error) { return newClient(ctx, ov) },
	}
	callCtx := ctx
	if len(ov.Nodes) > 0 {
		callCtx = client.WithNode(ctx, ov.Nodes[0])
	}
	return c, callCtx, nil
}

func newClient(ctx context.Context, ov Override) (*client.Client, error) {
	opts := []client.OptionFunc{client.WithDefaultConfig()}
	if ov.Context != "" {
		opts = append(opts, client.WithContextName(ov.Context))
	}
	if len(ov.Endpoints) > 0 {
		opts = append(opts, client.WithEndpoints(ov.Endpoints...))
	}
	c, err := client.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("new talos client: %w", err)
	}
	return c, nil
}

func (c *Client) Talos() *client.Client {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.inner
}

func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.inner == nil {
		return nil
	}
	err := c.inner.Close()
	c.inner = nil
	return err
}

func (c *Client) Run(ctx context.Context, op func(tc *client.Client) error) error {
	return retry.Constant(10*time.Second, retry.WithUnits(100*time.Millisecond)).Retry(func() error {
		err := op(c.Talos())
		if err == nil {
			return nil
		}
		if !IsTransientError(err) {
			return err
		}
		if rerr := c.refresh(ctx); rerr != nil {
			return retry.ExpectedError(fmt.Errorf("refresh client: %w (original: %v)", rerr, err))
		}
		return retry.ExpectedError(err)
	})
}

func (c *Client) refresh(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	next, err := c.build(ctx)
	if err != nil {
		return err
	}
	if c.inner != nil {
		_ = c.inner.Close()
	}
	c.inner = next
	return nil
}
