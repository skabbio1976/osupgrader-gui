package vcenter

import (
	"github.com/vmware/govmomi/vim25/types"
)

// VMInfo innehåller metadata för visning/urval
type VMInfo struct {
	Name   string
	Folder string
	OS     string
	Domain string
	Ref    types.ManagedObjectReference
}

// SnapshotEntry representerar en snapshot
type SnapshotEntry struct {
	VMName       string
	SnapshotName string
	Ref          types.ManagedObjectReference
}

// GuestCreds innehåller guest OS credentials
type GuestCreds struct {
	User string
	Pass string
}
