//go:build windows

package jobserver

import (
	"fmt"
	"runtime"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	kernel32                = windows.NewLazySystemDLL("kernel32.dll")
	procOpenSemaphoreA      = kernel32.NewProc("OpenSemaphoreA")
	procReleaseSemaphore    = kernel32.NewProc("ReleaseSemaphore")
	procWaitForSingleObject = kernel32.NewProc("WaitForSingleObject")
)

const (
	WAIT_TIMEOUT = 0x00000102

	SEMAPHORE_MODIFY_STATE      = 0x0002
	SYNCHRONIZATION_SYNCHRONIZE = 0x00100000
)

const (
	// Hard-coded value used for explicit slots acquired from a Windows
	// semaphore. Windows jobservers are counted, not token-based, so the
	// value is never written back anywhere.
	explicitSlotValue = 1
)

func openSemaphoreA(desiredAccess uint32, inheritHandle bool, name string) (windows.Handle, error) {
	p, err := windows.BytePtrFromString(name)
	if err != nil {
		return 0, err
	}
	var inherit uintptr
	if inheritHandle {
		inherit = 1
	}
	r1, _, e1 := procOpenSemaphoreA.Call(
		uintptr(desiredAccess),
		inherit,
		uintptr(unsafe.Pointer(p)),
	)
	if r1 == 0 {
		return 0, e1
	}
	return windows.Handle(r1), nil
}

func releaseSemaphore(h windows.Handle, releaseCount int32) error {
	r1, _, e1 := procReleaseSemaphore.Call(
		uintptr(h),
		uintptr(releaseCount),
		0,
	)
	if r1 == 0 {
		return e1
	}
	return nil
}

func waitForSingleObject(h windows.Handle, millis uint32) (uint32, error) {
	r1, _, e1 := procWaitForSingleObject.Call(uintptr(h), uintptr(millis))
	if uint32(r1) == windows.WAIT_FAILED {
		return uint32(r1), e1
	}
	return uint32(r1), nil
}

type windowsClient struct {
	handle          windows.Handle
	hasImplicitSlot bool
}

func (c *windowsClient) TryAcquire() Slot {
	if c.IsValid() {
		if c.hasImplicitSlot {
			c.hasImplicitSlot = false
			return CreateImplicitSlot()
		}

		event, err := waitForSingleObject(c.handle, 0)
		if err == nil && event == windows.WAIT_OBJECT_0 {
			return CreateExplicitSlot(explicitSlotValue)
		}
	}
	return NewSlot()
}

func (c *windowsClient) Release(slot Slot) {
	if !slot.Valid() {
		return
	}
	if slot.Implicit() {
		if c.hasImplicitSlot {
			panic("Implicit slot cannot be released twice!")
		}
		c.hasImplicitSlot = true
		return
	}
	// Nothing useful can be done if the semaphore was closed underneath us;
	// ninja ignores this error as well.
	_ = releaseSemaphore(c.handle, 1)
}

func (c *windowsClient) InitWithSemaphore(name string) error {
	if name == "" {
		return fmt.Errorf("Empty semaphore name")
	}
	h, err := openSemaphoreA(
		SYNCHRONIZATION_SYNCHRONIZE|SEMAPHORE_MODIFY_STATE,
		false,
		name,
	)
	if err != nil {
		return fmt.Errorf("Error opening semaphore %q: %s", name, err)
	}
	c.handle = h
	c.hasImplicitSlot = true

	runtime.AddCleanup(c, func(h windows.Handle) {
		_ = syscall.CloseHandle(syscall.Handle(h))
	}, h)
	return nil
}

func (c *windowsClient) IsValid() bool {
	return c.handle != 0
}

func (c *windowsClient) Create(config *Config) error {
	if config.Mode == ModeWin32Semaphore || config.Mode == ModeDefault {
		return c.InitWithSemaphore(config.Path)
	}
	return fmt.Errorf("Unsupported jobserver mode")
}

func NewClient() Client {
	return &windowsClient{}
}
