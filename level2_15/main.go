package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

// Minishell — простой Unix-подобный шелл
type Minishell struct {
	scanner *bufio.Scanner // считывает команды пользователя
}

// NewMinishell создаёт новый экземпляр шелла.
func NewMinishell() *Minishell {
	return &Minishell{scanner: bufio.NewScanner(os.Stdin)}
}

// Run — основной цикл шелла
func (sh *Minishell) Run() {
	// Канал для сигналов (например, Ctrl+C)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT)

	// Отдельная горутина для обработки прерываний
	go func() {
		for range sigCh {
			fmt.Println("\nInterrupted")
		}
	}()

	// Основной REPL (Read-Eval-Print Loop)
	for {
		fmt.Print("$ ") // выводим приглашение

		// Читаем строку
		if !sh.scanner.Scan() {
			if sh.scanner.Err() == nil {
				// EOF (Ctrl+D)
				fmt.Println("EOF received, exiting...")
				return
			}
			// Ошибка при вводе
			fmt.Fprintf(os.Stderr, "Error reading input: %v\n", sh.scanner.Err())
			continue
		}

		// Получаем введённую строку
		line := strings.TrimSpace(sh.scanner.Text())
		if line == "" {
			continue
		}

		// Выполняем команду
		if err := sh.execute(line); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
	}
}

// execute — выполняет строку с учетом логических операторов (&&, ||) и пайпов
func (sh *Minishell) execute(line string) error {
	// Разбиваем строку по && и ||
	tokens := splitByLogical(line)

	var lastStatus error // хранит статус последней команды
	for i := 0; i < len(tokens); i++ {
		op, cmdLine := tokens[i].op, tokens[i].cmd

		// Если стоит &&, но предыдущая команда упала — пропускаем
		if op == "&&" && lastStatus != nil {
			continue
		}
		// Если стоит ||, но предыдущая команда прошла успешно — пропускаем
		if op == "||" && lastStatus == nil {
			continue
		}

		// Разбиваем на пайплайн (ls | grep txt)
		pipeline := parsePipeline(cmdLine)

		// Проверяем встроенные команды (работают только без пайпов)
		if len(pipeline) == 1 {
			args := pipeline[0]
			if sh.handleBuiltin(args) {
				lastStatus = nil
				continue
			}
		}

		// Выполняем пайплайн
		err := runPipeline(pipeline)
		lastStatus = err
	}

	return lastStatus
}

// runPipeline — выполняет цепочку команд, соединённых через pipe
func runPipeline(cmds [][]string) error {
	var processes []*exec.Cmd
	var prevOut *os.File // хранит выход предыдущей команды

	for i, args := range cmds {
		// Создаём процесс
		cmd := exec.Command(args[0], args[1:]...)

		// Вход: либо stdin, либо предыдущий pipe
		if i == 0 {
			cmd.Stdin = os.Stdin
		} else {
			cmd.Stdin = prevOut
		}

		// Выход: либо stdout, либо новый pipe
		if i == len(cmds)-1 {
			cmd.Stdout = os.Stdout
		} else {
			r, w, err := os.Pipe()
			if err != nil {
				return err
			}
			cmd.Stdout = w
			prevOut = r

			// Закрываем writer в родителе, когда выйдем из функции
			defer w.Close()
		}

		cmd.Stderr = os.Stderr
		processes = append(processes, cmd)

		// Запускаем команду
		if err := cmd.Start(); err != nil {
			return err
		}

		// Закрываем stdout предыдущей команды в родителе
		if i > 0 {
			if f, ok := processes[i-1].Stdout.(*os.File); ok {
				f.Close()
			}
		}
	}

	// Ждём завершения всех процессов пайплайна
	var lastErr error
	for _, cmd := range processes {
		if err := cmd.Wait(); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

// parsePipeline — парсит пайплайн ("ls | grep x") в массив команд
func parsePipeline(line string) [][]string {
	parts := strings.Split(line, "|")
	var cmds [][]string
	for _, p := range parts {
		args := strings.Fields(strings.TrimSpace(p))
		if len(args) > 0 {
			cmds = append(cmds, args)
		}
	}
	return cmds
}

// структура для логических команд (&& и ||)
type logicalCmd struct {
	op  string
	cmd string
}

// splitByLogical — парсит строку на части по && и ||
func splitByLogical(line string) []logicalCmd {
	var result []logicalCmd
	cur := ""
	lastOp := ""
	for i := 0; i < len(line); {
		if strings.HasPrefix(line[i:], "&&") {
			result = append(result, logicalCmd{op: lastOp, cmd: strings.TrimSpace(cur)})
			cur = ""
			lastOp = "&&"
			i += 2
		} else if strings.HasPrefix(line[i:], "||") {
			result = append(result, logicalCmd{op: lastOp, cmd: strings.TrimSpace(cur)})
			cur = ""
			lastOp = "||"
			i += 2
		} else {
			cur += string(line[i])
			i++
		}
	}
	if strings.TrimSpace(cur) != "" {
		result = append(result, logicalCmd{op: lastOp, cmd: strings.TrimSpace(cur)})
	}
	return result
}

// handleBuiltin — встроенные команды (cd, pwd, echo, kill, ps)
func (sh *Minishell) handleBuiltin(args []string) bool {
	switch args[0] {
	case "cd":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "cd: missing argument")
			return true
		}
		if err := os.Chdir(args[1]); err != nil {
			fmt.Fprintf(os.Stderr, "cd: %v\n", err)
		}
		return true
	case "pwd":
		wd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "pwd: %v\n", err)
			return true
		}
		fmt.Println(wd)
		return true
	case "echo":
		fmt.Println(strings.Join(args[1:], " "))
		return true
	case "kill":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "kill: missing pid")
			return true
		}
		pid, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "kill: %v\n", err)
			return true
		}
		if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
			fmt.Fprintf(os.Stderr, "kill: %v\n", err)
		}
		return true
	case "ps":
		cmd := exec.Command("ps")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
		return true
	}
	return false
}

// main — точка входа
func main() {
	sh := NewMinishell()
	sh.Run()
}
