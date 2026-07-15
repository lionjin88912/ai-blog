//go:build windows

package web

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/gorilla/websocket"
)

// winpty DLL function pointers (loaded once).
var (
	wpDLL          *syscall.DLL
	wpInit         sync.Once
	wpInitErr      error
	wpConfigNew    *syscall.Proc
	wpConfigFree   *syscall.Proc
	wpConfigSetSz  *syscall.Proc
	wpOpen         *syscall.Proc
	wpConInName    *syscall.Proc
	wpConOutName   *syscall.Proc
	wpSpawnCfgNew  *syscall.Proc
	wpSpawnCfgFree *syscall.Proc
	wpSpawn        *syscall.Proc
	wpSetSize      *syscall.Proc
	wpFree         *syscall.Proc
	wpErrFree      *syscall.Proc
	wpErrMsg       *syscall.Proc
)

func initWinPTY(sandboxDir string) error {
	wpInit.Do(func() {
		dllPath := filepath.Join(sandboxDir, "git", "usr", "bin", "winpty.dll")
		if _, err := os.Stat(dllPath); err != nil {
			wpInitErr = fmt.Errorf("winpty.dll not found: %s", dllPath)
			return
		}

		// Ensure DLL dependencies (msys-2.0.dll etc.) are findable.
		dllDir := filepath.Dir(dllPath)
		os.Setenv("PATH", dllDir+string(os.PathListSeparator)+os.Getenv("PATH"))

		dll, err := syscall.LoadDLL(dllPath)
		if err != nil {
			wpInitErr = fmt.Errorf("load winpty.dll: %w", err)
			return
		}
		wpDLL = dll

		wpConfigNew, _ = dll.FindProc("winpty_config_new")
		wpConfigFree, _ = dll.FindProc("winpty_config_free")
		wpConfigSetSz, _ = dll.FindProc("winpty_config_set_initial_size")
		wpOpen, _ = dll.FindProc("winpty_open")
		wpConInName, _ = dll.FindProc("winpty_conin_name")
		wpConOutName, _ = dll.FindProc("winpty_conout_name")
		wpSpawnCfgNew, _ = dll.FindProc("winpty_spawn_config_new")
		wpSpawnCfgFree, _ = dll.FindProc("winpty_spawn_config_free")
		wpSpawn, _ = dll.FindProc("winpty_spawn")
		wpSetSize, _ = dll.FindProc("winpty_set_size")
		wpFree, _ = dll.FindProc("winpty_free")
		wpErrFree, _ = dll.FindProc("winpty_error_free")
		wpErrMsg, _ = dll.FindProc("winpty_error_msg")

		if wpConfigNew == nil || wpOpen == nil || wpSpawn == nil ||
			wpConInName == nil || wpConOutName == nil || wpSpawnCfgNew == nil {
			wpInitErr = fmt.Errorf("winpty.dll missing required exports")
			wpDLL = nil
		}
	})
	return wpInitErr
}

// winPTY wraps a single winpty instance with named-pipe I/O handles.
type winPTY struct {
	wp      uintptr
	conIn   syscall.Handle
	conOut  syscall.Handle
	process syscall.Handle
}

func openWinPTY(sandboxDir string, cols, rows int) (*winPTY, error) {
	if err := initWinPTY(sandboxDir); err != nil {
		return nil, err
	}

	var errPtr uintptr

	cfg, _, _ := wpConfigNew.Call(0, uintptr(unsafe.Pointer(&errPtr)))
	if cfg == 0 {
		return nil, fmt.Errorf("winpty_config_new: %s", getWPError(errPtr))
	}
	defer wpConfigFree.Call(cfg)

	if wpConfigSetSz != nil {
		wpConfigSetSz.Call(cfg, uintptr(cols), uintptr(rows))
	}

	wp, _, _ := wpOpen.Call(cfg, uintptr(unsafe.Pointer(&errPtr)))
	if wp == 0 {
		return nil, fmt.Errorf("winpty_open: %s", getWPError(errPtr))
	}

	conInPtr, _, _ := wpConInName.Call(wp)
	conOutPtr, _, _ := wpConOutName.Call(wp)

	conInName := utf16PtrToString(conInPtr)
	conOutName := utf16PtrToString(conOutPtr)
	log.Printf("winpty pipes: in=%s out=%s", conInName, conOutName)

	conIn, err := openNamedPipe(conInName, syscall.GENERIC_WRITE)
	if err != nil {
		wpFree.Call(wp)
		return nil, fmt.Errorf("open conin: %w", err)
	}

	conOut, err := openNamedPipe(conOutName, syscall.GENERIC_READ)
	if err != nil {
		syscall.CloseHandle(conIn)
		wpFree.Call(wp)
		return nil, fmt.Errorf("open conout: %w", err)
	}

	return &winPTY{wp: wp, conIn: conIn, conOut: conOut}, nil
}

func (w *winPTY) spawn(cmdline, cwd string, env []string) error {
	var errPtr uintptr

	cmdUTF16, _ := syscall.UTF16PtrFromString(cmdline)

	var cwdUTF16 *uint16
	if cwd != "" {
		cwdUTF16, _ = syscall.UTF16PtrFromString(cwd)
	}

	// Build double-null-terminated UTF-16 environment block.
	env = deduplicateEnv(env)
	envSlice := buildEnvBlock(env)
	var envPtr *uint16
	if len(envSlice) > 0 {
		envPtr = &envSlice[0]
	}

	const winptySpawnFlagAutoShutdown = 1

	spawnCfg, _, _ := wpSpawnCfgNew.Call(
		uintptr(winptySpawnFlagAutoShutdown),
		0, // appname (nil – use cmdline)
		uintptr(unsafe.Pointer(cmdUTF16)),
		uintptr(unsafe.Pointer(cwdUTF16)),
		uintptr(unsafe.Pointer(envPtr)),
		uintptr(unsafe.Pointer(&errPtr)),
	)
	runtime.KeepAlive(envSlice)
	runtime.KeepAlive(cmdUTF16)
	runtime.KeepAlive(cwdUTF16)

	if spawnCfg == 0 {
		return fmt.Errorf("winpty_spawn_config_new: %s", getWPError(errPtr))
	}
	defer wpSpawnCfgFree.Call(spawnCfg)

	var processHandle, threadHandle uintptr
	var createErr uint32

	ret, _, _ := wpSpawn.Call(
		w.wp,
		spawnCfg,
		uintptr(unsafe.Pointer(&processHandle)),
		uintptr(unsafe.Pointer(&threadHandle)),
		uintptr(unsafe.Pointer(&createErr)),
		uintptr(unsafe.Pointer(&errPtr)),
	)
	if ret == 0 {
		return fmt.Errorf("winpty_spawn (err %d): %s", createErr, getWPError(errPtr))
	}

	w.process = syscall.Handle(processHandle)
	if threadHandle != 0 {
		syscall.CloseHandle(syscall.Handle(threadHandle))
	}
	return nil
}

func (w *winPTY) Read(buf []byte) (int, error) {
	var n uint32
	err := syscall.ReadFile(w.conOut, buf, &n, nil)
	return int(n), err
}

func (w *winPTY) Write(buf []byte) (int, error) {
	var n uint32
	err := syscall.WriteFile(w.conIn, buf, &n, nil)
	return int(n), err
}

func (w *winPTY) resize(cols, rows int) {
	if wpSetSize != nil && w.wp != 0 {
		var errPtr uintptr
		wpSetSize.Call(w.wp, uintptr(cols), uintptr(rows), uintptr(unsafe.Pointer(&errPtr)))
	}
}

func (w *winPTY) closeInput() {
	if w.conIn != 0 {
		syscall.CloseHandle(w.conIn)
		w.conIn = 0
	}
}

func (w *winPTY) close() {
	if w.conIn != 0 {
		syscall.CloseHandle(w.conIn)
	}
	if w.conOut != 0 {
		syscall.CloseHandle(w.conOut)
	}
	if w.process != 0 {
		syscall.CloseHandle(w.process)
	}
	if w.wp != 0 {
		wpFree.Call(w.wp)
	}
}

// startPTYSession drives a full WebSocket ↔ winpty session.
// Returns true if the session was handled, false to fall back to pipes.
func startPTYSession(conn *websocket.Conn, sandboxDir, shellPath string, shellArgs, env []string) bool {
	pty, err := openWinPTY(sandboxDir, 80, 24)
	if err != nil {
		log.Printf("winpty DLL: %v", err)
		return false
	}

	cmdline := buildCmdLine(shellPath, shellArgs)
	cwd, _ := os.Getwd()
	log.Printf("winpty spawn: %s (cwd: %s)", cmdline, cwd)

	if err := pty.spawn(cmdline, cwd, env); err != nil {
		log.Printf("winpty spawn failed: %v", err)
		pty.close()
		return false
	}

	log.Println("winpty session started")

	var wg sync.WaitGroup

	// PTY output → WebSocket
	wg.Add(1)
	go func() {
		defer wg.Done()
		buf := make([]byte, 4096)
		lastLog := time.Now()
		totalBytes := 0
		for {
			n, err := pty.Read(buf)
			if n > 0 {
				totalBytes += n
				// Log every 2 seconds or when there's new data after a pause
				now := time.Now()
				if now.Sub(lastLog) > 2*time.Second {
					// Show a preview of the data (first 200 bytes, printable only)
					preview := string(buf[:n])
					if len(preview) > 200 {
						preview = preview[:200]
					}
					log.Printf("[pty→ws] %d bytes (total %d): %q", n, totalBytes, preview)
					lastLog = now
				}
				if wErr := conn.WriteMessage(websocket.BinaryMessage, buf[:n]); wErr != nil {
					log.Printf("[pty→ws] write error: %v", wErr)
					break
				}
			}
			if err != nil {
				log.Printf("[pty→ws] read ended: %v (total %d bytes)", err, totalBytes)
				break
			}
		}
	}()

	// WebSocket → PTY input
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer pty.closeInput()
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				break
			}
			if len(msg) > 0 && msg[0] == '{' {
				if handleControlMsg(msg, pty, pty.resize) {
					continue
				}
			}
			if _, err := pty.Write(msg); err != nil {
				break
			}
		}
	}()

	wg.Wait()
	pty.close()
	return true
}

// --- helpers ---

func buildCmdLine(path string, args []string) string {
	parts := make([]string, 0, 1+len(args))
	if strings.Contains(path, " ") {
		parts = append(parts, `"`+path+`"`)
	} else {
		parts = append(parts, path)
	}
	for _, a := range args {
		if strings.Contains(a, " ") {
			parts = append(parts, `"`+a+`"`)
		} else {
			parts = append(parts, a)
		}
	}
	return strings.Join(parts, " ")
}

// buildEnvBlock creates a double-null-terminated UTF-16 environment block.
func buildEnvBlock(env []string) []uint16 {
	if len(env) == 0 {
		return nil
	}
	var block []uint16
	for _, e := range env {
		utf16, _ := syscall.UTF16FromString(e)
		block = append(block, utf16...) // includes null terminator
	}
	block = append(block, 0) // double-null termination
	return block
}

// deduplicateEnv keeps the last value for each key (case-insensitive).
func deduplicateEnv(env []string) []string {
	seen := make(map[string]int)
	for i, e := range env {
		if idx := strings.IndexByte(e, '='); idx > 0 {
			seen[strings.ToLower(e[:idx])] = i
		}
	}
	result := make([]string, 0, len(seen))
	for i, e := range env {
		idx := strings.IndexByte(e, '=')
		if idx > 0 {
			if seen[strings.ToLower(e[:idx])] == i {
				result = append(result, e)
			}
		} else {
			result = append(result, e)
		}
	}
	return result
}

func openNamedPipe(name string, access uint32) (syscall.Handle, error) {
	nameUTF16, _ := syscall.UTF16PtrFromString(name)
	return syscall.CreateFile(
		nameUTF16, access,
		0, nil,
		syscall.OPEN_EXISTING,
		0, 0,
	)
}

func utf16PtrToString(ptr uintptr) string {
	if ptr == 0 {
		return ""
	}
	var chars []uint16
	for i := 0; ; i++ {
		c := *(*uint16)(unsafe.Pointer(ptr + uintptr(i*2)))
		if c == 0 {
			break
		}
		chars = append(chars, c)
	}
	return syscall.UTF16ToString(chars)
}

func getWPError(errPtr uintptr) string {
	if errPtr == 0 || wpErrMsg == nil {
		return "unknown error"
	}
	msgPtr, _, _ := wpErrMsg.Call(errPtr)
	msg := "unknown error"
	if msgPtr != 0 {
		msg = utf16PtrToString(msgPtr)
	}
	if wpErrFree != nil {
		wpErrFree.Call(errPtr)
	}
	return msg
}
