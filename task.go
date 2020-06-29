package flexi

// Image describes what will be executed.
type Image struct {
	// Type is usually docker.
	Type string `json:"type"`
	// Name is usually an address to a docker image stored
	// in some registry.
	Name string `json:"name"`
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

type RemoteProcess struct {
}

type Spawner interface {
	Spawn(Task) (*RemoteProcess, error)
}
