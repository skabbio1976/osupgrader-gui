package vcenter

import (
	"context"
	"errors"
	"fmt"

	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/methods"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

// CreateSnapshot skapar en snapshot för en VM
func CreateSnapshot(ctx context.Context, vm *object.VirtualMachine, name, description string, memory, quiesce bool) error {
	task, err := vm.CreateSnapshot(ctx, name, description, memory, quiesce)
	if err != nil {
		return err
	}
	if err := task.Wait(ctx); err != nil {
		return err
	}

	// Verifiera att snapshot faktiskt skapades
	var o mo.VirtualMachine
	if err := vm.Properties(ctx, vm.Reference(), []string{"snapshot"}, &o); err != nil {
		return fmt.Errorf("kunde inte verifiera snapshot: %w", err)
	}
	if o.Snapshot == nil || o.Snapshot.CurrentSnapshot == nil {
		return fmt.Errorf("snapshot skapades men verifiering misslyckades")
	}

	return nil
}

// ListSnapshots listar alla snapshots för en VM
func ListSnapshots(ctx context.Context, vm *object.VirtualMachine, vmName string) ([]SnapshotEntry, error) {
	var o mo.VirtualMachine
	if err := vm.Properties(ctx, vm.Reference(), []string{"snapshot"}, &o); err != nil {
		return nil, err
	}
	if o.Snapshot == nil || o.Snapshot.RootSnapshotList == nil {
		return nil, nil
	}

	out := []SnapshotEntry{}
	var walk func(list []types.VirtualMachineSnapshotTree)
	walk = func(list []types.VirtualMachineSnapshotTree) {
		for _, n := range list {
			out = append(out, SnapshotEntry{VMName: vmName, SnapshotName: n.Name, Ref: n.Snapshot})
			if n.ChildSnapshotList != nil {
				walk(n.ChildSnapshotList)
			}
		}
	}
	walk(o.Snapshot.RootSnapshotList)
	return out, nil
}

// RemoveSnapshot tar bort en snapshot
func RemoveSnapshot(ctx context.Context, snapRef types.ManagedObjectReference) error {
	c := GetCachedClient()
	if c == nil {
		return errors.New("ingen aktiv klient")
	}

	req := &types.RemoveSnapshot_Task{This: snapRef, RemoveChildren: false}
	res, err := methods.RemoveSnapshot_Task(ctx, c, req)
	if err != nil {
		return fmt.Errorf("RemoveSnapshot_Task: %w", err)
	}
	if res == nil || res.Returnval.Type == "" {
		return fmt.Errorf("tomt svar från RemoveSnapshot_Task")
	}
	task := object.NewTask(c, res.Returnval)
	return task.Wait(ctx)
}
