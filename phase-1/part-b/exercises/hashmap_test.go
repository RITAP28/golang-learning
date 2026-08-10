package main

import (
	"testing"
)

// TestSeparateChaining tests the Linked List version
func TestSeparateChaining(t *testing.T) {
	m := &HashMap{
		buckets:  make([]*Node, 5),
		capacity: 5,
	}

	// Test Insertion
	m.Put("apple", 10)
	val, exists := m.Get("apple")
	if !exists || val != 10 {
		t.Errorf("Expected 10, got %d", val)
	}

	// Test Update
	m.Put("apple", 99)
	val, _ = m.Get("apple")
	if val != 99 {
		t.Errorf("Expected 99 after update, got %d", val)
	}

	// Test Resize
	// Capacity is 5, 0.75 threshold is 3. Adding 4th and 5th items triggers resize.
	m.Put("banana", 20)
	m.Put("cherry", 30)
	m.Put("date", 40)

	if m.capacity <= 5 {
		t.Errorf("Expected resize to happen, but capacity is still %d", m.capacity)
	}

	// Verify data after resize
	if v, ok := m.Get("apple"); !ok || v != 99 {
		t.Error("Lost 'apple' during resize")
	}
}

// TestOpenAddressing tests the Linear Probing version
func TestOpenAddressing(t *testing.T) {
	atlas := &HashMap_OA{
		buckets:  make([]Entry, 4),
		capacity: 4,
	}

	// Test Collision & Put
	atlas.Put("alpha", 1)
	atlas.Put("beta", 2)
	atlas.Put("gamma", 3)

	// Test Delete (Tombstone creation)
	atlas.Delete("beta")

	// Test Get through Tombstone
	val, found := atlas.Get("gamma")
	if !found || val != 3 {
		t.Errorf("Failed to find 'gamma' after deleting 'beta'. Got %d", val)
	}

	// Test Tombstone Recycling
	atlas.Put("delta", 4)
	
	// Verify "delta" exists
	if v, ok := atlas.Get("delta"); !ok || v != 4 {
		t.Error("Failed to Put/Get 'delta' into recycled tombstone slot")
	}

	// Test Resize
	atlas.Put("omega", 5)
	if atlas.capacity <= 4 {
		t.Error("Expected Open Addressing map to resize")
	}

	// Verify all items exist after resize
	keys := []string{"alpha", "gamma", "delta", "omega"}
	for _, k := range keys {
		if _, ok := atlas.Get(k); !ok {
			t.Errorf("Lost key %s after resize", k)
		}
	}
}