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
)

type Alphabet []string
type Tape []string
type Program map[State]map[Head]*Command

type Command struct {
	nSymbol    string
	nState     State
	transition string
}
type Head string
type State int

type wLog struct {
	TapeBefore      Tape
	TapeAfter       Tape
	HeadIndexBefore int
	HeadIndexAfter  int
	HeadBefore      Head
	StateBefore     State
	Cmd             *Command
}

func main() {

	var (
		alphabetPath = flag.String("alph", "", "path to alphabet file")
		tapePath     = flag.String("tape", "", "path to tape file")
		programPath  = flag.String("prog", "", "path to program file")

		logsPath = flag.String("logs", "./logs.txt", "logs will be saved to specified file")
		verbose  = flag.Bool("v", false, "verbose output")
		//display  = flag.Bool("d", false, "display result")
	)

	flag.Parse()

	if *alphabetPath == "" {
		log.Fatalf("Incorrect alphabet file path: [%s]", *alphabetPath)
	} else if *tapePath == "" {
		log.Fatalf("Incorrect tape filepath: [%s]", *tapePath)
	} else if *programPath == "" {
		log.Fatalf("Incorrect program file path: [%s]", *programPath)
	}

	alphabet, err := LoadAlphabet(*alphabetPath)
	if err != nil {
		log.Fatal(err)
	}

	tape, err := LoadTape(*tapePath, alphabet)
	if err != nil {
		log.Fatal(err)
	}

	program, err := LoadProgram(*programPath, alphabet)
	if err != nil {
		log.Fatal(err)
	}

	if *verbose {
		err = prepareLogFile(*logsPath, alphabet, tape, program)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (wl *wLog) Log(verbose bool, logsPath string) error {

	var (
		f   *os.File
		err error
	)

	if verbose {
		f, err = os.OpenFile(logsPath, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
		if err != nil {
			return errors.New(fmt.Sprintf("Log: open file: [%s] err: [%s]", logsPath, err))
		}

		f.WriteString("\n --- ")
		f.WriteString("\n" + strings.Join(wl.TapeBefore, ""))
		f.WriteString("\n" + strings.Repeat(" ", wl.HeadIndexBefore) + "^")
		f.WriteString("\n" + fmt.Sprintf("%d%s->%d%s%s", wl.StateBefore, wl.HeadBefore, wl.Cmd.nState, wl.Cmd.nSymbol, wl.Cmd.transition))
		f.WriteString("\n" + strings.Join(wl.TapeAfter, ""))
		f.WriteString("\n" + strings.Repeat(" ", wl.HeadIndexAfter) + "^")
		f.WriteString("\n --- \n")

		return f.Close()
	}

	return nil
}

func prepareLogFile(logsPath string, alphabet *Alphabet, tape *Tape, program *Program) error {

	var (
		f   *os.File
		err error
	)

	f, err = os.Create(logsPath)
	if err != nil {
		if os.IsExist(err) {
			return f.Close()
		}

		return errors.New(fmt.Sprintf("prepareLogFile: create file: [%s] err: [%s]", logsPath, err))
	}

	f.WriteString("\n----------\n")
	f.WriteString("----------")

	f.WriteString("\n\n" + fmt.Sprintf("Alphabet: %s\n", alphabet))
	f.WriteString("\n" + fmt.Sprintf("Tape: %s\n\n", tape))

	for state := range *program {
		for head := range (*program)[state] {
			cmd := (*program)[state][head]

			f.WriteString(fmt.Sprintf("%d%s->%d%s%s\n", state, head, cmd.nState, cmd.nSymbol, cmd.transition))
		}
	}

	f.WriteString("\n----------\n")
	f.WriteString("----------\n")

	return f.Close()
}

func LoadTape(tapePath string, alphabet *Alphabet) (*Tape, error) {

	var (
		ok   = false
		tape = &Tape{}
	)

	f, err := os.Open(tapePath)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("LoadTape: open file: [%s] err: [%s]", tapePath, err))
	}

	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		*tape = strings.Split(scanner.Text(), "")
	}

	for _, t := range *tape {
		for _, a := range *alphabet {
			if t == a {
				ok = true
			}
		}

		if !ok {
			return nil, errors.New(fmt.Sprintf("LoadTape: unknown character: [%s]", t))
		}
		ok = false
	}

	return tape, f.Close()
}

func LoadAlphabet(alphabetPath string) (*Alphabet, error) {

	var alphabet = &Alphabet{}

	f, err := os.Open(alphabetPath)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("LoadAlphabet: open file: [%s] err: [%s]", alphabetPath, err))
	}

	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		*alphabet = strings.Split(scanner.Text(), " ")
	}

	return alphabet, f.Close()
}

func LoadProgram(programPath string, alphabet *Alphabet) (*Program, error) {

	var (
		program = make(Program)

		lineIndex = 0
		line      []string

		cStateInt int
		nStateInt int

		cState     State
		nState     State
		head       Head
		nSymbol    string
		transition string

		ok = false
	)

	f, err := os.Open(programPath)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("LoadProgram: open file: [%s] err: [%s]", programPath, err))
	}

	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		lineIndex++

		line = strings.Split(scanner.Text(), "")

		if len(line) == 0 || line[0] == "#" {
			continue
		}

		if len(line) != 7 {
			return nil, errors.New(fmt.Sprintf("LoadProgram: parse program err: incorrect line [%s] line: [%d]", line, lineIndex))
		}

		cStateInt, err = strconv.Atoi(line[0])
		if err != nil {
			return nil, errors.New(fmt.Sprintf("LoadProgram: parse program err: convert current state to int err: [%s] line: [%d]", err, lineIndex))
		}
		cState = State(cStateInt)

		ok = false
		for _, a := range *alphabet {
			if a == line[1] {
				ok = true
			}
		}
		if !ok {
			return nil, errors.New(fmt.Sprintf("LoadProgram: unknown character: [%s] line: [%d]", line[1], lineIndex))
		}
		head = Head(line[1])

		if line[6] != "<" && line[6] != ">" && line[6] != "!" {
			return nil, errors.New(fmt.Sprintf("LoadProgram: parse program err: parse head move err: incorrect symbol: [%s] line: [%d]", line[6], lineIndex))
		}
		transition = line[6]

		nStateInt, err = strconv.Atoi(line[4])
		if err != nil {
			return nil, errors.New(fmt.Sprintf("LoadProgram: convert next state to int err: [%s] line: [%d]", err, lineIndex))
		}
		nState = State(nStateInt)

		ok = false
		for _, a := range *alphabet {
			if a == line[5] {
				ok = true
			}
		}
		if !ok {
			return nil, errors.New(fmt.Sprintf("LoadProgram: unknown character: [%s] line: [%d]", line[5], lineIndex))
		}
		nSymbol = line[5]

		program[cState] = make(map[Head]*Command)
		program[cState][head] = &Command{
			nState:     nState,
			transition: transition,
			nSymbol:    nSymbol,
		}
	}

	if len(program) == 0 {
		return nil, errors.New(fmt.Sprint("LoadProgram: load programm err: empty program"))
	}

	return &program, f.Close()
}
