package vcenter

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

// GetVMInfos fetches VM information from vCenter
func GetVMInfos() ([]VMInfo, error) {
	c := GetCachedClient()
	if c == nil {
		return nil, fmt.Errorf("no active govmomi client")
	}

	ctx := context.Background()
	mgr := view.NewManager(c)
	v, err := mgr.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"VirtualMachine"}, true)
	if err != nil {
		return nil, fmt.Errorf("CreateContainerView: %w", err)
	}
	defer v.Destroy(ctx)

	var vms []mo.VirtualMachine
	if err := v.Retrieve(ctx, []string{"VirtualMachine"}, []string{"name", "guest", "parent"}, &vms); err != nil {
		return nil, fmt.Errorf("retrieve: %w", err)
	}

	pc := property.DefaultCollector(c)
	out := make([]VMInfo, 0, len(vms))
	for _, vm := range vms {
		if vm.Name == "" {
			continue
		}
		osName := "Unknown"
		domain := ""
		if vm.Guest != nil {
			if vm.Guest.GuestFullName != "" {
				osName = vm.Guest.GuestFullName
			}
			if vm.Guest.HostName != "" {
				domain = vm.Guest.HostName
			}
		}
		folder := "/"
		if vm.Parent != nil {
			if fp, err := folderPath(ctx, pc, *vm.Parent); err == nil {
				folder = fp
			}
		}
		out = append(out, VMInfo{Name: vm.Name, Folder: folder, OS: osName, Domain: domain, Ref: vm.Reference()})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

// folderPath builds full path by traversing up the tree
func folderPath(ctx context.Context, pc *property.Collector, ref types.ManagedObjectReference) (string, error) {
	var segments []string
	current := ref
	for {
		var obj mo.ManagedEntity
		if err := pc.RetrieveOne(ctx, current, []string{"name", "parent"}, &obj); err != nil {
			return "", err
		}
		if obj.Parent == nil || obj.Parent.Type == "Datacenter" {
			break
		}
		segments = append([]string{obj.Name}, segments...)
		current = *obj.Parent
	}
	if len(segments) == 0 {
		return "/", nil
	}
	return "/" + strings.Join(segments, "/"), nil
}
