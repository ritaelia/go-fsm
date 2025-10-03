package fsm

import (
	"math/rand"
	"testing"
)

// Reference implementation of mod 3 arithmetic (used for property testing).
// Reads a binary string and computes remainder = value % 3.
func mod3Ref(binary string) (int, error) {
	rem := 0
	for _, r := range binary {
		switch r {
		case '0', '1':
			rem = (rem*2 + int(r-'0')) % 3
		case ' ', '_', '\t':
			// separators are ignored
			continue
		default:
			return 0, ErrInvalidInput
		}
	}
	return rem, nil
}

// Define state and symbol types for mod-three DFA.
type State int

const (
	S0 State = iota
	S1
	S2
)

type Bit byte

const (
	Zero Bit = '0'
	One  Bit = '1'
)

// Build the mod-three DFA using the generic library.
func buildModThree() *DFA[State, Bit] {
	states := []State{S0, S1, S2}
	alphabet := []Bit{Zero, One}
	q0 := S0
	finals := []State{S0, S1, S2} // all states are final in this DFA

	delta := TransitionFn[State, Bit]{
		S0: Row(struct {
			On   Bit
			Next State
		}{Zero, S0}, struct {
			On   Bit
			Next State
		}{One, S1}),
		S1: Row(struct {
			On   Bit
			Next State
		}{Zero, S2}, struct {
			On   Bit
			Next State
		}{One, S0}),
		S2: Row(struct {
			On   Bit
			Next State
		}{Zero, S1}, struct {
			On   Bit
			Next State
		}{One, S2}),
	}
	return Must(NewDFA(states, alphabet, q0, finals, delta, true))
}

// Unit tests with known test vectors.
func TestModThreeVectors(t *testing.T) {
	cases := map[string]int{
		"": 0, "0": 0, "1": 1, "10": 2,
		"11": 0, "1011": 2, "1111": 0, "1111_000": 0,
	}
	d := buildModThree()
	for in, want := range cases {
		var syms []Bit
		for _, r := range in {
			if r == '0' {
				syms = append(syms, Zero)
			}
			if r == '1' {
				syms = append(syms, One)
			}
		}
		q, err := d.Run(syms)
		if err != nil {
			t.Fatal(err)
		}
		got := map[State]int{S0: 0, S1: 1, S2: 2}[q]
		if got != want {
			t.Fatalf("%q -> %d, want %d", in, got, want)
		}
	}
}

// Property test: DFA result must equal arithmetic reference for many random strings.
func TestPropertyRandom(t *testing.T) {
	r := rand.New(rand.NewSource(42))
	d := buildModThree()
	for n := 1; n < 500; n++ { // test strings of length 1..499
		var s string
		for i := 0; i < n; i++ {
			if r.Intn(2) == 0 {
				s += "0"
			} else {
				s += "1"
			}
		}
		ref, _ := mod3Ref(s)
		var syms []Bit
		for _, r := range s {
			if r == '0' {
				syms = append(syms, Zero)
			} else {
				syms = append(syms, One)
			}
		}
		q, err := d.Run(syms)
		if err != nil {
			t.Fatal(err)
		}
		got := map[State]int{S0: 0, S1: 1, S2: 2}[q]
		if got != ref {
			t.Fatalf("%q -> %d, want %d", s, got, ref)
		}
	}
}
