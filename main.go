package main

import "fmt"

func main() {
	fmt.Println("START level 1.1")
	
	a := Action{}
	a.SetName("Zangar")
	a.SetAge(31)
	
	fmt.Printf("My name is %s and I'm %d years old!\n",a.GetName(), a.GetAge())
	fmt.Println("END level 1.1")
	fmt.Println("----------------------------------------")

	fmt.Println("START level 1.2")
	level_1_2()
	fmt.Println("END level 1.2")
	fmt.Println("----------------------------------------")
}
