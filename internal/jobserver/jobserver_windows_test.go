//go:build windows

package jobserver

import (
	"fmt"
	"os"
	"testing"
	"time"
	"unsafe"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

var procCreateSemaphoreW = windows.NewLazySystemDLL("kernel32.dll").NewProc("CreateSemaphoreW")

func createTestSemaphore(t *testing.T, initial, max int32) (windows.Handle, string) {
	t.Helper()
	name := fmt.Sprintf("reninja-jobserver-test-%d-%d", os.Getpid(), time.Now().UnixNano())
	namePtr, err := windows.UTF16PtrFromString(name)
	require.NoError(t, err)
	handle, _, callErr := procCreateSemaphoreW.Call(
		0,
		uintptr(initial),
		uintptr(max),
		uintptr(unsafe.Pointer(namePtr)),
	)
	require.NotZero(t, handle, "CreateSemaphoreW failed: %v", callErr)
	return windows.Handle(handle), name
}

func TestWindowsJobserverSemaphoreLifecycle(t *testing.T) {
	handle, name := createTestSemaphore(t, 2, 2)
	defer windows.CloseHandle(handle)

	config := NewConfig(ModeWin32Semaphore)
	config.Path = name
	client := NewClient()
	require.NoError(t, client.Create(config))

	// The process itself owns one implicit slot.
	implicit := client.TryAcquire()
	require.True(t, implicit.Implicit())

	first := client.TryAcquire()
	require.True(t, first.Explicit())
	assert.Equal(t, uint8(1), first.ExplicitValue())

	second := client.TryAcquire()
	require.True(t, second.Explicit())

	assert.False(t, client.TryAcquire().Valid(), "semaphore should be exhausted")

	client.Release(first)
	client.Release(second)
	reacquired := client.TryAcquire()
	require.True(t, reacquired.Explicit())
	client.Release(reacquired)

	client.Release(implicit)
	assert.True(t, client.TryAcquire().Implicit())
}

func TestWindowsJobserverUnsupportedMode(t *testing.T) {
	client := NewClient()
	require.Error(t, client.Create(NewConfig(ModeNone)))
	require.Error(t, client.Create(NewConfig(ModePosixFifo)))
}

func TestWindowsJobserverMissingSemaphore(t *testing.T) {
	client := NewClient()
	config := NewConfig(ModeWin32Semaphore)
	config.Path = fmt.Sprintf("reninja-missing-semaphore-%d", time.Now().UnixNano())
	require.Error(t, client.Create(config))
}

func TestWindowsJobserverEmptyName(t *testing.T) {
	client := NewClient()
	require.Error(t, client.Create(NewConfig(ModeWin32Semaphore)))
}
