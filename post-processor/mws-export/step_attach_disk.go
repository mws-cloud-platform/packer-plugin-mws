// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package mwsexport

import (
	"context"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/mws-cloud-platform/packer-plugin-mws/builder/mws"
	computeref "go.mws.cloud/go-sdk/service/resources/references/compute"
)

type StepAttachDisk struct {
	Project        string
	Zone           string
	DiskType       string
	DiskIOPS       int64
	ImageRef       computeref.ImageRef
	CleanupTimeout time.Duration
}

func (s *StepAttachDisk) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get(mws.UIKey).(packer.Ui)
	driver := state.Get(mws.DriverKey).(Driver)
	prefix := state.Get(mws.PrefixKey).(string)
	vmName := state.Get(mws.InstanceIDKey).(string)
	diskName := prefix + "disk-for-export"

	ui.Sayf("Getting image %q info...", s.ImageRef.String())
	image, err := driver.GetImage(ctx, s.ImageRef)
	if err != nil {
		return mws.ActionHaltWithError(state, err)
	}
	if image.GetStatus().GetMinDiskSize() == nil {
		ui.Error("Image for export has unknown minimum disk size")
		return multistep.ActionHalt
	}

	ui.Say("Creating disk from image for export...")
	if err := driver.CreateDisk(ctx, mws.CreateDiskParams{
		DiskName: diskName,
		DiskType: s.DiskType,
		Size:     *image.GetStatus().GetMinDiskSize(),
		Iops:     s.DiskIOPS,
		ImageRef: &s.ImageRef,
		Zone:     s.Zone,
	}); err != nil {
		return mws.ActionHaltWithError(state, err)
	}

	state.Put(DiskForExportNameKey, diskName)
	diskRef := computeref.NewDiskRef(s.Project, diskName)

	ui.Sayf("Attaching disk for export %q to the virtual machine %q...", diskName, vmName)
	if err := driver.AttachDiskToVirtualMachine(ctx, vmName, diskRef); err != nil {
		return mws.ActionHaltWithError(state, err)
	}

	ui.Sayf("Disk for export %q attached", diskName)

	return multistep.ActionContinue
}

func (s *StepAttachDisk) Cleanup(state multistep.StateBag) {
	diskName, ok := state.Get(DiskForExportNameKey).(string)
	if !ok {
		return
	}

	ui := state.Get(mws.UIKey).(packer.Ui)
	driver := state.Get(mws.DriverKey).(Driver)
	vmName := state.Get(mws.InstanceIDKey).(string)

	ctx, cancel := context.WithTimeout(context.Background(), s.CleanupTimeout)
	defer cancel()

	if err := driver.DetachSecondaryDisksFromVirtualMachine(ctx, vmName); err != nil {
		ui.Errorf("Error detaching disk for export %q from virtual machine %q. "+
			"Error: %v.\n"+
			"Trying to delete it anyway...", diskName, vmName, err)
	} else {
		ui.Say("Secondary disk detached")
	}

	if err := driver.DeleteDisk(ctx, diskName); err != nil {
		ui.Errorf("Error deleting disk for export %q. "+
			"Please detach (if necessary) and delete it manually.\n"+
			"Error: %v.", diskName, err)
	} else {
		ui.Sayf("Disk for export %q deleted", diskName)
	}
}
