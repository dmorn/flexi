package flexi

import (
	"context"
)

type Image struct {
	Type    string `json:"type"`
	Name    string `json:"name"`
	Service string `json:"service"`
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
	ID    string `json:"id"`
	Image *Image `json:"image"`
	Caps  *Caps  `json:"capabilities"`
}

type RemoteProcess interface {
	Addr() string
	Name() string
}

type Spawner interface {
	Spawn(context.Context, Task) (RemoteProcess, error)
	Kill(context.Context, RemoteProcess) error
}
