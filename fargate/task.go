// SPDX-FileCopyrightText: 2020 jecoz
//
// SPDX-License-Identifier: BSD-3-Clause

package fargate

type Image struct {
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
type Task struct {
	ID        string `json:"id"`
	ImageType string `json:"image_type"`
	Image     *Image `json:"image"`
	Caps      *Caps  `json:"capabilities"`
}

type Container struct {
	Addr    string `json:"addr"`
	Name    string `json:"name"`
	Cluster string `json:"cluster"`
}
