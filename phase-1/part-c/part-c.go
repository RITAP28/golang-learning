// phase 1, part c
// run: go run ./phase-1/part-c

package main

import (
	"encoding/json"
	"fmt"
	"unsafe"
	"time"
	"log"
	"os"
)

type Point struct { X, Y int }
func move(p Point) { p.X = 100 }

// --------------- new concept ---------------
// concepts: memory alignment and padding
// the following struct Bad occupies more memory
// due to bad alignment of variables
type Bad struct {
	a	bool	// 8 bytes (1 byte occupied + 7 bytes unoccupied)
	b	int64	// 8 bytes (64 bits occupied)
	c	bool	// 8 bytes (1 byte occupied + 7 bytes unoccupied)

	// total size: 24 bytes
}

// the following struct Good occupies less memory relative to Bad
// mainly because of better alignment of variables 
type Good struct {
	b	int64	// 8 bytes (64 bits occupied)
	a	bool	// 1 byte
	c	bool	// 1 byte (a + c + 6 bytes unoccupied)

	// total size: 16 bytes
}

// --------------- new concept ---------------
// concept: embedding (composition, not inheritance)
// in java, a 'Manager' is an 'Employee', it inherits everything from the parent
// class Employee { String name; void work() { System.out.println(name + "is working.") } }
// class Manager extends Employee { int teamSize; void manage() { System.out.println(name + "is managing" + teamSize + "people.") } }
// Manager m = new Manager();
// m.name = "Alice";
// m.work();
// If we have a function void Process(Employee e), we can pass it a Manager because Manager is an Employee. --> Polymorphism in Java
// in golang, it's different
// example 1
type Employee struct { Name	string }
func (e Employee) Work() { fmt.Println(e.Name, "is working.") }
type Manager struct {
	Employee	// anonymous embedded field
	teamSize	int
}

// example 2
type Animal struct { Name	string }
func (a Animal) Speak() string { return a.Name + "makes a sound." }
type Dog struct {
	Animal		// embedded: looks like a field without a name
	Breed	 	string
}

// --------------- new concept ---------------
// struct tags: they are string literals attached to the struct fields that provide metadata about how the fields should be handled by external packages at runtime
type User struct {
	ID		int		`json:"id" db:"user_id"`
	Email	string	`json:"email,omitempty" validate:"required,email"`
	Pass	string	`json:"-"`
}

// we have to use fields with capital letters in the struct when declaring json fields using encoding/json package
// otherwise the json printed to the terminal will become empty
// so in order to name the fields as 'name' instead of 'Name', we use Struct Tags
type UserTwo struct {
	Name		string		`json:"name"`
	// Password	string		`json:"password"`  // sensitive field, must be avoided to be printed, even when set
	Password	string		`json:"-"`
	Nicknames	[]string	`json:"nicknames,omitempty"` // omits this field from the version printed out if unset
	CreatedAt	time.Time	`json:"createdAt"`
}

// comparability of two structs
type A struct {
	ID		int
	Name	string
}
type Team struct {
	Name	string
	Players	[]string	// slices not comparable
}

// it's a pointer method
type MyInt int
func (input *MyInt) Double() {
	*input *= 2
}

type List struct {
	Val		int
	Next	*List
}

func (l *List) Len() int {
	if l == nil { return 0 }
	return 1 + l.Next.Len()
}

func sum(nums ...int) int {
	total := 0
	for _, n := range nums { total += n }
	return total
}

func zero(nums ...int) {
	for i := range nums { nums[i] = 0 }
}

func main() {
	fmt.Println("part c")

	p := Point{1,2}
	q := p			// full copy
	q.X = 99
	fmt.Println(p)
	fmt.Println(q)

	move(p)
	fmt.Println(p)

	// printing out the memory occupied by both the structs
	fmt.Println(unsafe.Sizeof(Bad{}), unsafe.Alignof(Bad{}))
	fmt.Println(unsafe.Sizeof(Good{}), unsafe.Alignof(Good{}))
	fmt.Println(unsafe.Offsetof(Bad{}.a), unsafe.Offsetof(Bad{}.b), unsafe.Offsetof(Bad{}.c))

	d := Dog{Animal{"Rex"}, "Lab"}
	fmt.Println(d.Name)
	fmt.Println(d.Speak())
	fmt.Println(d.Animal.Name)

	m := Manager{Employee{"Ritap"}, 10}
	m.Work()	// Method Promotion: m.Employee.Work() becomes m.Work()

	u := &UserTwo{
		Name: 		"Nick Olsen",
		Password: 	"IamGreat@07",
		// Nicknames:	[]string{"tallboy", "ben10"},
		CreatedAt: 	time.Now(),
	}

	out, err := json.MarshalIndent(u, "", " ")
	// out, err := json.Marshal(u)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	fmt.Println(string(out))

	fmt.Println()

	// testing the comparability concept
	fmt.Println("------- comparability concept -------")
	u1 := A{ID: 1, Name: "Alice"}
	u2 := A{ID:1, Name: "Alice"}
	u3 := A{ID:2, Name: "Ritap"}

	fmt.Println(u1 == u2)	// true since all the fields are equal
	fmt.Println(u2 == u3)	// false since all fields don't match

	// t1 := Team{Name: "Giants", Players: []string{"Alice"}}
	// t2 := Team{Name: "Giants", Players: []string{"Alice"}}
	// fmt.Println(t1 == t2) ---> uncomparable in nature since slices are not comparable

	fmt.Println()
 
	// testing the distinctions between zero-value and zero-sized structs
	fmt.Println("------- distinctions between zero-value and zero-sized structs -------")
	var a1 A	// zero-value: {0, ""}
	var a2 A	// zero-value: {0, ""}
	fmt.Println(&a1 == &a2)		// false since they have different addresses

	var e1 struct{}
	var e2 struct{}
	// the golang docs say that two distinct zero-sized structs 'may' have the same global 'zerobase' address
	fmt.Println(&e1 == &e2)		// not a guarantee for it to be true

	fmt.Println()

	// ---------- pointers in golang ----------
	fmt.Println("------- Pointers ------- ")
	x := MyInt(12)
	want := MyInt(24)

	p1 := &x
	p1.Double()	// entering the argument as a pointer to int

	// the variable x remains the same since it is passed into the function by the value
	// this process is called 'pass by value'
	fmt.Println(x)
	fmt.Println(want)
	if want != x {
		fmt.Println("x did not double it's value")
	}

	// var p2 int
	// fmt.Println(*p2)	// runtime panic: cannot dereference nil
	
	x2 := 100
	p2 := &x2			// points to x2
	fmt.Println(*p2)

	*p2 = 42
	fmt.Println(x2)

	var l *List				// nil pointer
	fmt.Println(l.Len()) 	// calling pointer-receiver method on a nil pointer does not cause panic
	// fmt.Println(*l) 		// causes runtime panic when dereferencing a nil pointer

	arr := [4]int{1,2,3,4}
	p3 := &arr
	p3[0] = 99
	fmt.Println("length of the array", len(p3))
	for idx, val := range p3 {
		fmt.Printf("%d = %d", idx, val)
		fmt.Println()
	}
	

	fmt.Println("------- Functions, Closures and Defer ------- ")
	fmt.Println("the sum calculated: ", sum(1, 2, 3))

	s := []int{1,2,3}
	fmt.Println("the sum calculated: ", sum(s...))		// sum(s...) passes s itself, not a copy

	// s... in the parameters sends the actual array and not the copy
	// so the real values inside the array are mutated
	zero(s...)
	fmt.Println("output of the zeroed function: ", s)
}
