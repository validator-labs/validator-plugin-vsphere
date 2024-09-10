package entity

import (
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestMarshalYAML(t *testing.T) {
	e := ResourcePool
	expected := []byte("Resource Pool\n")

	out, err := yaml.Marshal(e)
	if err != nil {
		t.Errorf("failed to marshal ResourcePool: %v", err)
	}
	if !reflect.DeepEqual(out, expected) {
		t.Errorf("got %s != expected %s", string(out), string(expected))
	}
}

func TestUnMarshalYAML(t *testing.T) {
	in := []byte("Resource Pool\n")
	expected := ResourcePool

	var e Entity
	if err := yaml.Unmarshal(in, &e); err != nil {
		t.Errorf("failed to marshal ResourcePool: %v", err)
	}
	if !reflect.DeepEqual(e, expected) {
		t.Errorf("got %v != expected %v", e, expected)
	}
}
