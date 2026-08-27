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

// the following struct Good occupies less memory related to Bad
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
}
