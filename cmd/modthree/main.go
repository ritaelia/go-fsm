package main

import (
	"fmt"
	"fsm/fsm"
	"os"
	"strings"
)

// Define states (Q).
type State int

const (
	S0 State = iota
	S1
	S2
)

// Define symbols (Σ).
type Bit byte

const (
	Zero Bit = '0'
	One  Bit = '1'
)

// Build the mod-three DFA.
func buildModThree() *fsm.DFA[State, Bit] {
	states := []State{S0, S1, S2}
	alphabet := []Bit{Zero, One}
	q0 := S0
	finals := []State{S0, S1, S2}

	// Transition function δ
	delta := fsm.TransitionFn[State, Bit]{
		S0: fsm.Row(struct {
			On   Bit
			Next State
		}{Zero, S0}, struct {
			On   Bit
			Next State
		}{One, S1}),
		S1: fsm.Row(struct {
			On   Bit
			Next State
		}{Zero, S2}, struct {
			On   Bit
			Next State
		}{One, S0}),
		S2: fsm.Row(struct {
			On   Bit
			Next State
		}{Zero, S1}, struct {
			On   Bit
			Next State
		}{One, S2}),
	}
	return fsm.Must(fsm.NewDFA(states, alphabet, q0, finals, delta, true))
}

// Parse a string like "1011_000" into []Bit.
// Spaces, underscores, and tabs are ignored.
func parseBinary(s string) ([]Bit, error) {
	var out []Bit
	for _, r := range s {
		switch r {
		case '0':
			out = append(out, Zero)
		case '1':
			out = append(out, One)
		case ' ', '\t', '_':
			continue
		default:
			return nil, fmt.Errorf("%w: %q", fsm.ErrInvalidInput, r)
		}
	}
	return out, nil
}

// Map final state to remainder value.
func remainderFromState(s State) int {
	switch s {
	case S0:
		return 0
	case S1:
		return 1
	case S2:
		return 2
	default:
		return -1
	}
}

func main() {
	// Require an input argument.
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <binary-string>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Example: %s 1111_000\n", os.Args[0])
		os.Exit(1)
	}

	input := strings.TrimSpace(os.Args[1])

	// Build the DFA
	d := buildModThree()

	// Parse input string into symbols
	syms, err := parseBinary(input)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Parse error:", err)
		os.Exit(1)
	}

	// Run DFA
	final, err := d.Run(syms)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Run error:", err)
		os.Exit(1)
	}

	// Map final state → remainder
	rem := remainderFromState(final)

	// Print result
	fmt.Printf("Input: %s\nFinal state: %v\nRemainder (mod 3): %d\n", input, final, rem)
}
