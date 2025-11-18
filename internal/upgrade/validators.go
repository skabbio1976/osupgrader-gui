package upgrade

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
	"github.com/yourusername/osupgrader-gui/internal/vcenter"
)

// ValidateISOPath checks that ISO path has correct format and that the datastore exists
// NOTE: Does NOT check if the file actually exists - that happens at mount step
func ValidateISOPath(ctx context.Context, isoPath string) error {
	c := vcenter.GetCachedClient()
	if c == nil {
		return errors.New("no active govmomi client")
	}

	// Parse datastore path format: [datastore1] path/to/file.iso
	isoPath = strings.TrimSpace(isoPath)
	if !strings.HasPrefix(isoPath, "[") || !strings.Contains(isoPath, "]") {
		return fmt.Errorf("invalid ISO path format (expects [datastore] path/file.iso): %s", isoPath)
	}

	// Extract datastore name and path
	parts := strings.SplitN(isoPath, "]", 2)
	if len(parts) != 2 {
		return fmt.Errorf("could not parse ISO path: %s", isoPath)
	}
	dsName := strings.TrimSpace(strings.TrimPrefix(parts[0], "["))
	filePath := strings.TrimSpace(parts[1])

	// Remove leading slash if present for normalization
	filePath = strings.TrimPrefix(filePath, "/")

	if dsName == "" || filePath == "" {
		return fmt.Errorf("datastore name or file path is empty: %s", isoPath)
	}

	// Check that filePath ends with .iso
	if !strings.HasSuffix(strings.ToLower(filePath), ".iso") {
		return fmt.Errorf("file path must end with .iso: %s", filePath)
	}

	// Use view.Manager to find all datastores (official method)
	m := view.NewManager(c)

	v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"Datastore"}, true)
	if err != nil {
		return fmt.Errorf("could not create container view: %w", err)
	}
	defer v.Destroy(ctx)

	// Fetch all datastores
	var dss []mo.Datastore
	err = v.Retrieve(ctx, []string{"Datastore"}, []string{"summary"}, &dss)
	if err != nil {
		return fmt.Errorf("could not fetch datastores: %w", err)
	}

	// Look for our datastore
	dsNameLower := strings.ToLower(dsName)
	found := false
	for _, ds := range dss {
		if strings.ToLower(ds.Summary.Name) == dsNameLower {
			found = true
			break
		}
	}

	if !found {
		// Show available datastores in error message
		var availableDS []string
		for _, ds := range dss {
			availableDS = append(availableDS, ds.Summary.Name)
		}
		return fmt.Errorf("datastore '%s' not found. Available: %v", dsName, availableDS)
	}

	// If we got here, format is OK and datastore exists
	// The actual file is checked at mount step
	return nil
}

// CheckUpgradeInProgress checks if an upgrade is already in progress on the VM
func CheckUpgradeInProgress(ctx context.Context, vm *object.VirtualMachine) (bool, error) {
	var o mo.VirtualMachine
	if err := vm.Properties(ctx, vm.Reference(), []string{"guest.guestOperationsReady", "runtime.powerState"}, &o); err != nil {
		return false, err
	}

	// Check if VM is powered off or restarting
	if o.Runtime.PowerState != types.VirtualMachinePowerStatePoweredOn {
		return true, fmt.Errorf("VM is not powered on (state: %s)", o.Runtime.PowerState)
	}

	return false, nil
}

// GetSystemDrive finds system drive (usually C:\) but can be other
func GetSystemDrive(ctx context.Context, vm *object.VirtualMachine) (string, error) {
	var o mo.VirtualMachine
	if err := vm.Properties(ctx, vm.Reference(), []string{"guest.disk", "config.guestId"}, &o); err != nil {
		return "", err
	}

	if o.Guest == nil || o.Guest.Disk == nil {
		return "", errors.New("no guest disk info (VMware Tools?)")
	}

	// Windows system is almost always C:\ but check
	for _, d := range o.Guest.Disk {
		path := strings.ToLower(d.DiskPath)
		if strings.HasPrefix(path, "c:") {
			return "C:\\", nil
		}
	}

	// If C:\ not found, return first disk
	if len(o.Guest.Disk) > 0 {
		firstDisk := o.Guest.Disk[0].DiskPath
		if strings.Contains(firstDisk, ":") {
			return strings.Split(firstDisk, ":")[0] + ":\\", nil
		}
	}

	return "C:\\", nil // Fallback
}

// WaitForToolsRunningWithTimeout waits for VMware Tools to become ready with timeout
func WaitForToolsRunningWithTimeout(ctx context.Context, vm *object.VirtualMachine) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			var o mo.VirtualMachine
			if err := vm.Properties(ctx, vm.Reference(), []string{"guest.toolsRunningStatus"}, &o); err != nil {
				return err
			}
			if o.Guest != nil && o.Guest.ToolsRunningStatus == "guestToolsRunning" {
				return nil
			}
			// Wait 5 seconds before next attempt
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(5 * time.Second):
			}
		}
	}
}

// GetDiskFreeGB fetches free disk space in GB for a specific drive
func GetDiskFreeGB(ctx context.Context, vm *object.VirtualMachine, drive string) (int64, error) {
	var o mo.VirtualMachine
	if err := vm.Properties(ctx, vm.Reference(), []string{"guest.disk"}, &o); err != nil {
		return 0, err
	}
	if o.Guest == nil || o.Guest.Disk == nil {
		return 0, errors.New("no guest disk info (VMware Tools?)")
	}
	dLower := strings.ToLower(drive)
	for _, d := range o.Guest.Disk {
		if strings.ToLower(d.DiskPath) == dLower {
			return d.FreeSpace / (1024 * 1024 * 1024), nil
		}
	}
	return 0, fmt.Errorf("drive %s not found", drive)
}
