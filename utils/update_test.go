package utils

import (
	"testing"
)

type Data struct {
	ID   int
	Name string `update:"true"`
	Age  int    `update:"true"`
}

type PartialData struct {
	Name string `update:"true"`
}

func TestUpdateStruct_UpdatesFieldsWithTag(t *testing.T) {
	current := &Data{ID: 1, Name: "Old Name", Age: 30}
	newStruct := &Data{Name: "New Name"}

	UpdateStruct(current, newStruct)

	if current.Name != "New Name" {
		t.Errorf("expected Name to be 'New Name', got '%s'", current.Name)
	}
	if current.Age != 30 {
		t.Errorf("expected Age to remain 30, got %d", current.Age)
	}
}

func TestUpdateStruct_IgnoresFieldsWithoutTag(t *testing.T) {
	current := &Data{ID: 1, Name: "Old Name", Age: 30}
	newStruct := &Data{ID: 2}

	UpdateStruct(current, newStruct)

	if current.ID != 1 {
		t.Errorf("expected ID to remain 1, got %d", current.ID)
	}
}

func TestUpdateStruct_IgnoresZeroValues(t *testing.T) {
	current := &Data{ID: 1, Name: "Old Name", Age: 30}
	newStruct := &Data{Name: ""}

	UpdateStruct(current, newStruct)

	if current.Name != "Old Name" {
		t.Errorf("expected Name to remain 'Old Name', got '%s'", current.Name)
	}
}

func TestUpdateStruct_HandlesPartialStruct(t *testing.T) {
	current := &Data{ID: 1, Name: "Old Name", Age: 30}
	newStruct := &PartialData{Name: "New Name"}

	UpdateStruct(current, newStruct)

	if current.Name != "New Name" {
		t.Errorf("expected Name to be 'New Name', got '%s'", current.Name)
	}
	if current.Age != 30 {
		t.Errorf("expected Age to remain 30, got %d", current.Age)
	}
}

func BenchmarkUpdateStruct(b *testing.B) {
	current := &Data{ID: 1, Name: "Old Name", Age: 30}
	newStruct := &Data{Name: "New Name"}

	for i := 0; i < b.N; i++ {
		UpdateStruct(current, newStruct)
	}
}
