package main

type customError struct {
	msg string
}

func (e *customError) Error() string {
	return e.msg
}

func test() *customError {
	// ... do something
	return nil
}

func main() {
	var err error
	err = test()
	if err != nil {
		println("error")
		return
	}
	println("ok")
}


/*
Вывод: error
Причина: Интерфейс err получает тип *customError и значение nil из test(), что делает err != nil. 
Условие if срабатывает, выводя "error".
*/