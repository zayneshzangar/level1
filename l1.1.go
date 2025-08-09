package main

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
