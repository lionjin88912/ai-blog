//go:build !windows

// Package pwshshim: on non-Windows builds the pwsh shim is irrelevant, so the
// embedded binary is empty (keeps the darwin/linux exes from carrying ~2MB).
package pwshshim

// Binary is empty on non-Windows platforms.
var Binary []byte
