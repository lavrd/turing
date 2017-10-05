package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

// type for command
type command struct {
	symbol, transition string
	state              int
}

// type for write log to file
type wLog struct {
	tapeBefore, tapeAfter, headBefore            string
	headIndexBefore, headIndexAfter, stateBefore int
	cmd                                          *command
}

func main() {

	var (
		alphabetPath = flag.String("alph", "./files/alphabet", "path to alphabet file")
		tapePath     = flag.String("tape", "./files/tape", "path to tape file")
		programPath  = flag.String("prog", "./files/program", "path to program file")
		logsPath     = flag.String("logs", "./files/logs", "logs will be saved to specified file")
		verbose      = flag.Bool("v", false, "verbose output")

		f *os.File
	)

	flag.Parse()

	if *alphabetPath == "" {
		log.Fatalf("Incorrect alphabet file path: [%s]", *alphabetPath)
	} else if *tapePath == "" {
		log.Fatalf("Incorrect tape file path: [%s]", *tapePath)
	} else if *programPath == "" {
		log.Fatalf("Incorrect program file path: [%s]", *programPath)
	}

	alphabet, err := loadAlphabet(*alphabetPath)
	if err != nil {
		log.Fatalf("loadAlphabet: %s", err)
	}

	tape, err := loadTape(*tapePath, alphabet)
	if err != nil {
		log.Fatalf("loadTape: %s", err)
	}

	program, err := loadProgram(*programPath, alphabet)
	if err != nil {
		log.Fatalf("loadProgram: %s", err)
	}

	if *verbose {
		err := prepareLogsFile(*logsPath, alphabet, tape, program)
		if err != nil {
			log.Fatalf("prepareLogsFile: %s", err)
		}

		f, err = os.OpenFile(*logsPath, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
		if err != nil {
			log.Fatalf("main: open file: [%s] err: [%s]", logsPath, err)
		}
		defer f.Close()
	}

	err = run(tape, program, *verbose, f)
	if err != nil {
		log.Fatalf("run: %s", err)
	}
}

// run starts machine
func run(tape []string, program *map[int]map[string]*command, verbose bool, f *os.File) error {

	var (
		// current head index
		i = 0
		// current state
		state = 0
		// to check for end of tape to increase it
		iMax = len(tape)
		// for write logs
		wl = &wLog{}
	)

	for {
		// save before for logs
		wl.stateBefore = state
		wl.headBefore = tape[i]
		wl.headIndexBefore = i
		wl.tapeBefore = strings.Join(tape, "")

		// get new command
		cmd := (*program)[state][tape[i]]
		// and save it
		wl.cmd = cmd

		// get new state
		state = cmd.state

		// change head symbol
		tape[i] = cmd.symbol
		// and save it
		wl.tapeAfter = strings.Join(tape, "")

		// get transaction
		switch cmd.transition {
		case ">":
			// move head right
			i++
			// if end of tape increase it
			if i == iMax {
				tape = append(tape, "_")
				// new end of tape
				iMax++
			}
			break
		case "<":
			// move head left
			i--
			// that index doesn't out of range we increase tape from start
			if i == -1 {
				tape = append([]string{"_"}, tape...)
				i++
			}
			break
		case "!":
			// stop machine
			return wl.log(verbose, f)
		}

		// save head index after
		wl.headIndexAfter = i

		// write log
		err := wl.log(verbose, f)
		if err != nil {
			return errors.New(fmt.Sprintf("log: [%s]", err))
		}
	}
}

// log write log to logs file
func (wl *wLog) log(verbose bool, f *os.File) error {

	// if verbose output
	if verbose {
		_, err := f.WriteString("\n" + wl.tapeBefore)
		if err != nil {
			return errors.New(fmt.Sprintf("write string err: [%s]", err))
		}
		_, err = f.WriteString("\n" + strings.Repeat(" ", wl.headIndexBefore) + "^")
		if err != nil {
			return errors.New(fmt.Sprintf("write string err: [%s]", err))
		}
		_, err = f.WriteString("\n" + fmt.Sprintf("%d%s->%d%s%s", wl.stateBefore,
			wl.headBefore, wl.cmd.state, wl.cmd.symbol, wl.cmd.transition))
		if err != nil {
			return errors.New(fmt.Sprintf("write string err: [%s]", err))
		}
		_, err = f.WriteString("\n" + wl.tapeAfter)
		if err != nil {
			return errors.New(fmt.Sprintf("write string err: [%s]", err))
		}
		_, err = f.WriteString("\n" + strings.Repeat(" ", wl.headIndexAfter) + "^\n")
		if err != nil {
			return errors.New(fmt.Sprintf("write string err: [%s]", err))
		}
	}

	return nil
}

// prepareLogsFile prepare logs file
func prepareLogsFile(logsPath string, alphabet []string, tape []string, program *map[int]map[string]*command) error {

	var (
		f   *os.File
		err error
	)

	// create if not exists
	f, err = os.Create(logsPath)
	if err != nil {
		if os.IsExist(err) {
			return f.Close()
		}

		return errors.New(fmt.Sprintf("create file: [%s] err: [%s]", logsPath, err))
	}

	// write alphabet, tape, program and date
	f.WriteString("\n----------\n")
	f.WriteString("----------")

	f.WriteString("\n\n" + fmt.Sprintf("Date: %s\n", time.Now().Local()))
	f.WriteString(fmt.Sprintf("Alphabet: %s\n", alphabet))
	f.WriteString(fmt.Sprintf("Tape: %s\n", tape))
	f.WriteString("Program:\n")

	for state := range *program {
		for head := range (*program)[state] {
			cmd := (*program)[state][head]
			f.WriteString(fmt.Sprintf("\t%d%s->%d%s%s\n", state, head, cmd.state, cmd.symbol, cmd.transition))
		}
	}

	f.WriteString("\n----------\n")
	f.WriteString("----------\n")

	return f.Close()
}

// loadTape load tape from file
func loadTape(tapePath string, alphabet []string) ([]string, error) {

	var tape []string

	f, err := os.Open(tapePath)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("open file: [%s] err: [%s]", tapePath, err))
	}

	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		tape = strings.Split(scanner.Text(), "")
	}

	// check for unknown characters
	for _, t := range tape {
		if !strings.Contains(strings.Join(alphabet, ""), t) {
			return nil, errors.New(fmt.Sprintf("unknown character: [%v]", t))
		}
	}

	return tape, f.Close()
}

// loadAphabet load alphabet from file
func loadAlphabet(alphabetPath string) ([]string, error) {

	var alphabet []string

	f, err := os.Open(alphabetPath)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("open file: [%s] err: [%s]", alphabetPath, err))
	}

	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		alphabet = strings.Split(scanner.Text(), " ")
	}

	return alphabet, f.Close()
}

// loadProgram load program from file
func loadProgram(programPath string, alphabet []string) (*map[int]map[string]*command, error) {

	var (
		program = make(map[int]map[string]*command)
		// line index, current state and new state from command
		lineIndex, cState, nState int
		// current file line
		line []string
		// command transition, current head, command symbol
		transition, head, symbol string
	)

	f, err := os.Open(programPath)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("open file: [%s] err: [%s]", programPath, err))
	}

	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line = strings.Split(scanner.Text(), "")
		lineIndex++

		// skip empty and comment line
		if len(line) == 0 || line[0] == "#" {
			continue
		}

		// if line len != 7 => line is incorrect
		if len(line) != 7 {
			return nil, errors.New(fmt.Sprintf("incorrect line [%s] line: [%d]", line, lineIndex))
		}

		// get current state and convert to int
		cState, err = strconv.Atoi(line[0])
		if err != nil {
			return nil, errors.New(fmt.Sprintf("convert current state err: [%s] line: [%d]", err, lineIndex))
		}

		// get current head and check that it is in alphabet
		head = line[1]
		if !strings.Contains(strings.Join(alphabet, ""), head) {
			return nil, errors.New(fmt.Sprintf("unknown character: [%s] line: [%d]", head, lineIndex))
		}

		// get transition and check
		transition = line[6]
		if transition != "<" && transition != ">" && transition != "!" {
			return nil, errors.New(fmt.Sprintf("incorrect symbol: [%s] line: [%d]", transition, lineIndex))
		}

		// get new state and convert to int
		nState, err = strconv.Atoi(line[4])
		if err != nil {
			return nil, errors.New(fmt.Sprintf("convert new state err: [%s] line: [%d]", err, lineIndex))
		}

		// get command symbol and check that it is in alphabet
		symbol = line[5]
		if !strings.Contains(strings.Join(alphabet, ""), symbol) {
			return nil, errors.New(fmt.Sprintf("unknown character: [%s] line: [%d]", symbol, lineIndex))
		}

		// added command to program and init second map if not exists
		if _, ok := program[cState]; !ok {
			program[cState] = make(map[string]*command)
		}
		program[cState][head] = &command{
			state:      nState,
			transition: transition,
			symbol:     symbol,
		}
	}

	// check for empty program
	if len(program) == 0 {
		return nil, errors.New(fmt.Sprint("load program err: empty program"))
	}

	return &program, f.Close()
}
