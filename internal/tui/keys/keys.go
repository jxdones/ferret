package keys

import "charm.land/bubbles/v2/key"

// Map holds ferret's key bindings.
type Map struct {
	SendRequest   key.Binding
	NewRequest    key.Binding
	CancelRequest key.Binding

	Collection      key.Binding
	CollectionCycle key.Binding

	WorkspacePick key.Binding

	URLFocus key.Binding

	MethodCycle  key.Binding
	MethodPicker key.Binding
	EnvCycle     key.Binding

	NextTab  key.Binding
	PrevTab  key.Binding
	NewTab   key.Binding
	CloseTab key.Binding

	Help key.Binding
}

// Default is the shared ferret keymap.
var Default = Map{
	SendRequest:     key.NewBinding(key.WithKeys("ctrl+r"), key.WithHelp("^r", "send")),
	NewRequest:      key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "new request")),
	Collection:      key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "requests")),
	CollectionCycle: key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "cycle collection")),
	WorkspacePick:   key.NewBinding(key.WithKeys("C", "shift+c"), key.WithHelp("C", "pick a collection")),
	Help:            key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),

	URLFocus: key.NewBinding(key.WithKeys("ctrl+u"), key.WithHelp("^u", "edit url")),

	MethodCycle:   key.NewBinding(key.WithKeys("m"), key.WithHelp("m", "cycle method")),
	MethodPicker:  key.NewBinding(key.WithKeys("M", "shift+m"), key.WithHelp("M", "pick a method")),
	EnvCycle:      key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "cycle env")),
	NextTab:       key.NewBinding(key.WithKeys("ctrl+n"), key.WithHelp("^n", "next tab")),
	PrevTab:       key.NewBinding(key.WithKeys("ctrl+p"), key.WithHelp("^p", "prev tab")),
	NewTab:        key.NewBinding(key.WithKeys("T", "shift+t"), key.WithHelp("T", "new tab")),
	CloseTab:      key.NewBinding(key.WithKeys("X", "shift+x"), key.WithHelp("X", "close tab")),
	CancelRequest: key.NewBinding(key.WithKeys("ctrl+x"), key.WithHelp("^x", "cancel request")),
}

// HelpBindings returns the bindings shown in the compact footer shortcuts bar.
func HelpBindings() []key.Binding {
	return []key.Binding{
		Default.SendRequest,
		Default.NewRequest,
		Default.Collection,
		Default.CollectionCycle,
		Default.WorkspacePick,
		Default.MethodCycle,
		Default.EnvCycle,
		Default.Help,
	}
}

// FullHelpGroups returns all bindings grouped by category for the ? help modal.
func FullHelpGroups() [][]key.Binding {
	return [][]key.Binding{
		{Default.SendRequest, Default.NewRequest, Default.URLFocus, Default.MethodCycle, Default.MethodPicker, Default.EnvCycle},
		{Default.Collection, Default.CollectionCycle, Default.WorkspacePick},
		{
			key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next main pane")),
			key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "prev main pane")),
			Default.NextTab,
			Default.PrevTab,
			Default.NewTab,
			Default.CloseTab,
		},
		{Default.Help},
	}
}
