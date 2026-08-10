// building a hash map from scratch via separate chaining method

package main

import (
	"fmt"
)

// node struct represents the link in the linked list chain
type Node struct {
	key			string
	value		int
	next		*Node		// the part which is pointing to the next node
}

type HashMap struct {
	buckets		[]*Node		// array of nodes arranged in the form of linked list
	count		int			// total number of buckets available
	capacity	int			// total number of key-value pairs saved inside the hashmap
}

func NewHashMap(initialCapacity int) *HashMap {
	if initialCapacity <= 0 {
		panic("capacity must be positive")
	}

	return &HashMap{
		buckets: make([]*Node, initialCapacity),
		capacity: initialCapacity,
	}
}

func (m *HashMap) resize() {
	// doubling the old capacity
	newCapacity := m.capacity * 2

	// making a new slice of buckets with the new capacity
	newBuckets := make([]*Node, newCapacity)

	// iterating through the old buckets to find the existing keys and values
	for i:=0; i<m.capacity; i++ {
		current := m.buckets[i]

		// ---------- important thinking here ----------
		for current != nil {
			next := current.next

			// calculating the new index for this node
			newIndex := FNV1aHash([]byte(current.key)) % uint32(newCapacity)

			// prepending the node into the NEW bucket
			// reusing the existing node struct to save memory
			current.next = newBuckets[newIndex]
			newBuckets[newIndex] = current

			// moving to the next node in the old chain
			current = next
		}
	}

	// updating the hashmap fields to use the new data
	m.buckets = newBuckets
	m.capacity = newCapacity
}

// creating a FNV-1a algorithm for hashing
// used in scenarios where performance is critical, like hash-tables and checksums
func FNV1aHash(data []byte) uint32 {
	// FNV parameters
	const (
		FNVPrime32 = 0x01000193
		FNVOffset32 = 0x811c9dc5
	)

	hash := FNVOffset32
	for _, b := range data {
		hash ^= int(b)
		hash *= FNVPrime32
	}

	return uint32(hash)
}

func (m *HashMap) Put(key string, value int) {
	// hashing the key
	hashedKey := FNV1aHash([]byte(key))
	fmt.Println("hashed key: ", hashedKey)

	// mapping the hash to an index
	index := hashedKey % uint32(m.capacity)
	fmt.Println("index calculated: ", index)

	// checking the existence of the key
	current := m.buckets[index]
	fmt.Println("current: ", current)

	// after finding the specific bucket, we start traversing the chain
	for current != nil {
		if current.key == key {
			fmt.Println("existing found: ", current.key)
			// since the key is found, update the value and exit
			current.value = value
		}

		current = current.next
	}

	// this step is when the current key does not exist
	// then we create a new node and add it to the front of linked list
	// by pointing this node to the head of that linked list, hence prepending
	newNode := &Node{
		key: key,
		value: value,
		next: m.buckets[index],
	}

	// setting the bucket's head to our new node created above
	m.buckets[index] = newNode
	fmt.Println("new node created")

	// incrementing the count of items
	m.count++
	fmt.Println("count: ", m.count)

	// checking if we need to resize the bucket
	// resizing happens when the load factor becomes > than 0.75
	// this is done to reduce the chances of collisions
	var load_factor = float64(m.count) / float64(m.capacity)
	if load_factor > 0.75 {
		// creating a new slice, doubling the capacity
		// re-hash every single item in the old hashmap
		// and put them in new buckets in the new hashmap

		m.resize()
	}
}

func (m *HashMap) Get(key string) (int, bool) {
	hashedKey := FNV1aHash([]byte(key))
	index := hashedKey % uint32(m.capacity)

	current := m.buckets[index]
	for current != nil {
		if current.key == key {
			fmt.Println("current value: ", current.value)
			return current.value, true
		}

		current = current.next
	}

	fmt.Println("no existing key found")
	return 0, false
}

// implementing delete function -> a standard linked list node removal technique
func (m *HashMap) Delete(key string) {
	hashedKey := FNV1aHash([]byte(key))
	index := hashedKey % uint32(m.capacity)

	current := m.buckets[index]
	var prev *Node

	for current != nil {
		if current.key == key {
			if prev != nil {
				prev.next = current.next
			} else {
				m.buckets[index] = current.next
			}			
			return
		}

		prev = current
		current = current.next
	}
}

func main() {
	fmt.Println("------------------- Separate Chaining Method -------------------")
	// 1. Initialize with a SMALL capacity to force a resize earlier
	myMap := &HashMap{
		buckets:  make([]*Node, 5),
		capacity: 5,
	}

	fmt.Println("TEST 1: Basic Insertion")
	myMap.Put("apple", 10)
	myMap.Put("banana", 20)
	myMap.Put("cherry", 30)
	myMap.DebugPrint()

	fmt.Println("\nTEST 2: Updating existing key")
	myMap.Put("apple", 99) // Should overwrite 10
	val, _ := myMap.Get("apple")
	fmt.Printf("Apple value (should be 99): %d\n", val)

	fmt.Println("\nTEST 3: Get non-existent key")
	val, exists := myMap.Get("dragonfruit")
	fmt.Printf("Exists: %v, Value: %d\n", exists, val)

	fmt.Println("\nTEST 4: Forcing Resize")
	// Current capacity is 5. Load factor 0.75 of 5 is 3.75.
	// We already have 3 items. Adding 2 more should trigger resize.
	myMap.Put("date", 40)
	myMap.Put("elderberry", 50)
	myMap.Put("fig", 60)
	
	myMap.DebugPrint() // Capacity should now be 10

	fmt.Println("\nTEST 5: Verify data after resize")
	// Make sure we can still find the original items
	keys := []string{"apple", "banana", "cherry", "date", "elderberry", "fig"}
	allFound := true
	for _, k := range keys {
		if _, exists := myMap.Get(k); !exists {
			fmt.Printf("BUG: Lost key %s during resize!\n", k)
			allFound = false
		}
	}
	if allFound {
		fmt.Println("All keys successfully migrated during resize!")
	}

	fmt.Println()
	fmt.Println()
	fmt.Println()

	fmt.Println("------------------- Linear Probing Method -------------------")
	// 1. Initialize with a very small capacity (4) 
	// This makes it very easy to see collisions and trigger a resize.
	atlas := &HashMap_OA{
		buckets:  make([]Entry, 4),
		capacity: 4,
	}

	fmt.Println("--- Phase 1: Basic Insertion & Collisions ---")
	atlas.Put("alpha", 1)
	atlas.Put("beta", 2)
	atlas.Put("gamma", 3) 
	// At this point, 3/4 is 75%. The next Put should trigger a resize.
	atlas.DebugPrint()

	fmt.Println("\n--- Phase 2: Deletion & Tombstones ---")
	// Let's delete the middle item to create a tombstone
	atlas.Delete("beta")
	atlas.DebugPrint() 
	
	// Verify "gamma" is still reachable (this proves you can jump over tombstones)
	val, found := atlas.Get("gamma")
	fmt.Printf("Get 'gamma' after 'beta' deleted: Value=%d, Found=%v\n", val, found)

	fmt.Println("\n--- Phase 3: Recycling Tombstones ---")
	// If we put a new key, it should take the spot where "beta" used to be
	atlas.Put("delta", 4)
	atlas.DebugPrint()

	fmt.Println("\n--- Phase 4: Forced Resize & Cleanup ---")
	// Adding "omega" should definitely trigger a resize if it hasn't happened yet.
	// Resize should remove all tombstones and double the capacity.
	atlas.Put("omega", 5)
	atlas.Put("epsilon", 6)
	atlas.DebugPrint()

	fmt.Println("\n--- Phase 5: Final Validation ---")
	// Final check: Can we find the very first and very last items?
	aVal, aFound := atlas.Get("alpha")
	oVal, oFound := atlas.Get("omega")
	fmt.Printf("Alpha: %d (Found: %v)\n", aVal, aFound)
	fmt.Printf("Omega: %d (Found: %v)\n", oVal, oFound)
}
