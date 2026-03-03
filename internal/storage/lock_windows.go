//go:build windows
// +build windows

// Package storage handles encrypted storage of projects and API keys using AES-256-GCM encryption.
package storage

import (
	"fmt"
	"os"

	"golang.org/x/sys/windows"
)

// FileLock provides cross-process file locking on Windows using LockFileEx.
type FileLock struct {
	file *os.File
	path string
}

const (
	// LOCKFILE_EXCLUSIVE_LOCK requests an exclusive lock
	lockfileExclusiveLock = 0x00000002
	// LOCKFILE_FAIL_IMMEDIATELY returns immediately if lock cannot be acquired
	lockfileFailImmediately = 0x00000001
)

// AcquireLock acquires an exclusive file lock, blocking until available.
// This uses the Windows LockFileEx API for cross-process safety.
func AcquireLock(lockPath string) (*FileLock, error) {
	file, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, fmt.Errorf("failed to open lock file: %w", err)
	}

	handle := windows.Handle(file.Fd())
	ol := new(windows.Overlapped)

	// LockFileEx with LOCKFILE_EXCLUSIVE_LOCK blocks until lock is acquired
	err = windows.LockFileEx(handle, lockfileExclusiveLock, 0, 1, 0, ol)
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to lock file: %w", err)
	}

	return &FileLock{file: file, path: lockPath}, nil
}

// TryAcquireLock attempts to acquire an exclusive file lock without blocking.
// Returns (nil, nil) if the lock is held by another process.
func TryAcquireLock(lockPath string) (*FileLock, error) {
	file, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, fmt.Errorf("failed to open lock file: %w", err)
	}

	handle := windows.Handle(file.Fd())
	ol := new(windows.Overlapped)

	// LOCKFILE_EXCLUSIVE_LOCK | LOCKFILE_FAIL_IMMEDIATELY = non-blocking exclusive lock
	err = windows.LockFileEx(handle, lockfileExclusiveLock|lockfileFailImmediately, 0, 1, 0, ol)
	if err != nil {
		file.Close()
		// Lock is held by another process
		return nil, nil
	}

	return &FileLock{file: file, path: lockPath}, nil
}

// Release releases the file lock and closes the underlying file.
func (l *FileLock) Release() error {
	if l == nil || l.file == nil {
		return nil
	}

	handle := windows.Handle(l.file.Fd())
	ol := new(windows.Overlapped)

	// Unlock the file region
	_ = windows.UnlockFileEx(handle, 0, 1, 0, ol)

	return l.file.Close()
}
