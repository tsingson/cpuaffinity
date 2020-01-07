package gofine

import (
	"os"
	"runtime"
	"testing"

	"golang.org/x/sys/unix"
)

var (
	busyLgore      = lgore{coreId: 0, state: Busy}
	availableLgore = lgore{coreId: 0, state: Available}
	original       unix.CPUSet
)

func TestOccupyBusy(t *testing.T) {
	err := busyLgore.occupy()

	if err == nil {
		t.Fatal("err should not be nil")
	}
}

func TestReleaseAvailable(t *testing.T) {
	err := availableLgore.release(original)

	if err != nil {
		t.Fatal("err should be nil")
	}
}

func TestOccupyRelease(t *testing.T) {
	lgore := availableLgore
	// do occupy
	err := lgore.occupy()
	if err != nil {
		t.Fatalf("err: %v", err.Error())
	}

	if lgore.state != Busy {
		t.Fatal("state should be Busy")
	}

	var cpuset unix.CPUSet
	unix.SchedGetaffinity(0, &cpuset)
	if !cpuset.IsSet(lgore.coreId) {
		t.Fatal("coreId should be set")
	}

	// do release
	err = lgore.release(original)
	if err != nil {
		t.Fatalf("err: %v", err.Error())
	}

	if lgore.state != Available {
		t.Fatal("state should be Available")
	}

	unix.SchedGetaffinity(0, &cpuset)
	if cpuset != original {
		t.Fatal("cpuset should be equal to original")
	}
}

func setup() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	unix.SchedGetaffinity(0, &original)
}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	os.Exit(code)
}
