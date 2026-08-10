// phase 1, part b
// run: go run ./phase-1/part-b

package main

import (
	"fmt"
	"maps"
)

// though function tries to change the element of the array
// it cannot since it is operating on the copy of the array too
func modify(arr [3]int) {
	arr[0] = 100
}

func main() {
	// ------------- arrays -------------
	fmt.Println("------------- arrays -------------")

	// assignment and function calls copy the whole arrays
	a := [3]int{1, 2, 3}
	b := a			// 'b' copies the whole array 'a'
	b[0] = 99
	fmt.Println(a)	// 'a' remains unchanged: [1,2,3]
	fmt.Println(b)	// 'b' becomes: [99, 2, 3]

	// counting the number of elements via [...]
	f := [...]int{100, 3: 400, 500}
	fmt.Println("idx: ", f)

	modify(a)
	// 'a' remains unchanged
	fmt.Println("after modification: ", a)

	// comparing two arrays with == operation
	c := [3]int{1, 2, 3}
	d := [3]int{1, 2, 3}
	e := [3]int{4, 5, 6}

	fmt.Println(c == d)
	fmt.Println(d == e)
	fmt.Println(e == c)

	// ------------- slices -------------
	fmt.Println("------------- slices -------------")

	// uninitialised slices have no elements and their length = 0
	var s []string
	fmt.Println("uninit", s, s == nil, len(s) == 0)

	// creating a slice with the make keyword
	s = make([]string, 3)
	fmt.Println("emp: ", s, "len: ", len(s), "cap: ", cap(s))

	// setting and getting slice elements
	s[0] = "a"
	s[1] = "b"
	s[2] = "c"
	fmt.Println("set: ", s)
	fmt.Println("get: ", s[2])

	// we can enter new values into a slice via the append method
	// dynamic sizing in slices compared to arrays
	s = append(s, "d")
	s = append(s, "e", "f")
	fmt.Println("apd: ", s)

	// one slice can also be copied into another slice
	x := make([]string, len(s))
	copy(x, s)
	fmt.Println("cpy: ", x)

	// slice data types come with a slice operator
	l := s[2:5]			// slice[low:high], index [high] gets excluded
	fmt.Println("sl1: ", l)

	l = s[:5]			// slice[:high], index [high] gets excluded
	fmt.Println("sl2: ", l)

	l = s[2:]			// slice[low:], index[low] gets included
	fmt.Println("sl3: ", l)

	// ------------- maps -------------
	fmt.Println("------------- maps -------------")
	// initialising a map
	// m1 := make(map[key-type]val-type)
	m2 := map[string]int{"a": 1, "b": 2}
	fmt.Println("map 2: ", m2)

	var1 := m2["a"]
	var2 := m2["zzz"]

	fmt.Println("variable 1: ", var1)
	fmt.Println("variable 2: ", var2)

	// deleting an element from the map
	delete(m2, "a")
	fmt.Println("after deleting an element from the map, m2: ", m2)

	// calculating the length of the map
	map_length := len(m2)
	fmt.Println("map length: ", map_length)

	// way of clearing the whole map
	clear(m2)
	fmt.Println("after clearing the complete m2: ", m2)

	// comparing two maps
	n1 := map[string]int{"foo": 1, "bar": 2}
	n2 := map[string]int{"foo": 1, "bar": 2}
	if maps.Equal(n1, n2) {
		fmt.Println("n1 == n2")
	}

	// ------------- sets -------------
	fmt.Println("------------- sets -------------")

	// making a zoo which acts like a set of strings
	zoo := map[string]struct{}{}

	// adding some elements to the set
	zoo["elephant"] = struct{}{}
	zoo["tiger"] = struct{}{}
	zoo["owl"] = struct{}{}
	zoo["lion"] = struct{}{}

	fmt.Println("initial zoo: ", zoo)

	// adding a new member to the set
	zoo["bear"] = struct{}{}

	// adding an existing member to the set
	zoo["lion"] = struct{}{}

	// removing a member from the set
	delete(zoo, "owl")

	fmt.Println("final zoo: ", zoo)
}
