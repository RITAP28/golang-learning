// building a hash map from scratch via open addressing method
// data lives in one flat array instead of linked lists

package main

// defining the constants for the state of a slot
const (
	empty		= iota
	occupied
	tombstone
)

// Entry is a single slot in the flat array
type Entry struct {
	key			string
	value		int
	state		byte	// empty, occupied or tombstone
}

// HashMap_OA = hashmap via open addressing
type HashMap_OA struct {
	buckets		[]Entry
	count		int
	capacity	int
}

func NewHashMap_OA(initialCapacity int) *HashMap_OA {
	if initialCapacity <= 0 {
		panic("initial capacity must be positive")
	} 

	return &HashMap_OA{
		buckets: make([]Entry, initialCapacity),
		capacity: initialCapacity,
	}
}

func (m *HashMap_OA) resize() {
	// double the size of the array
	// then allot new slots for the existing keys
	// and those new slots will be designated as occupied and rest as empty
	// no more tombstone states shall be there after resizing
	newCapacity := m.capacity * 2

	newBuckets := make([]Entry, newCapacity)

	for i:=0; i<m.capacity; i++ {
		current := m.buckets[i]

		switch current.state {
		// only care about the occupied state
		case occupied:
			newIndex := int(FNV1aHash([]byte(current.key)) % uint32(newCapacity))

			// we need to find the first empty spot in the new array
			for {
				if newBuckets[newIndex].state == empty {
					newBuckets[newIndex] = current
					break
				}

				// if this spot is taken in the new array, then move to the next spot
				newIndex = (newIndex + 1) % newCapacity
			}
			
		// dont care about the following state conditions
		case empty:
		case tombstone:
		}
	}

	// m.mount stays the same because we just moved the occupied items to the new array
	// assigning the old hashmap new and updated credentials
	m.buckets = newBuckets
	m.capacity = newCapacity
}

func (m *HashMap_OA) Put(key string, value int) {
	// checking load factor and resize if necessary
	if float64(m.count) / float64(m.capacity) >= 0.75 {
		m.resize()
	}

	index := int(FNV1aHash([]byte(key)) % uint32(m.capacity))
	firstTombstoneIdx := -1

	// looping through every slot at most once
	for i:=0; i<m.capacity; i++ {
		currIdx := (index + i) % m.capacity
		current := &m.buckets[currIdx]

		switch current.state {
		case occupied:
			if current.key == key {
				current.value = value	// updating existing key
				return
			}
			// collision: key doesn't match the current slot's key
			// keep traversing

		case empty:
			targetIndex := currIdx
			if firstTombstoneIdx == 1 {
				targetIndex = firstTombstoneIdx
			}

			m.buckets[targetIndex] = Entry{
				key: key,
				value: value,
				state: occupied,
			}

			m.count++
			return

		case tombstone:
			// remembering the first tombstone for recycling
			if firstTombstoneIdx == -1 {
				firstTombstoneIdx = currIdx
			}

			// searching further
		}

		index = (index + 1) % m.capacity
	}
}

func (m *HashMap_OA) Get(key string) (int, bool) {
	// hashing the key to get the index
	// if we get the required key at the first time, then return it with the value and true
	// if we don't, then traverse the next index
	// if the next index is of the state -> "empty", then stop and return with false
	// but if the next index with state "tombstone", then skip to the next one

	index := int(FNV1aHash([]byte(key)) % uint32(m.capacity))

	// start with the index calculated and then traverse until found
	for i:=0; i<m.capacity; i++ {
		current := m.buckets[index]

		switch current.state {
		case occupied:
			if current.key == key {
				return current.value, true
			}

			// if the keys don't match, just skip it to the next index
		case empty:
			// the end of the search, meaning no similar keys are found
			return 0, false
		case tombstone:
			// skip to the next index and check again
			// missed edge-case: if index starts at the very last slot of the array and the state is tombstone
			// then, index = index + 1 will go out of bounds, to keep it inside, % modulo is used
		}

		index = (index + 1) % m.capacity
	}

	// in the case, when there are no similar keys, return 0 with false
	return 0, false
}

func (m *HashMap_OA) Delete(key string) bool {
	// hashing the key to get the index
	// go to that index and consider the three states
	// we need it occupied, so at least we can check the key
	// if the key doesn't match, we skip to the next index, else we delete the key and turn it into tombstone
	// if the state is empty or tombstone, we skip to the next one
	// if we are at the last slot of the array and find it useless, then stop the operation, meaning key doesn't exist

	index := int(FNV1aHash([]byte(key)) % uint32(m.capacity))

	for i:=0; i<m.capacity; i++ {
		current := &m.buckets[index]
		switch current.state {
		case occupied:
			// if the keys match, then set up the state as tombstone
			if current.key == key {
				current.state = tombstone

				// forgot: decrementing the count
				m.count--
				return true
			}

			// if the keys don't match, then skip to the next index
		case empty:
			// our stop sign
			return false
		case tombstone:
			// skip to the next index
		}

		// careful for out of bounds panic
		index = (index + 1) % m.capacity
	}

	return false
}



