package main

import (
    "fmt"
    "math"
    "math/big"
)

// Calculator представляет структуру для арифметических операций.
type Calculator struct {
    a, b int64
    useBig bool
    aBig, bBig *big.Int
}

// NewCalculator создаёт новый калькулятор с проверкой входных данных.
func NewCalculator(a, b int64) (*Calculator, error) {
    if a <= 1<<20 || b <= 1<<20 {
        return nil, fmt.Errorf("both numbers must be > 2^20 (1,048,576)")
    }
    maxInt64 := int64(math.MaxInt64)
    useBig := a > maxInt64/2 || b > maxInt64/2 || a*b > maxInt64 // Проверка на переполнение

    c := &Calculator{
        a:      a,
        b:      b,
        useBig: useBig,
    }
    if useBig {
        c.aBig = big.NewInt(a)
        c.bBig = big.NewInt(b)
    }
    return c, nil
}

// Add выполняет сложение.
func (c *Calculator) Add() interface{} {
    if !c.useBig {
        return c.a + c.b
    }
    return new(big.Int).Add(c.aBig, c.bBig)
}

// Subtract выполняет вычитание.
func (c *Calculator) Subtract() interface{} {
    if !c.useBig {
        return c.a - c.b
    }
    return new(big.Int).Sub(c.aBig, c.bBig)
}

// Multiply выполняет умножение.
func (c *Calculator) Multiply() interface{} {
    if !c.useBig {
        return c.a * c.b
    }
    return new(big.Int).Mul(c.aBig, c.bBig)
}

// Divide выполняет деление.
func (c *Calculator) Divide() (interface{}, error) {
    if !c.useBig {
        if c.b == 0 {
            return nil, fmt.Errorf("division by zero")
        }
        return c.a / c.b, nil
    }
    if c.bBig.Cmp(big.NewInt(0)) == 0 {
        return nil, fmt.Errorf("division by zero")
    }
    return new(big.Int).Div(c.aBig, c.bBig), nil
}

func main() {
    // Примеры чисел > 2^20
    testCases := []struct {
        a, b int64
        desc string
    }{
        {1 << 21, 1 << 22, "Small numbers (within int64)"}, // 2^21, 2^22
        {1 << 40, 1 << 40, "Large numbers (beyond int64)"}, // 2^40, 2^40
    }

    for _, tc := range testCases {
        calc, err := NewCalculator(tc.a, tc.b)
        if err != nil {
            fmt.Println("Error:", err)
            continue
        }

        fmt.Printf("\nTest case: %s\n", tc.desc)
        fmt.Printf("a = %d, b = %d\n", tc.a, tc.b)
        switch result := calc.Add().(type) {
        case int64:
            fmt.Printf("a + b = %d\n", result)
        case *big.Int:
            fmt.Printf("a + b = %s\n", result.String())
        }
        switch result := calc.Subtract().(type) {
        case int64:
            fmt.Printf("a - b = %d\n", result)
        case *big.Int:
            fmt.Printf("a - b = %s\n", result.String())
        }
        switch result := calc.Multiply().(type) {
        case int64:
            fmt.Printf("a * b = %d\n", result)
        case *big.Int:
            fmt.Printf("a * b = %s\n", result.String())
        }
        result, err := calc.Divide()
        if err != nil {
            fmt.Println("Error:", err)
        } else {
            switch r := result.(type) {
            case int64:
                fmt.Printf("a / b = %d\n", r)
            case *big.Int:
                fmt.Printf("a / b = %s\n", r.String())
            }
        }
        fmt.Printf("Using big numbers? %v\n", calc.useBig)
    }
}
