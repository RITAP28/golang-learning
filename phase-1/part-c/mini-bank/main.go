package main

import "fmt"


// it's a small mini bank account system in golang
// it's a CLI project having features like: create account, deposit, withdraw, transfer, show account and exit


type Account struct {
	number		int
	owner		string
	balance		float64
}

func main() {
	fmt.Println("here i am")
}