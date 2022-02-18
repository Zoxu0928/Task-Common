package main

import "fmt"

func main() {
	a := get()
	for i, v := range a {
		fmt.Println(i, v)
	}
}

func get() []string {
	return nil
}
