package main

import (
	"flag"
	"os"
	"log"
	"bufio"
	"strings"
	"strconv"
	"errors"
	"fmt"
)

const (
	LEFT  = iota
	RIGHT
	STOP
)

type Alphabet []string

type Tape []string

type NState struct {
	nHead  Head
	nState State
	Move   int
}

type Head string

type State int

type Instructions map[State]map[Head]*NState

func main() {

	var (
		alphabetPath     = flag.String("alph", "", "path to alphabet file")
		tapePath         = flag.String("tape", "", "path to tape file")
		instructionsPath = flag.String("inst", "", "path to instructions file")
	)

	flag.Parse()

	if *alphabetPath == "" {
		log.Fatalf("Incorrect alphabet file path: [%s]", *alphabetPath)
	} else if *tapePath == "" {
		log.Fatalf("Incorrect tape filepath: [%s]", *tapePath)
	} else if *instructionsPath == "" {
		log.Fatalf("Incorrect instructions file path: [%s]", *instructionsPath)
	}
}

func LoadTape(tapePath string) (*Tape, error) {

	var tape = &Tape{}

	f, err := os.Open(tapePath)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("LoadTape: open file: [%s] err: [%s]", tapePath, err))
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		*tape = strings.Split(scanner.Text(), "")
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
	for scanner.Scan() {
		*alphabet = strings.Split(scanner.Text(), " ")
	}

	return alphabet, nil
}

func LoadInstructions(instructionsPath string) (*Instructions, error) {

	var (
		instructions = make(Instructions)

		line        = 0
		instruction []string
		cStateInt   int
		nStateInt   int

		cState State
		nState State
		cHead  Head
		nHead  Head
		move   int
	)

	f, err := os.Open(instructionsPath)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("LoadInstructions: open file: [%s] err: [%s]", instructionsPath, err))
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line++

		instruction = strings.Split(scanner.Text(), "")

		if len(instruction) == 0 || instruction[0] == "#" {
			continue
		}

		if len(instruction) != 7 {
			return nil, errors.New(fmt.Sprintf("LoadInstructions: parse instruction err: incorrect line [%s] line number: [%d]", instruction, line))
		}

		cStateInt, err = strconv.Atoi(instruction[0])
		if err != nil {
			return nil, errors.New(fmt.Sprintf("LoadInstructions: parse instruction err: convert current state to int err: [%s]", err))
		}
		cState = State(cStateInt)
		cHead = Head(instruction[1])

		switch instruction[6] {
		case ">":
			move = RIGHT
			break
		case "<":
			move = LEFT
			break
		case "!":
			move = STOP
			break
		default:
			return nil, errors.New(fmt.Sprintf("LoadInstructions: parse instruction err: parse head move err: incorrect symbol: [%s]", instruction[7]))
		}

		nStateInt, err = strconv.Atoi(instruction[4])
		if err != nil {
			return nil, errors.New(fmt.Sprintf("LoadInstructions: convert nezt state to int err: [%s]", err))
		}
		nState = State(nStateInt)
		nHead = Head(instruction[5])

		instructions[cState] = make(map[Head]*NState)
		instructions[cState][cHead] = &NState{
			nState: nState,
			Move:   move,
			nHead:  nHead,
		}
	}

	if len(instruction) == 0 {
		return nil, errors.New(fmt.Sprint("LoadInstructions: empty instructions"))
	}

	return &instructions, nil
}
