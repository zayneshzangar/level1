package main

import (
	"fmt"
	"math"
)

// GroupTemperatures группирует температуры по диапазонам с шагом 10 градусов.
func GroupTemperatures(temps []float64) map[int][]float64 {
	groups := make(map[int][]float64)
	for _, temp := range temps {
		// Находим нижнюю границу диапазона: floor(temp / 10) * 10
		rangeStart := int(math.Floor(temp/10)) * 10
		groups[rangeStart] = append(groups[rangeStart], temp)
	}
	return groups
}

func main() {
	temps := []float64{-25.4, -27.0, 13.0, 19.0, 15.5, 24.5, -21.0, 32.5}
	groups := GroupTemperatures(temps)
	fmt.Println(groups)
}
	