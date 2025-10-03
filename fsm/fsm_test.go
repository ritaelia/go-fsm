package fsm

import (
	"math/rand"
	"testing"
)

//
// ---------- Reference function ----------
//

// mod3Ref is a ground-truth implementation of mod-3 arithmetic.
// It scans a binary string (with optional separators) and returns remainder % 3.
// Used for property tests to compare DFA output with arithmetic result.
func mod3Ref(binary string) (int, error) {
	rem := 0
	for _, r := range binary {
		switch r {
		case '0', '1':
			rem = (rem*2 + int(r-'0')) % 3
		case ' ', '_', '\t':
			// separators ignored (like app-level parsing)
			continue
		default:
			// reject invalid characters
			return 0, ErrInvalidInput
		}
	}
	return rem, nil
}

//
// ---------- DFA types for this test ----------
//

// State represents DFA states (remainder classes).
type State int

const (
	S0 State = iota // remainder 0
	S1              // remainder 1
	S2              // remainder 2
)

// Bit is the input symbol type.
type Bit byte

const (
	Zero Bit = '0'
	One  Bit = '1'
)

//
// ---------- Helper: Build the DFA ----------
//

// buildModThree constructs the canonical modulo-3 DFA using the generic library.
func buildModThree() *DFA[State, Bit] {
	states := []State{S0, S1, S2}
	alphabet := []Bit{Zero, One}
	q0 := S0
	finals := []State{S0, S1, S2} // all states final

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

//
// ---------- Unit tests ----------
//

// TestModThreeVectors checks known binary → remainder mappings.
func TestModThreeVectors(t *testing.T) {
	cases := map[string]int{
		"":         0, // empty string → remainder 0
		"0":        0,
		"1":        1,
		"10":       2,
		"11":       0,
		"1011":     2,
		"1111":     0,
		"1111_000": 0, // separators ignored by parsing
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
		final, err := d.Run(syms)
		if err != nil {
			t.Fatal(err)
		}
		got := map[State]int{S0: 0, S1: 1, S2: 2}[final]
		if got != want {
			t.Fatalf("%q -> %d, want %d", in, got, want)
		}
	}
}

// TestEmptyInput ensures that DFA.Run(nil) returns q0.
func TestEmptyInput(t *testing.T) {
	d := buildModThree()
	final, err := d.Run(nil)
	if err != nil {
		t.Fatal(err)
	}
	if final != S0 {
		t.Fatalf("empty input should yield S0, got %v", final)
	}
}

// TestAccepts_AllStatesFinal ensures Accepts always true since all states are final.
func TestAccepts_AllStatesFinal(t *testing.T) {
	d := buildModThree()
	ok, final, err := d.Accepts([]Bit{Zero, One, One})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatalf("expected Accepts=true, got false (final=%v)", final)
	}
}

//
// ---------- Constructor validation tests ----------
//

// Invalid q0 (not in Q) should fail.
func TestNewDFA_InvalidQ0(t *testing.T) {
	states := []State{S0}
	alphabet := []Bit{Zero}
	finals := []State{S0}
	delta := TransitionFn[State, Bit]{S0: Row(struct {
		On   Bit
		Next State
	}{Zero, S0})}
	if _, err := NewDFA(states, alphabet, S1, finals, delta, true); err == nil {
		t.Fatal("expected error for q0 not in Q")
	}
}

// Final state not in Q should fail.
func TestNewDFA_FinalNotInQ(t *testing.T) {
	states := []State{S0}
	alphabet := []Bit{Zero}
	finals := []State{S1} // invalid
	delta := TransitionFn[State, Bit]{S0: Row(struct {
		On   Bit
		Next State
	}{Zero, S0})}
	if _, err := NewDFA(states, alphabet, S0, finals, delta, true); err == nil {
		t.Fatal("expected error for F ⊄ Q")
	}
}

// Transition uses symbol not in Σ should fail.
func TestNewDFA_SymbolNotInAlphabet(t *testing.T) {
	states := []State{S0}
	alphabet := []Bit{Zero}
	finals := []State{S0}
	// δ references symbol One not in Σ
	delta := TransitionFn[State, Bit]{S0: Row(struct {
		On   Bit
		Next State
	}{Zero, S0}, struct {
		On   Bit
		Next State
	}{One, S0})}
	if _, err := NewDFA(states, alphabet, S0, finals, delta, true); err == nil {
		t.Fatal("expected error for δ symbol not in Σ")
	}
}

// Transition points to a state not in Q should fail.
func TestNewDFA_TargetStateNotInQ(t *testing.T) {
	states := []State{S0}
	alphabet := []Bit{Zero}
	finals := []State{S0}
	// δ maps to S1 which is not in Q
	delta := TransitionFn[State, Bit]{S0: Row(struct {
		On   Bit
		Next State
	}{Zero, S1})}
	if _, err := NewDFA(states, alphabet, S0, finals, delta, true); err == nil {
		t.Fatal("expected error for δ target not in Q")
	}
}

// requireComplete=true should reject missing edges.
func TestNewDFA_RequireComplete(t *testing.T) {
	states := []State{S0}
	alphabet := []Bit{Zero, One}
	finals := []State{S0}
	// Missing transition for One
	delta := TransitionFn[State, Bit]{S0: Row(struct {
		On   Bit
		Next State
	}{Zero, S0})}

	// Should error when requireComplete=true
	if _, err := NewDFA(states, alphabet, S0, finals, delta, true); err == nil {
		t.Fatal("expected error with requireComplete=true")
	}

	// With requireComplete=false, build works, but Step should fail on missing input
	d, err := NewDFA(states, alphabet, S0, finals, delta, false)
	if err != nil {
		t.Fatalf("unexpected error with requireComplete=false: %v", err)
	}
	if _, err := d.Step(S0, One); err == nil {
		t.Fatal("expected Step error for missing (S0,One)")
	}
}

//
// ---------- Property tests ----------
//

// TestPropertyRandom generates random binary strings and checks DFA output against arithmetic ref.
func TestPropertyRandom(t *testing.T) {
	r := rand.New(rand.NewSource(42))
	d := buildModThree()
	for n := 1; n < 500; n++ {
		var s string
		for i := 0; i < n; i++ {
			if r.Intn(2) == 0 {
				s += "0"
			} else {
				s += "1"
			}
		}
		ref, _ := mod3Ref(s)

		// convert to []Bit
		var syms []Bit
		for _, r := range s {
			if r == '0' {
				syms = append(syms, Zero)
			} else {
				syms = append(syms, One)
			}
		}

		final, err := d.Run(syms)
		if err != nil {
			t.Fatal(err)
		}
		got := map[State]int{S0: 0, S1: 1, S2: 2}[final]
		if got != ref {
			t.Fatalf("%q -> %d, want %d", s, got, ref)
		}
	}
}
