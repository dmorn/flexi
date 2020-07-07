// SPDX-FileCopyrightText: 2020 jecoz
//
// SPDX-License-Identifier: BSD-3-Clause

package flexi

import (
	"context"
)

// TODO: this is just one type of image that we might want
// to pass around. Make the unmarshaling more flexible.
type FargateImage struct {
	Name           string   `json:"name"`
	Service        string   `json:"service"`
	Cluster        string   `json:"cluster"`
	Subnets        []string `json:"subnets"`
	SecurityGroups []string `json:"security_groups"`
}

// Based on the required capabilities, we'll choose where the
// container should be executed.
type Caps struct {
	CPU int `json:"cpu"`
	Ram int `json:"ram"`
	GPU int `json:"gpu"`
}

// Task defines **what** should be executed, on **which** hardware.
// TODO: Image should be of type interface{}.
type Task struct {
	ID        string        `json:"id"`
	ImageType string        `json:"image_type"`
	Image     *FargateImage `json:"image"`
	Caps      *Caps         `json:"capabilities"`
}

type RemoteProcess struct {
	Tags    []string `json:"tags"`
	Addr    string   `json:"addr"`
	Name    string   `json:"name"`
	Cluster string   `json:"cluster"`
}

type Spawner interface {
	Spawn(context.Context, Task) (*RemoteProcess, error)
	Kill(context.Context, *RemoteProcess) error
}