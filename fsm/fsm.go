// Package fsm provides a generic Deterministic Finite Automaton (DFA) library.
// A DFA is defined by the 5-tuple (Q, Σ, q0, F, δ):
//   Q  = set of states
//   Σ  = input alphabet
//   q0 = initial state
//   F  = final/accepting states
//   δ  = transition function Q × Σ → Q
//
// This package lets you define states and symbols as any comparable type
// (enums, ints, strings, etc.), build a DFA with validation, and run it
// on an input sequence.

package fsm

import (
	"errors"
	"fmt"
)

// ---------- Helpers ----------

// Set is a simple generic set implementation based on a map.
// Used for representing Q (states), Σ (alphabet), and F (final states).
type Set[T comparable] map[T]struct{}

// NewSet creates a new Set[T] from a slice of values.
func NewSet[T comparable](xs ...T) Set[T] {
	s := make(Set[T], len(xs))
	for _, x := range xs {
		s[x] = struct{}{}
	}
	return s
}

// Has checks membership in the set.
func (s Set[T]) Has(x T) bool { _, ok := s[x]; return ok }

// TransitionFn encodes the transition function δ as nested maps.
// Example: delta[q][symbol] = nextState
type TransitionFn[Q comparable, Sigma comparable] map[Q]map[Sigma]Q

// ---------- DFA definition ----------

// DFA is a generic Deterministic Finite Automaton.
// It stores:
//   Q      = set of states
//   Sigma  = alphabet
//   Q0     = initial state
//   F      = set of accepting/final states
//   Delta  = transition function
type DFA[Q comparable, Sigma comparable] struct {
	Q     Set[Q]
	Sigma Set[Sigma]
	Q0    Q
	F     Set[Q]
	Delta TransitionFn[Q, Sigma]
}

// ---------- Constructor ----------

// NewDFA builds a new DFA and validates it.
// - It checks that q0 ∈ Q.
// - It checks that F ⊆ Q.
// - It checks that every δ(q,σ) target is in Q.
// - If requireComplete=true, it ensures every (q,σ) pair has a transition.
func NewDFA[Q comparable, Sigma comparable](
	states []Q,
	alphabet []Sigma,
	q0 Q,
	finals []Q,
	delta TransitionFn[Q, Sigma],
	requireComplete bool,
) (*DFA[Q, Sigma], error) {
	Qset := NewSet(states...)
	Sset := NewSet(alphabet...)
	Fset := NewSet(finals...)

	// Check initial state
	if !Qset.Has(q0) {
		return nil, fmt.Errorf("q0 %v not in Q", q0)
	}
	// Check finals
	for f := range Fset {
		if !Qset.Has(f) {
			return nil, fmt.Errorf("final %v not in Q", f)
		}
	}
	// Validate delta transitions
	for q, row := range delta {
		if !Qset.Has(q) {
			return nil, fmt.Errorf("delta references unknown state %v", q)
		}
		for a, qNext := range row {
			if !Sset.Has(a) {
				return nil, fmt.Errorf("delta row %v has symbol %v not in Σ", q, a)
			}
			if !Qset.Has(qNext) {
				return nil, fmt.Errorf("delta(%v,%v) → %v not in Q", q, a, qNext)
			}
		}
	}
	// If completeness required, check every (q,a)
	if requireComplete {
		for q := range Qset {
			row, ok := delta[q]
			if !ok {
				return nil, fmt.Errorf("delta missing row for state %v", q)
			}
			for a := range Sset {
				if _, ok := row[a]; !ok {
					return nil, fmt.Errorf("delta missing (%v,%v)", q, a)
				}
			}
		}
	}

	return &DFA[Q, Sigma]{
		Q:     Qset,
		Sigma: Sset,
		Q0:    q0,
		F:     Fset,
		Delta: delta,
	}, nil
}

// ---------- Core ops ----------

// Step applies a single transition: q' = δ(q,a).
// Returns an error if the transition is undefined.
func (d *DFA[Q, Sigma]) Step(q Q, a Sigma) (Q, error) {
	row, ok := d.Delta[q]
	if !ok {
		return q, fmt.Errorf("no row for state %v", q)
	}
	qNext, ok := row[a]
	if !ok {
		return q, fmt.Errorf("no transition for (%v,%v)", q, a)
	}
	return qNext, nil
}

// Run consumes an input sequence (slice of symbols) and returns the final state.
func (d *DFA[Q, Sigma]) Run(input []Sigma) (Q, error) {
	q := d.Q0
	var err error
	for _, a := range input {
		q, err = d.Step(q, a)
		if err != nil {
			return q, err
		}
	}
	return q, nil
}

// Accepts runs the DFA and checks if the final state is in F (accepting).
func (d *DFA[Q, Sigma]) Accepts(input []Sigma) (bool, Q, error) {
	q, err := d.Run(input)
	if err != nil {
		return false, q, err
	}
	_, ok := d.F[q]
	return ok, q, nil
}

// ---------- Convenience ----------

// Must panics if err != nil. Useful for quick demos.
func Must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

// Row helper to build δ rows in a clean way.
// Example:
//   delta := fsm.TransitionFn[State,Bit]{
//     S0: fsm.Row(struct{On Bit; Next State}{Zero,S0}, ...),
//   }
func Row[Q comparable, Sigma comparable](pairs ...struct {
	On   Sigma
	Next Q
}) map[Sigma]Q {
	m := make(map[Sigma]Q, len(pairs))
	for _, p := range pairs {
		m[p.On] = p.Next
	}
	return m
}

// Sentinel error for invalid characters or inputs.
var ErrInvalidInput = errors.New("invalid input")
