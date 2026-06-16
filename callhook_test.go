package main

import (
	"context"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func callReq(name string) mcp.CallToolRequest {
	r := mcp.CallToolRequest{}
	r.Params.Name = name
	return r
}

func okHandler(called *bool) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		*called = true
		return &mcp.CallToolResult{Content: []mcp.Content{mcp.NewTextContent("ok")}}, nil
	}
}

func TestCallHook_NotGated_PassesThrough(t *testing.T) {
	called := false
	mw := newCallHookMiddleware("srv", &CallHookConfig{Command: []string{"false"}, RequireFor: []string{"gated"}})
	res, err := mw(okHandler(&called))(context.Background(), callReq("ungated"))
	if err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("ungated tool should pass straight through to next")
	}
	if res.IsError {
		t.Fatal("ungated call should not be an error")
	}
}

func TestCallHook_Approve(t *testing.T) {
	called := false
	mw := newCallHookMiddleware("srv", &CallHookConfig{Command: []string{"true"}, RequireFor: []string{"gated"}})
	res, err := mw(okHandler(&called))(context.Background(), callReq("gated"))
	if err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("exit 0 should approve and reach next")
	}
	if res.IsError {
		t.Fatal("approved call should not be an error")
	}
}

func TestCallHook_Deny(t *testing.T) {
	called := false
	mw := newCallHookMiddleware("srv", &CallHookConfig{
		Command:    []string{"sh", "-c", "echo nope >&2; exit 1"},
		RequireFor: []string{"gated"},
	})
	res, err := mw(okHandler(&called))(context.Background(), callReq("gated"))
	if err != nil {
		t.Fatal(err)
	}
	if called {
		t.Fatal("denied call must NOT reach next")
	}
	if !res.IsError {
		t.Fatal("denied call should be IsError")
	}
	txt := res.Content[0].(mcp.TextContent).Text
	if !strings.Contains(txt, "nope") {
		t.Fatalf("expected denial reason surfaced, got %q", txt)
	}
}

func TestCallHook_Timeout_FailsClosed(t *testing.T) {
	called := false
	mw := newCallHookMiddleware("srv", &CallHookConfig{
		Command:    []string{"sleep", "5"},
		RequireFor: []string{"gated"},
		TimeoutSec: 1,
	})
	res, err := mw(okHandler(&called))(context.Background(), callReq("gated"))
	if err != nil {
		t.Fatal(err)
	}
	if called {
		t.Fatal("timed-out hook must fail closed (deny), not reach next")
	}
	if !res.IsError {
		t.Fatal("timeout should be IsError")
	}
	if txt := res.Content[0].(mcp.TextContent).Text; !strings.Contains(txt, "timed out") {
		t.Fatalf("expected timeout reason, got %q", txt)
	}
}

func TestCallHook_Wildcard_GatesEverything(t *testing.T) {
	called := false
	mw := newCallHookMiddleware("srv", &CallHookConfig{Command: []string{"false"}, RequireFor: []string{"*"}})
	res, err := mw(okHandler(&called))(context.Background(), callReq("any_unlisted_tool"))
	if err != nil {
		t.Fatal(err)
	}
	if called {
		t.Fatal(`"*" should gate every tool; an unlisted tool must hit the hook (and be denied)`)
	}
	if !res.IsError {
		t.Fatal("denied call should be IsError")
	}
}

func TestCallHook_Wildcard_ApproveReachesNext(t *testing.T) {
	called := false
	mw := newCallHookMiddleware("srv", &CallHookConfig{Command: []string{"true"}, RequireFor: []string{"*"}})
	res, err := mw(okHandler(&called))(context.Background(), callReq("any_tool"))
	if err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("exit 0 should approve and reach next even under wildcard")
	}
	if res.IsError {
		t.Fatal("approved call should not be an error")
	}
}
