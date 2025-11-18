package upgrade

import (
	"context"
	"errors"

	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

// MountISO mounts an ISO to the VM's CD-ROM
func MountISO(ctx context.Context, vm *object.VirtualMachine, isoPath string) error {
	var o mo.VirtualMachine
	if err := vm.Properties(ctx, vm.Reference(), []string{"config.hardware.device"}, &o); err != nil {
		return err
	}
	var cd *types.VirtualCdrom
	for _, dev := range o.Config.Hardware.Device {
		if v, ok := dev.(*types.VirtualCdrom); ok {
			cd = v
			break
		}
	}
	if cd == nil {
		return errors.New("no CD/DVD device")
	}
	cdCopy := *cd
	cd = &cdCopy
	cd.Backing = &types.VirtualCdromIsoBackingInfo{
		VirtualDeviceFileBackingInfo: types.VirtualDeviceFileBackingInfo{
			FileName: isoPath,
		},
	}
	cd.Connectable = &types.VirtualDeviceConnectInfo{
		StartConnected:   true,
		Connected:        true,
		AllowGuestControl: true,
	}
	spec := types.VirtualMachineConfigSpec{
		DeviceChange: []types.BaseVirtualDeviceConfigSpec{
			&types.VirtualDeviceConfigSpec{
				Operation: types.VirtualDeviceConfigSpecOperationEdit,
				Device:    cd,
			},
		},
	}
	task, err := vm.Reconfigure(ctx, spec)
	if err != nil {
		return err
	}
	return task.Wait(ctx)
}

// UnmountISO unmounts ISO from the VM's CD-ROM
func UnmountISO(ctx context.Context, vm *object.VirtualMachine) error {
	var o mo.VirtualMachine
	if err := vm.Properties(ctx, vm.Reference(), []string{"config.hardware.device"}, &o); err != nil {
		return err
	}
	var cd *types.VirtualCdrom
	for _, dev := range o.Config.Hardware.Device {
		if v, ok := dev.(*types.VirtualCdrom); ok {
			cd = v
			break
		}
	}
	if cd == nil {
		return errors.New("no CD/DVD device")
	}
	cdCopy := *cd
	cd = &cdCopy
	cd.Connectable = &types.VirtualDeviceConnectInfo{
		StartConnected:   false,
		Connected:        false,
		AllowGuestControl: true,
	}
	cd.Backing = &types.VirtualCdromRemotePassthroughBackingInfo{}
	spec := types.VirtualMachineConfigSpec{
		DeviceChange: []types.BaseVirtualDeviceConfigSpec{
			&types.VirtualDeviceConfigSpec{
				Operation: types.VirtualDeviceConfigSpecOperationEdit,
				Device:    cd,
			},
		},
	}
	task, err := vm.Reconfigure(ctx, spec)
	if err != nil {
		return err
	}
	return task.Wait(ctx)
}
