package flexi

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestEncode_RemoteProcess(t *testing.T) {
	type inner struct {
		Name string
		Age  int
	}
	name := "foo"
	age := 101
	want := &inner{Name: name, Age: age}
	b := new(bytes.Buffer)
	if err := json.NewEncoder(b).Encode(want); err != nil {
		t.Fatal(err)
	}

	rp := RemoteProcess{
		ID:      1,
		Addr:    "34.244.110.122:564",
		Name:    "bar",
		Spawned: b.Bytes(),
	}
	b.Truncate(0)
	if err := json.NewEncoder(b).Encode(&rp); err != nil {
		t.Fatal(err)
	}

	var rp1 RemoteProcess
	if err := json.NewDecoder(b).Decode(&rp1); err != nil {
		t.Fatal(err)
	}
	t.Logf("remote process decoded: %+v", rp1)

	var have inner
	if err := json.NewDecoder(rp1.SpawnedReader()).Decode(&have); err != nil {
		t.Fatal(err)
	}

	if have.Name != name {
		t.Fatalf("have [%v], want [%v]", have.Name, name)
	}
	if have.Age != age {
		t.Fatalf("have [%v], want [%v]", have.Age, age)
	}
}
