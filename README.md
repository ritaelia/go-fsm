
# go-fsm
A generic finite state machine (FSM) library in Go, with an example implementation of a modulo-3 automaton.

### Finite State Machine (DFA) Library (Go)

This repo contains a small, reusable Deterministic Finite Automaton (DFA) library implemented in Go, plus an example application that computes the remainder of a binary number modulo 3 using a DFA (“mod-three”).

The DFA follows the formal 5-tuple definition:
(Q, Σ, q₀, F, δ) where

Q: set of states
Σ: input alphabet
q₀: initial state
F: accepting/final states
δ: transition function Q × Σ → Q

At a high level, a finite automaton starts in the initial state, processes a sequence of symbols one by one, and ends in a final state that determines whether the input is accepted or what value should be produced.

## Example Walkthrough

Input: `"1011"` (binary for 11)

1. Start at **S0**
2. Read `'1'`: δ(S0,1) → **S1**
3. Read `'0'`: δ(S1,0) → **S2**
4. Read `'1'`: δ(S2,1) → **S2**
5. Read `'1'`: δ(S2,1) → **S2**

Final state = **S2** → remainder = **2**.

## Mapping Theory to Code

* **States (Q):** represented as Go enums (`S0`, `S1`, `S2`)
* **Alphabet (Σ):** represented as symbols (`Zero`, `One`)
* **Initial state (q₀):** configured in `NewDFA` as `S0`
* **Final states (F):** configured as a set of all states `{S0, S1, S2}`
* **Transition function (δ):** represented as a nested map (`TransitionFn`), built using the `fsm.Row(...)` helper
* **Step function:** implements δ(q, a) → q′ (`dfa.Step`)
* **Run function:** iterates over input symbols, applying `Step` until the sequence is exhausted (`dfa.Run`)


### Project structure

fsm/                          # project root (where go.mod lives)
│
├── go.mod                    # module definition
│
├── fsm/                      # library code (reusable)
│   ├── fsm.go                # DFA implementation
│   └── fsm_test.go           # unit + property tests (mod-three)
│
├── cmd/                      # executables 
│   └── modthree/             # specific app
│       └── main.go           # CLI that uses the library (mod-three)
│
└── README.md                 # docs

### Requirements

Go 1.18+ (you’re on 1.18.2; generics are supported)
Windows or any OS supported by Go


### Quick start (Windows / PowerShell)

From the project root:

1.Initialize module
# `go mod init fsm`
# `go mod tidy`

If you plan to push this to GitHub, use:
go mod init github.com/<yourname>/fsm
and then change imports in cmd/modthree/main.go from
import "fsm/fsm" → import "github.com/<yourname>/fsm/fsm"


2.Run the CLI:
# `go run ./cmd/modthree 1011`
or
# `go build ./cmd/modthree`
# `.\modthree 1011`

Expected Output:
Input: 1011
Final state: 2
Remainder (mod 3): 2

3.Run Tests:
# `go test ./fsm -v`
or
# `go test -c ./fsm`
# `.\fsm.test.exe '-test.v'`



### Library overview (API)

# Package: fsm
type DFA[Q comparable, Sigma comparable] struct {
    Q     Set[Q]                        // States
    Sigma Set[Sigma]                    // Alphabet
    Q0    Q                             // Initial state
    F     Set[Q]                        // Final states
    Delta TransitionFn[Q, Sigma]        // δ map[q][symbol] = nextState
}

func NewDFA[Q comparable, Sigma comparable](
    states []Q,
    alphabet []Sigma,
    q0 Q,
    finals []Q,
    delta TransitionFn[Q, Sigma],
    requireComplete bool,
) (*DFA[Q, Sigma], error)

func (d *DFA[Q, Sigma]) Step(q Q, a Sigma) (Q, error)
func (d *DFA[Q, Sigma]) Run(input []Sigma) (Q, error)
func (d *DFA[Q, Sigma]) Accepts(input []Sigma) (bool, Q, error)

// Helpers
type Set[T comparable] map[T]struct{}
type TransitionFn[Q comparable, Sigma comparable] map[Q]map[Sigma]Q
func Row[Q comparable, Sigma comparable](pairs ...struct{ On Sigma; Next Q }) map[Sigma]Q
var ErrInvalidInput = errors.New("invalid input")
func Must[T any](v T, err error) T  // panics on err (handy for demos)


# Example: mod-three DFA
Q = {S0, S1, S2}
Σ = {'0', '1'}
q₀ = S0
F = all states (since we map final state → remainder)
δ:
δ(S0,0)=S0, δ(S0,1)=S1
δ(S1,0)=S2, δ(S1,1)=S0
δ(S2,0)=S1, δ(S2,1)=S2

Check cmd/modthree/main.go for a full, commented example.

# Usage pattern
Choose types for states and symbols (enums work great).
Define states, alphabet, q0, finals.
Build delta with fsm.Row(...).
NewDFA(...) (optionally require a complete δ).
Parse your input → []Sigma.
Run to get the final state (and/or Accepts if using F to recognize a language).


### Tests

Located in fsm/fsm_test.go.
1.Unit vectors (sanity cases)
2.Property test against an arithmetic reference (rem = (rem*2 + bit) % 3)

Run:
# standard
go test ./fsm -v

# workaround if Defender locks temp exe
go test -c ./fsm
.\fsm.test.exe -test.v


### Assumptions & Design Notes

1.Deterministic FA (DFA) only (one next state per (q, symbol)).
2.Generics: states and symbols must be comparable (OK: int, string, custom enums).
3.Validation: NewDFA ensures:
    q0 ∈ Q, F ⊆ Q
    δ only references valid states/symbols
    If requireComplete = true: δ covers all (q, a)
4.Separation of concerns:
    Library handles DFA structure & execution.
    Example app handles parsing strings (ignoring _, spaces, tabs) and mapping final state → remainder.
5.Ergonomics: Row(...) helper makes δ concise; Must(...) is for quick demos (avoid in production APIs).