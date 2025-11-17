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

// ValidateISOPath kontrollerar att ISO-path har korrekt format och att datastoren finns
// OBS: Kontrollerar INTE att filen faktiskt finns - det sker vid mount-steget
func ValidateISOPath(ctx context.Context, isoPath string) error {
	c := vcenter.GetCachedClient()
	if c == nil {
		return errors.New("ingen aktiv govmomi-klient")
	}

	// Parse datastore path format: [datastore1] path/to/file.iso
	isoPath = strings.TrimSpace(isoPath)
	if !strings.HasPrefix(isoPath, "[") || !strings.Contains(isoPath, "]") {
		return fmt.Errorf("ogiltigt ISO-path format (förväntar [datastore] path/file.iso): %s", isoPath)
	}

	// Extrahera datastore-namn och path
	parts := strings.SplitN(isoPath, "]", 2)
	if len(parts) != 2 {
		return fmt.Errorf("kunde inte parse ISO-path: %s", isoPath)
	}
	dsName := strings.TrimSpace(strings.TrimPrefix(parts[0], "["))
	filePath := strings.TrimSpace(parts[1])

	// Ta bort ledande slash om det finns för normalisering
	filePath = strings.TrimPrefix(filePath, "/")

	if dsName == "" || filePath == "" {
		return fmt.Errorf("datastore-namn eller filpath är tomt: %s", isoPath)
	}

	// Kontrollera att filePath slutar med .iso
	if !strings.HasSuffix(strings.ToLower(filePath), ".iso") {
		return fmt.Errorf("filpath måste sluta med .iso: %s", filePath)
	}

	// Använd view.Manager för att hitta alla datastores (officiell metod)
	m := view.NewManager(c)

	v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"Datastore"}, true)
	if err != nil {
		return fmt.Errorf("kunde inte skapa container view: %w", err)
	}
	defer v.Destroy(ctx)

	// Hämta alla datastores
	var dss []mo.Datastore
	err = v.Retrieve(ctx, []string{"Datastore"}, []string{"summary"}, &dss)
	if err != nil {
		return fmt.Errorf("kunde inte hämta datastores: %w", err)
	}

	// Leta efter vår datastore
	dsNameLower := strings.ToLower(dsName)
	found := false
	for _, ds := range dss {
		if strings.ToLower(ds.Summary.Name) == dsNameLower {
			found = true
			break
		}
	}

	if !found {
		// Visa tillgängliga datastores i felmeddelandet
		var availableDS []string
		for _, ds := range dss {
			availableDS = append(availableDS, ds.Summary.Name)
		}
		return fmt.Errorf("datastore '%s' ej hittad. Tillgängliga: %v", dsName, availableDS)
	}

	// Om vi kom hit så är formatet OK och datastoren finns
	// Själva filen kontrolleras vid mount-steget
	return nil
}

// CheckUpgradeInProgress kontrollerar om en uppgradering redan pågår på VM
func CheckUpgradeInProgress(ctx context.Context, vm *object.VirtualMachine) (bool, error) {
	var o mo.VirtualMachine
	if err := vm.Properties(ctx, vm.Reference(), []string{"guest.guestOperationsReady", "runtime.powerState"}, &o); err != nil {
		return false, err
	}

	// Kontrollera om VM är avstängd eller startar om
	if o.Runtime.PowerState != types.VirtualMachinePowerStatePoweredOn {
		return true, fmt.Errorf("VM är inte påslagen (state: %s)", o.Runtime.PowerState)
	}

	return false, nil
}

// GetSystemDrive hittar system-drive (vanligtvis C:\) men kan vara annat
func GetSystemDrive(ctx context.Context, vm *object.VirtualMachine) (string, error) {
	var o mo.VirtualMachine
	if err := vm.Properties(ctx, vm.Reference(), []string{"guest.disk", "config.guestId"}, &o); err != nil {
		return "", err
	}

	if o.Guest == nil || o.Guest.Disk == nil {
		return "", errors.New("ingen guest disk info (VMware Tools?)")
	}

	// Windows system är nästan alltid C:\ men kontrollera
	for _, d := range o.Guest.Disk {
		path := strings.ToLower(d.DiskPath)
		if strings.HasPrefix(path, "c:") {
			return "C:\\", nil
		}
	}

	// Om C:\ inte hittas, returnera första disken
	if len(o.Guest.Disk) > 0 {
		firstDisk := o.Guest.Disk[0].DiskPath
		if strings.Contains(firstDisk, ":") {
			return strings.Split(firstDisk, ":")[0] + ":\\", nil
		}
	}

	return "C:\\", nil // Fallback
}

// WaitForToolsRunningWithTimeout väntar på att VMware Tools blir redo med timeout
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
			// Vänta 5 sekunder innan nästa försök
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(5 * time.Second):
			}
		}
	}
}

// GetDiskFreeGB hämtar ledigt diskutrymme i GB för en specifik drive
func GetDiskFreeGB(ctx context.Context, vm *object.VirtualMachine, drive string) (int64, error) {
	var o mo.VirtualMachine
	if err := vm.Properties(ctx, vm.Reference(), []string{"guest.disk"}, &o); err != nil {
		return 0, err
	}
	if o.Guest == nil || o.Guest.Disk == nil {
		return 0, errors.New("ingen guest disk info (VMware Tools?)")
	}
	dLower := strings.ToLower(drive)
	for _, d := range o.Guest.Disk {
		if strings.ToLower(d.DiskPath) == dLower {
			return d.FreeSpace / (1024 * 1024 * 1024), nil
		}
	}
	return 0, fmt.Errorf("drive %s ej hittad", drive)
}
