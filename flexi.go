package flexi

import (
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

// Task defines **what** should be executed, on **which** hardware.
type Task struct {
	ID      string `json:"id"`
	Image   *Image `json:"image"`
	Caps    *Caps  `json:"caps"`
	RegAddr string `json:"reg_addr,omitempty"`
}

type Msg struct {
	ContentType string
	Type        string
	SessionId   string
	Body        []byte
}

func NewMsg(msgType, sessionId string) *Msg {
	return &Msg{Type: msgType, SessionId: sessionId}
}

func (m *msg) Marshal(v interface{}) error {
	body, err := json.Marshal(v)
	if err != nil {
		return err
	}
	m.Body = body
	m.ContentType = "application/json"
	return nil
}

func (m *Msg) Unmarshal(v interface{}) error {
	return json.Unmarshal(m.Body, v)
}
