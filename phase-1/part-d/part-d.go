// phase 1, part d
// run: go run ./phase-1/part-d

package main

import (
	"fmt"
	"math"
)

type Celcius float64
func (c Celcius) Fahrenheit() Celcius { return c*9/5 + 32 }

type Vertex struct { X, Y float64 }
// it's a method which is a function with a special receiver argument, which here is Vertex
func (v Vertex) Abs() float64 { return math.Sqrt(v.X*v.X + v.Y*v.Y) }
func (v *Vertex) Scale(f float64) {
	v.X = v.X * f
	v.Y = v.Y * f
	fmt.Printf("%f, %f", v.X, v.Y)
	fmt.Println()
}

type Counter struct { value	int }
func (c Counter) IncValue() {
	c.value++
	fmt.Println("c: ", c.value)		// here the value of the copy c counter is updated, not the original one
}	// here the receiver is a copy of the CounterTwo struct
func (c *Counter) IncPtr() { c.value++ }		// here the receiver is a pointer to CounterTwo struct, the original one

func main() {
	t := Celcius(100)
	fmt.Println(t.Fahrenheit())

	v := Vertex{4, 3}
	fmt.Println(v.Abs())
	v.Scale(10)
	fmt.Println("x: ", v.X)
	fmt.Println("y: ", v.Y)

	c := Counter{}		// c.value = 0 by default
	c.IncValue()
	fmt.Println("inc value: ", c.value)

	// the receiver mentioned in the IncPtr method is a pointer i.e. *Counter
	// here c is Counter and not *Counter
	// here, golang automatically turns c.IncPtr() into (&c).IncPtr() because c is addressable
	// this is automatic address-taking
	c.IncPtr()
	fmt.Println("inc pointer: ", c.value)

	// m := map[string]Counter{ "a": {},}
	// p := &m["a"]
}
