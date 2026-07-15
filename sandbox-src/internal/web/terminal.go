package web

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Only allow connections from localhost
		host := r.Host
		return host == "127.0.0.1" || host == "localhost" ||
			len(host) > 10 && (host[:10] == "127.0.0.1:" || host[:10] == "localhost:")
	},
}

// controlMsg is a JSON message from the frontend for resize, etc.
type controlMsg struct {
	Type string `json:"type"`
	Cols int    `json:"cols"`
	Rows int    `json:"rows"`
	Data string `json:"data"` // For base64 image data
}

// handleControlMsg processes messages like resize or image paste.
func handleControlMsg(msg []byte, stdin io.Writer, ptyResize func(cols, rows int)) bool {
	var ctrl controlMsg
	if json.Unmarshal(msg, &ctrl) != nil || ctrl.Type == "" {
		return false
	}

	switch ctrl.Type {
	case "resize":
		if ptyResize != nil {
			ptyResize(ctrl.Cols, ctrl.Rows)
		}
		return true
	case "image-paste":
		handleImagePaste(ctrl.Data, stdin)
		return true
	}
	return false
}

// maxImageBase64 limits incoming base64 image data to ~10 MB (≈7.5 MB decoded).
const maxImageBase64 = 10 * 1024 * 1024

func handleImagePaste(base64Data string, stdin io.Writer) {
	if base64Data == "" {
		return
	}

	if len(base64Data) > maxImageBase64 {
		log.Println("image-paste: data too large, ignoring")
		return
	}

	// Remove data:image/...;base64, prefix if present
	if idx := strings.Index(base64Data, ","); idx != -1 {
		base64Data = base64Data[idx+1:]
	}

	data, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		log.Printf("base64 decode: %v", err)
		return
	}

	workspaceDir := "workspace"
	_ = os.MkdirAll(workspaceDir, 0755)
	fileName := fmt.Sprintf("pasted_%d.png", time.Now().UnixMilli())
	filePath := filepath.Join(workspaceDir, fileName)

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		log.Printf("save image: %v", err)
		return
	}

	log.Printf("Image saved to %s (%d bytes)", filePath, len(data))

	// Type the command into the shell
	cmd := fmt.Sprintf("agy \"請分析這張圖片內容: ./%s\"\n", filePath)
	if _, err := io.WriteString(stdin, cmd); err != nil {
		log.Printf("write to stdin: %v", err)
	}
}

// TerminalHandler creates a WebSocket handler that connects to a shell process.
// On Windows it uses the winpty DLL for real PTY emulation; elsewhere (or as
// fallback) it bridges pipes.
func TerminalHandler(sandboxDir, shellPath string, shellArgs []string, env []string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("websocket upgrade: %v", err)
			return
		}
		defer conn.Close()

		// Try native PTY first (winpty DLL on Windows).
		if startPTYSession(conn, sandboxDir, shellPath, shellArgs, env) {
			return
		}

		// Fallback: pipe-based terminal.
		log.Println("PTY not available, using pipe-based terminal")
		startPipeSession(conn, shellPath, shellArgs, env)
	}
}

// startPipeSession runs the shell with stdin/stdout pipes (no TTY).
func startPipeSession(conn *websocket.Conn, shellPath string, shellArgs []string, env []string) {
	log.Printf("Launching (pipes): %s %v", shellPath, shellArgs)

	cmd := exec.Command(shellPath, shellArgs...)
	cmd.Env = env
	setSysProcAttr(cmd)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Printf("stdin pipe: %v", err)
		return
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("stdout pipe: %v", err)
		return
	}

	cmd.Stderr = cmd.Stdout // merge stderr into stdout

	if err := cmd.Start(); err != nil {
		log.Printf("start shell: %v", err)
		return
	}

	var wg sync.WaitGroup

	// stdout → websocket
	wg.Add(1)
	go func() {
		defer wg.Done()
		buf := make([]byte, 4096)
		for {
			n, err := stdout.Read(buf)
			if n > 0 {
				data := addCarriageReturns(buf[:n])
				if writeErr := conn.WriteMessage(websocket.BinaryMessage, data); writeErr != nil {
					break
				}
			}
			if err != nil {
				break
			}
		}
	}()

	// websocket → stdin (filter out control messages)
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer stdin.Close()
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				break
			}
			if len(msg) > 0 && msg[0] == '{' {
				if handleControlMsg(msg, stdin, nil) {
					continue
				}
			}
			if _, err := io.WriteString(stdin, string(msg)); err != nil {
				break
			}
		}
	}()

	wg.Wait()

	if cmd.Process != nil {
		_ = cmd.Process.Signal(os.Interrupt)
		_ = cmd.Wait()
	}
}

// addCarriageReturns converts bare \n to \r\n for xterm.js.
// Already-correct \r\n sequences are left untouched.
func addCarriageReturns(data []byte) []byte {
	count := 0
	for i, b := range data {
		if b == '\n' && (i == 0 || data[i-1] != '\r') {
			count++
		}
	}
	if count == 0 {
		return data
	}
	out := make([]byte, 0, len(data)+count)
	for i, b := range data {
		if b == '\n' && (i == 0 || data[i-1] != '\r') {
			out = append(out, '\r')
		}
		out = append(out, b)
	}
	return out
}
