package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const defaultCallHookTimeout = 120 * time.Second

// newCallHookMiddleware returns a tool-handler middleware that gates the tools
// listed in cfg.RequireFor behind an external command. Tools not in the list
// pass straight through; a single "*" entry gates every tool (default-deny
// posture, with the hook as the allowlist). For gated tools the command is run with the tool-call
// request marshaled as JSON on stdin and MCP_SERVER/MCP_TOOL in the
// environment. Exit code 0 approves the call; any non-zero exit, spawn error,
// or timeout denies it (fail-closed).
//
// A denial is returned as a CallToolResult with IsError set (not a Go error),
// so the model receives a readable reason as tool output and can adapt, rather
// than seeing an opaque transport failure.
func newCallHookMiddleware(serverName string, cfg *CallHookConfig) server.ToolHandlerMiddleware {
	gateAll := false
	gated := make(map[string]struct{}, len(cfg.RequireFor))
	for _, name := range cfg.RequireFor {
		if name == "*" {
			gateAll = true
			continue
		}
		gated[name] = struct{}{}
	}
	timeout := defaultCallHookTimeout
	if cfg.TimeoutSec > 0 {
		timeout = time.Duration(cfg.TimeoutSec) * time.Second
	}

	return func(next server.ToolHandlerFunc) server.ToolHandlerFunc {
		return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			if !gateAll {
				if _, ok := gated[req.Params.Name]; !ok {
					return next(ctx, req)
				}
			}

			payload, err := json.Marshal(req.Params)
			if err != nil {
				return denied(serverName, req.Params.Name, "could not encode request: "+err.Error()), nil
			}

			hookCtx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			cmd := exec.CommandContext(hookCtx, cfg.Command[0], cfg.Command[1:]...)
			cmd.Stdin = bytes.NewReader(payload)
			cmd.Env = append(os.Environ(),
				"MCP_SERVER="+serverName,
				"MCP_TOOL="+req.Params.Name,
			)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			if err := cmd.Run(); err != nil {
				reason := firstNonEmpty(
					strings.TrimSpace(stderr.String()),
					strings.TrimSpace(stdout.String()),
					err.Error(),
				)
				if hookCtx.Err() == context.DeadlineExceeded {
					reason = "approval timed out after " + timeout.String()
				}
				log.Printf("<%s> call hook DENIED tool %s: %s", serverName, req.Params.Name, reason)
				return denied(serverName, req.Params.Name, reason), nil
			}

			log.Printf("<%s> call hook approved tool %s", serverName, req.Params.Name)
			return next(ctx, req)
		}
	}
}

func denied(serverName, tool, reason string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{
			mcp.NewTextContent("Tool call '" + tool + "' was denied by the operator: " + reason),
		},
	}
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
