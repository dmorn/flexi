package main

import (
	"fmt"
	"os"
	"encoding/json"
)

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

type Task struct {
	ID string `json:"id"`
	Image *Image `json:"image"`
	Caps *Caps`json:"capabilities"`
}

func errorf(format string, args ...interface{}) {
	fmt.Printf("error * "+format+"\n", args...)
}

func logf(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}

func main() {
	var task Task
	if err := json.NewDecoder(os.Stdin).Decode(&task); err != nil {
		errorf("decode task error: %w", err)
		return
	}

	logf("decoded task %v", task.ID)
}
