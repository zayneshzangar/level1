package main

import "fmt"

type Human struct {
	name string
	age  int
}

func (h *Human) SetName(name string) {
	h.name = name
}

func (h *Human) SetAge(age int) {
	h.age = age
}

func (h *Human) GetName() string {
	return h.name
}

func (h *Human) GetAge() int {
	return h.age
}

type Action struct {
	Human
}

func main() {
	a := Action{}
	a.SetName("Zangar")
	a.SetAge(31)

	fmt.Printf("My name is %s and I'm %d years old!\n", a.GetName(), a.GetAge())
}
