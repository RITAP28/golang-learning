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