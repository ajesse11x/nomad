package driver

import (
	"testing"
	"time"

	"github.com/hashicorp/nomad/client/config"
	"github.com/hashicorp/nomad/nomad/structs"

	ctestutils "github.com/hashicorp/nomad/client/testutil"
)

func TestExecDriver_Fingerprint(t *testing.T) {
	ctestutils.ExecCompatible(t)
	d := NewExecDriver(testDriverContext(""))
	node := &structs.Node{
		Attributes: make(map[string]string),
	}
	apply, err := d.Fingerprint(&config.Config{}, node)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !apply {
		t.Fatalf("should apply")
	}
	if node.Attributes["driver.exec"] == "" {
		t.Fatalf("missing driver")
	}
}

/*
TODO: This test is disabled til a follow-up api changes the restore state interface.
The driver/executor interface will be changed from Open to Cleanup, in which
clean-up tears down previous allocs.

func TestExecDriver_StartOpen_Wait(t *testing.T) {
	ctestutils.ExecCompatible(t)
	task := &structs.Task{
		Name: "sleep",
		Config: map[string]string{
			"command": "/bin/sleep",
			"args":    "5",
		},
	}

	driverCtx := testDriverContext(task.Name)
	ctx := testDriverExecContext(task, driverCtx)
	defer ctx.AllocDir.Destroy()
	d := NewExecDriver(driverCtx)

	if task.Resources == nil {
		task.Resources = &structs.Resources{}
	}
	task.Resources.CPU = 0.5
	task.Resources.MemoryMB = 2

	handle, err := d.Start(ctx, task)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if handle == nil {
		t.Fatalf("missing handle")
	}

	// Attempt to open
	handle2, err := d.Open(ctx, handle.ID())
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if handle2 == nil {
		t.Fatalf("missing handle")
	}
}
*/

func TestExecDriver_Start_Wait(t *testing.T) {
	ctestutils.ExecCompatible(t)
	task := &structs.Task{
		Name: "sleep",
		Config: map[string]string{
			"command": "/bin/sleep",
			"args":    "1",
		},
	}

	driverCtx := testDriverContext(task.Name)
	ctx := testDriverExecContext(task, driverCtx)
	defer ctx.AllocDir.Destroy()
	d := NewExecDriver(driverCtx)

	handle, err := d.Start(ctx, task)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if handle == nil {
		t.Fatalf("missing handle")
	}

	// Update should be a no-op
	err = handle.Update(task)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// Task should terminate quickly
	select {
	case err := <-handle.WaitCh():
		if err != nil {
			t.Fatalf("err: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout")
	}
}

func TestExecDriver_Start_Kill_Wait(t *testing.T) {
	ctestutils.ExecCompatible(t)
	task := &structs.Task{
		Name: "sleep",
		Config: map[string]string{
			"command": "/bin/sleep",
			"args":    "1",
		},
	}

	driverCtx := testDriverContext(task.Name)
	ctx := testDriverExecContext(task, driverCtx)
	defer ctx.AllocDir.Destroy()
	d := NewExecDriver(driverCtx)

	handle, err := d.Start(ctx, task)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if handle == nil {
		t.Fatalf("missing handle")
	}

	go func() {
		time.Sleep(100 * time.Millisecond)
		err := handle.Kill()
		if err != nil {
			t.Fatalf("err: %v", err)
		}
	}()

	// Task should terminate quickly
	select {
	case err := <-handle.WaitCh():
		if err == nil {
			t.Fatalf("should err: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout")
	}
}
