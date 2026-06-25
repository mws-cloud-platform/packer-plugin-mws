package mwsexport

import (
	"cmp"
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/mws-cloud-platform/packer-plugin-mws/builder/mws"
	"go.mws.cloud/go-sdk/pkg/apimodels/units/bytesize"
	computeref "go.mws.cloud/go-sdk/service/resources/references/compute"
	"go.mws.cloud/util-toolset/pkg/utils/ptr"
)

type stepAttachDiskForExport struct {
	Config
}

func (s *stepAttachDiskForExport) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get(mws.DriverKey).(mws.Driver)
	ui := state.Get(mws.UIKey).(packer.Ui)
	instanceID := state.Get(mws.InstanceIDKey).(string)
	prefix := state.Get(mws.UUIDPrefixKey).(string)
	projectForExport := state.Get(mws.ProjectForExportKey).(string)
	imageForExport := state.Get(mws.ImageForExportKey).(string)

	image, err := driver.GetImage(ctx, projectForExport, imageForExport)
	if err != nil {
		return mws.ActionHaltWithError(state, fmt.Errorf("get image for export: %w", err))
	}

	imageMinDiskSize := cmp.Or(ptr.Value(image.Status.MinDiskSize), bytesize.MustParseString(mws.DefaultDiskSize))
	additionalSize, err := bytesize.ParseString(s.AdditionalDiskForExportSize)
	if err != nil {
		return mws.ActionHaltWithError(state, fmt.Errorf("parse additional_disk_for_export_size: %w", err))
	}
	diskSize := bytesize.MustNewFromBigInt(new(big.Int).Add(imageMinDiskSize.BigInt(), additionalSize.BigInt()), bytesize.B)

	diskName := cmp.Or(s.DiskName, prefix+"disk-for-export")

	ui.Say("Creating disk for export...")
	if err := driver.CreateDisk(ctx, mws.CreateDiskParams{
		DiskName: diskName,
		DiskType: s.DiskForExportType,
		Size:     diskSize,
		Iops:     s.DiskForExportIOPS,
		ImageRef: new(computeref.NewImageRef(projectForExport, imageForExport)),
		Zone:     s.Zone,
	}); err != nil {
		return mws.ActionHaltWithError(state, fmt.Errorf("create disk for export %q: %w", diskName, err))
	}
	state.Put(mws.DiskForExportNameKey, diskName)

	diskRef := new(computeref.NewDiskRef(s.Project, diskName))
	ui.Sayf("Disk for export %q created", diskName)

	ui.Sayf("Attaching disk...")
	if err := driver.AttachDiskToVirtualMachine(ctx, mws.AttachDiskToVirtualMachineParams{
		VirtualMachineName: instanceID,
		DiskRef:            diskRef,
	}); err != nil {
		return mws.ActionHaltWithError(state, fmt.Errorf("attach disk for export: %w", err))
	}

	ui.Sayf("Disk for export attached")

	return multistep.ActionContinue
}

func (s *stepAttachDiskForExport) Cleanup(state multistep.StateBag) {
	driver := state.Get(mws.DriverKey).(mws.Driver)
	ui := state.Get(mws.UIKey).(packer.Ui)
	virtualMachineName := state.Get(mws.InstanceIDKey).(string)
	diskName := mws.StateGetOkString(state, mws.DiskForExportNameKey)

	cleanupTimeout, _ := time.ParseDuration(s.CleanupTimeout)
	ctx, cancel := context.WithTimeout(context.Background(), cleanupTimeout)
	defer cancel()

	if err := driver.DetachDisksFromVirtualMachine(ctx, virtualMachineName); err != nil {
		ui.Errorf("Error detaching disk for export %q from vm %q.\n"+
			"Error: %v.", diskName, virtualMachineName, err)
	} else {
		ui.Sayf("Disk for export %q detached", diskName)
	}

	if err := driver.DeleteDisk(ctx, diskName); err != nil {
		ui.Errorf("Error deleting disk for export %q. Please delete it manually.\n"+
			"Error: %v.", diskName, err)
	} else {
		ui.Sayf("Disk for export %q deleted", diskName)
	}
}
