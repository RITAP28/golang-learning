package main

import "fmt"

func (m *HashMap) DebugPrint() {
	fmt.Printf("--- Map State (Count: %d, Capacity: %d) ---\n", m.count, m.capacity)
	for i, node := range m.buckets {
		if node == nil {
			continue
		}
		fmt.Printf("Bucket[%d]: ", i)
		curr := node
		for curr != nil {
			fmt.Printf("[%s: %d] -> ", curr.key, curr.value)
			curr = curr.next
		}
		fmt.Println("nil")
	}
}

func (m *HashMap_OA) DebugPrint() {
	fmt.Printf("--- Open Addressing Map (Count: %d, Cap: %d) ---\n", m.count, m.capacity)
	for i := 0; i < m.capacity; i++ {
		entry := m.buckets[i]
		stateStr := ""
		switch entry.state {
		case 0: stateStr = "empty"
		case 1: stateStr = "occupied"
		case 2: stateStr = "tombstone"
		}

		if entry.state == 1 {
			fmt.Printf("[%d]: %-10s -> %-10s: %d\n", i, stateStr, entry.key, entry.value)
		} else {
			fmt.Printf("[%d]: %-10s\n", i, stateStr)
		}
	}
}