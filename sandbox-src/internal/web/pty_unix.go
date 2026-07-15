//go:build !windows

package web

import (
	"log"
	"os"
	"os/exec"
	"sync"

	"github.com/creack/pty"
	"github.com/gorilla/websocket"
)

// startPTYSession drives a full WebSocket ↔ PTY session on macOS/Linux.
// Returns true if the session was handled, false to fall back to pipes.
func startPTYSession(conn *websocket.Conn, sandboxDir, shellPath string, shellArgs, env []string) bool {
	cmd := exec.Command(shellPath, shellArgs...)
	cmd.Env = env
	cmd.Dir, _ = os.Getwd()

	ptmx, err := pty.StartWithSize(cmd, &pty.Winsize{Rows: 24, Cols: 80})
	if err != nil {
		log.Printf("pty start: %v", err)
		return false
	}
	defer ptmx.Close()

	log.Printf("PTY session started: %s %v", shellPath, shellArgs)

	var wg sync.WaitGroup

	// PTY output → WebSocket
	wg.Add(1)
	go func() {
		defer wg.Done()
		buf := make([]byte, 4096)
		for {
			n, err := ptmx.Read(buf)
			if n > 0 {
				if wErr := conn.WriteMessage(websocket.BinaryMessage, buf[:n]); wErr != nil {
					break
				}
			}
			if err != nil {
				break
			}
		}
	}()

	// WebSocket → PTY input
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				break
			}
			if len(msg) > 0 && msg[0] == '{' {
				resizeFn := func(cols, rows int) {
					_ = pty.Setsize(ptmx, &pty.Winsize{
						Rows: uint16(rows),
						Cols: uint16(cols),
					})
				}
				if handleControlMsg(msg, ptmx, resizeFn) {
					continue
				}
			}
			if _, err := ptmx.Write(msg); err != nil {
				break
			}
		}
	}()

	wg.Wait()

	if cmd.Process != nil {
		_ = cmd.Process.Signal(os.Interrupt)
		_ = cmd.Wait()
	}

	return true
}
