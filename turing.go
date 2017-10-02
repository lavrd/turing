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

const (
	LEFT  = iota
	RIGHT
	STOP
)

type Alphabet []string
type Tape []string
type Program map[State]map[Head]*Command

type Command struct {
	nHead      Head
	nState     State
	transition int
}
type Head string
type State int

type wLog struct {
	TapeBefore  Tape
	TapeAfter   Tape
	HeadBefore  Head
	HeadAfter   Head
	StateBefore State
	StateAfter  State
	Cmd         Command
}

func main() {

	var (
		alphabetPath = flag.String("alph", "", "path to alphabet file")
		tapePath     = flag.String("tape", "", "path to tape file")
		programPath  = flag.String("prog", "", "path to program file")

		//savePath = flag.String("save", "", "if the flag is specified logs will be saved to specified file")
		//verbose  = flag.Bool("v", false, "verbose output")
		//display = flag.Bool("d", false, "display result")
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

	_, err = LoadTape(*tapePath, alphabet)
	if err != nil {
		log.Fatal(err)
	}

	_, err = LoadProgram(*programPath, alphabet)
	if err != nil {
		log.Fatal(err)
	}
}

func (wl *wLog) Log(savePath string, verbose bool) (bool, error) {
	return true, nil
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
	defer f.Close()

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

	return tape, nil
}

func LoadAlphabet(alphabetPath string) (*Alphabet, error) {

	var alphabet = &Alphabet{}

	f, err := os.Open(alphabetPath)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("LoadAlphabet: open file: [%s] err: [%s]", alphabetPath, err))
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		*alphabet = strings.Split(scanner.Text(), " ")
	}

	return alphabet, nil
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
		cHead      Head
		nHead      Head
		transition int

		ok = false
	)

	f, err := os.Open(programPath)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("LoadProgram: open file: [%s] err: [%s]", programPath, err))
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		lineIndex++

		line = strings.Split(scanner.Text(), "")

		if len(line) == 0 || line[0] == "#" {
			continue
		}

		if len(line) != 7 {
			return nil, errors.New(fmt.Sprintf("LoadProgram: parse program err: incorrect line [%s] line number: [%d]", line, lineIndex))
		}

		cStateInt, err = strconv.Atoi(line[0])
		if err != nil {
			return nil, errors.New(fmt.Sprintf("LoadProgram: parse program err: convert current state to int err: [%s]", err))
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
		cHead = Head(line[1])

		switch line[6] {
		case ">":
			transition = RIGHT
			break
		case "<":
			transition = LEFT
			break
		case "!":
			transition = STOP
			break
		default:
			return nil, errors.New(fmt.Sprintf("LoadProgram: parse program err: parse head move err: incorrect symbol: [%s]", line[7]))
		}

		nStateInt, err = strconv.Atoi(line[4])
		if err != nil {
			return nil, errors.New(fmt.Sprintf("LoadProgram: convert next state to int err: [%s]", err))
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
		nHead = Head(line[5])

		program[cState] = make(map[Head]*Command)
		program[cState][cHead] = &Command{
			nState:     nState,
			transition: transition,
			nHead:      nHead,
		}
	}

	if len(program) == 0 {
		return nil, errors.New(fmt.Sprint("LoadProgram: load programm err: empty program"))
	}

	return &program, nil
}
