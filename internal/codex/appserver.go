package codex

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"time"
)

var errNotInitialized = errors.New("codex app-server not initialized")

const (
	appServerInitTimeout    = 3 * time.Second
	appServerRequestTimeout = 4 * time.Second
)

type appServerClient struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	reader *bufio.Reader
	mu     sync.Mutex
	ch     chan rpcMessage
}

type rpcMessage struct {
	ID     json.RawMessage `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
	Result json.RawMessage `json:"result"`
	Error  *rpcError       `json:"error"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func newAppServerClient(ctx context.Context) (*appServerClient, error) {
	cmd := exec.CommandContext(ctx, "codex", "app-server")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("codex app-server stdin: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("codex app-server stdout: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("codex app-server start: %w", err)
	}

	client := &appServerClient{
		cmd:    cmd,
		stdin:  stdin,
		stdout: stdout,
		reader: bufio.NewReader(stdout),
		ch:     make(chan rpcMessage, 16),
	}

	go client.readLoop()

	if err := client.initialize(ctx); err != nil {
		client.Close()
		return nil, err
	}

	return client, nil
}

func (c *appServerClient) readLoop() {
	for {
		line, err := c.reader.ReadBytes('\n')
		if err != nil {
			return
		}

		line = bytesTrimSpace(line)
		if len(line) == 0 {
			continue
		}

		var msg rpcMessage
		if err := json.Unmarshal(line, &msg); err != nil {
			continue
		}

		c.ch <- msg
	}
}

func (c *appServerClient) initialize(ctx context.Context) error {
	initCtx, cancel := context.WithTimeout(ctx, appServerInitTimeout)
	defer cancel()

	params := map[string]any{
		"clientInfo": map[string]string{
			"name":    "ccstats",
			"version": "0.0.0",
		},
	}

	if err := c.sendRequest(initCtx, 1, "initialize", params, nil); err != nil {
		return err
	}

	return c.sendNotification("initialized", nil)
}

func (c *appServerClient) sendNotification(method string, params any) error {
	note := map[string]any{
		"method": method,
		"params": params,
	}

	payload, err := json.Marshal(note)
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	_, err = c.stdin.Write(append(payload, '\n'))
	return err
}

func (c *appServerClient) sendRequest(ctx context.Context, id int, method string, params any, out any) error {
	req := map[string]any{
		"id":     id,
		"method": method,
		"params": params,
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return err
	}

	c.mu.Lock()
	_, err = c.stdin.Write(append(payload, '\n'))
	c.mu.Unlock()
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg := <-c.ch:
			if msg.Error != nil {
				if msg.Error.Message == "Not initialized" {
					return errNotInitialized
				}
				return fmt.Errorf("codex app-server error: %s", msg.Error.Message)
			}

			if !idMatches(msg.ID, id) {
				continue
			}

			if out == nil {
				return nil
			}

			if err := json.Unmarshal(msg.Result, out); err != nil {
				return fmt.Errorf("codex app-server parse: %w", err)
			}
			return nil
		}
	}
}

func (c *appServerClient) Close() {
	_ = c.stdin.Close()
	_ = c.stdout.Close()
	_ = c.cmd.Process.Kill()
	_, _ = c.cmd.Process.Wait()
}

func idMatches(raw json.RawMessage, id int) bool {
	if len(raw) == 0 {
		return false
	}

	var intID int
	if err := json.Unmarshal(raw, &intID); err == nil {
		return intID == id
	}

	var strID string
	if err := json.Unmarshal(raw, &strID); err == nil {
		return strID == fmt.Sprintf("%d", id)
	}

	return false
}

func bytesTrimSpace(b []byte) []byte {
	start := 0
	end := len(b)
	for start < end && (b[start] == ' ' || b[start] == '\n' || b[start] == '\r' || b[start] == '\t') {
		start++
	}
	for end > start && (b[end-1] == ' ' || b[end-1] == '\n' || b[end-1] == '\r' || b[end-1] == '\t') {
		end--
	}
	return b[start:end]
}
