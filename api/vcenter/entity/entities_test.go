package entity

import (
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

type estruct struct {
	EntityType Entity `yaml:"entityType"`
}

func TestMarshalYAML(t *testing.T) {
	// string: correct casing
	e := ResourcePool
	expected := []byte("Resource Pool\n")

	out, err := yaml.Marshal(e)
	if err != nil {
		t.Errorf("failed to marshal ResourcePool: %v", err)
	}
	if !reflect.DeepEqual(out, expected) {
		t.Errorf("got %s != expected %s", string(out), string(expected))
	}

	// struct
	entityStruct := estruct{
		EntityType: ResourcePool,
	}
	expected = []byte("entityType: Resource Pool\n")

	out, err = yaml.Marshal(entityStruct)
	if err != nil {
		t.Errorf("failed to marshal entity struct: %v", err)
	}
	if !reflect.DeepEqual(out, expected) {
		t.Errorf("got %s != expected %s", string(out), string(expected))
	}
}

func TestUnMarshalYAML(t *testing.T) {
	// int
	in := []byte("8\n")
	expected := ResourcePool

	var e Entity
	if err := yaml.Unmarshal(in, &e); err != nil {
		t.Errorf("failed to unmarshal ResourcePool: %v", err)
	}
	if !reflect.DeepEqual(e, expected) {
		t.Errorf("got %v != expected %v", e, expected)
	}

	// string: correct casing
	in = []byte("Resource Pool\n")

	if err := yaml.Unmarshal(in, &e); err != nil {
		t.Errorf("failed to unmarshal ResourcePool: %v", err)
	}
	if !reflect.DeepEqual(e, expected) {
		t.Errorf("got %v != expected %v", e, expected)
	}

	// struct + case-insensitive
	in = []byte("entityType: resource pool\n")
	expectedS := estruct{EntityType: ResourcePool}

	var eS estruct
	if err := yaml.Unmarshal(in, &eS); err != nil {
		t.Errorf("failed to unmarshal ResourcePool: %v", err)
	}
	if !reflect.DeepEqual(eS, expectedS) {
		t.Errorf("got %v != expected %v", eS, expectedS)
	}
}
