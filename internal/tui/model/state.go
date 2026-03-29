package model

// pane is the type of a pane (request or response).
type pane int

const (
	requestPane pane = iota
	responsePane
)

// focusedTarget is the target of the focused pane.
type focusedTarget int

const (
	// focusGlobal is the global focus (all panes).
	focusGlobal focusedTarget = iota
	focusRequestPane
	focusResponsePane
	focusURLBar
	focusModalCollection
	focusModalWorkspace
	focusModalMethod
)

// modalKind is the kind of modal.
type modalKind int

const (
	modalNone modalKind = iota
	modalCollection
	modalWorkspace
	modalMethod
)
