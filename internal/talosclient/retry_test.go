package talosclient

import (
	"context"
	"errors"
	"net"
	"syscall"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestIsTransientError(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"deadline exceeded", context.DeadlineExceeded, true},
		{"canceled is not transient", context.Canceled, false},
		{"grpc Unavailable", status.Error(codes.Unavailable, "boom"), true},
		{"grpc ResourceExhausted", status.Error(codes.ResourceExhausted, "boom"), true},
		{"grpc DeadlineExceeded", status.Error(codes.DeadlineExceeded, "boom"), true},
		{"grpc NotFound is not transient", status.Error(codes.NotFound, "nope"), false},
		{"grpc InvalidArgument is not transient", status.Error(codes.InvalidArgument, "nope"), false},
		{"grpc PermissionDenied is not transient", status.Error(codes.PermissionDenied, "nope"), false},
		{"net timeout", &net.OpError{Op: "dial", Err: timeoutErr{}}, true},
		{"ECONNREFUSED", syscall.ECONNREFUSED, true},
		{"ECONNRESET", syscall.ECONNRESET, true},
		{"ETIMEDOUT", syscall.ETIMEDOUT, true},
		{"EPIPE", syscall.EPIPE, true},
		{"EACCES is not transient", syscall.EACCES, false},
		{"connection refused string", errors.New("dial tcp: connection refused"), true},
		{"eof string", errors.New("rpc error: stream EOF"), true},
		{"i/o timeout string", errors.New("read tcp: i/o timeout"), true},
		{"unrelated error", errors.New("invalid resource type"), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsTransientError(tc.err); got != tc.want {
				t.Fatalf("IsTransientError(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}

type timeoutErr struct{}

func (timeoutErr) Error() string   { return "timeout" }
func (timeoutErr) Timeout() bool   { return true }
func (timeoutErr) Temporary() bool { return true }

func TestIsTransientError_WrappedDeadline(t *testing.T) {
	wrapped := errors.Join(errors.New("outer"), context.DeadlineExceeded)
	if !IsTransientError(wrapped) {
		t.Fatalf("expected wrapped DeadlineExceeded to be transient")
	}
}

func TestIsTransientError_AvoidsFalsePositives(t *testing.T) {
	for _, msg := range []string{
		"resource not found",
		"unauthorized",
		"invalid token",
	} {
		if IsTransientError(errors.New(msg)) {
			t.Fatalf("did not expect %q to be transient", msg)
		}
	}
}
