// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package imagecustomizerapi

import (
	"fmt"
)

type Verity struct {
	// ID is used to correlate `Verity` objects with `FileSystem` objects.
	Id string `yaml:"id"`
	// The name of the mapper block device.
	// Must be 'root' for the rootfs (/) filesystem.
	Name string `yaml:"name"`
	// The ID of the 'Partition' to use as the data partition.
	DataDeviceId string `yaml:"dataDeviceId"`
	// The ID of the 'Partition' to use as the hash partition.
	HashDeviceId string `yaml:"hashDeviceId"`
	// How to handle corruption.
	CorruptionOption CorruptionOption `yaml:"corruptionOption"`
}

func (v *Verity) IsValid() error {
	if err := v.DataPartition.IsValid(); err != nil {
		return fmt.Errorf("invalid dataPartition: %v", err)
	}

	if err := v.HashPartition.IsValid(); err != nil {
		return fmt.Errorf("invalid hashPartition: %v", err)
	}

	if err := v.CorruptionOption.IsValid(); err != nil {
		return fmt.Errorf("invalid corruptionOption:\n%w", err)
	}

	return nil
}
