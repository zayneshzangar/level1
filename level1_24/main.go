package main

import (
	"fmt"
	"math"
)

type Point struct {
	x float64
	y float64
}

func NewPoint(x, y float64) *Point {
	return &Point{x: x, y: y}
}

func (p *Point) Distance(other *Point) float64 {
	dx := other.x - p.x
	dy := other.y - p.y
	return math.Sqrt(math.Pow(dx, 2) + math.Pow(dy, 2))
}

func main() {
	p1 := NewPoint(0.0, 0.0)
	p2 := NewPoint(3.0, 4.0)

	distance := p1.Distance(p2)
	fmt.Printf("Distance between (%.2f, %.2f) and (%.2f, %.2f) = %.2f\n", p1.x, p1.y, p2.x, p2.y, distance)


	p3 := NewPoint(1.0, 1.0)
	p4 := NewPoint(4.0, 5.0)
	distance = p3.Distance(p4)
	fmt.Printf("Distance between (%.2f, %.2f) and (%.2f, %.2f) = %.2f\n", p3.x, p3.y, p4.x, p4.y, distance)
}