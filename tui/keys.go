package tui

import (
	"github.com/charmbracelet/bubbles/key"
)

// contextualKeyMap represents key bindings for a specific screen context
type contextualKeyMap struct {
	bindings []key.Binding
}

func (k contextualKeyMap) ShortHelp() []key.Binding {
	return k.bindings
}

func (k contextualKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{k.bindings}
}

type keyMap struct {
	Up         key.Binding
	Down       key.Binding
	Left       key.Binding
	Right      key.Binding
	Enter      key.Binding
	Back       key.Binding
	Quit       key.Binding
	New        key.Binding
	NewFeature key.Binding
	NewStory   key.Binding
	Space      key.Binding
	Complete   key.Binding
	Search     key.Binding
	Tab        key.Binding
	Delete     key.Binding
	ScrollUp   key.Binding
	ScrollDown key.Binding
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("k/↑", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("j/↓", "down"),
	),
	Left: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("h/←", "left"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("l/→", "right"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "open"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	New: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "new"),
	),
	NewFeature: key.NewBinding(
		key.WithKeys("f"),
		key.WithHelp("f", "new feature"),
	),
	NewStory: key.NewBinding(
		key.WithKeys("u"),
		key.WithHelp("u", "new story"),
	),
	Space: key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp("space", "toggle"),
	),
	Complete: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "complete all"),
	),
	Search: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "search"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next field"),
	),
	Delete: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "delete"),
	),
	ScrollUp: key.NewBinding(
		key.WithKeys("ctrl+u"),
		key.WithHelp("ctrl+u", "scroll up"),
	),
	ScrollDown: key.NewBinding(
		key.WithKeys("ctrl+d"),
		key.WithHelp("ctrl+d", "scroll down"),
	),
}

// Key maps for different screens
func featureListKeys() contextualKeyMap {
	return contextualKeyMap{
		bindings: []key.Binding{keys.Up, keys.Down, keys.Enter, keys.NewFeature, keys.Search, keys.Quit},
	}
}

func featureDetailKeys() contextualKeyMap {
	return contextualKeyMap{
		bindings: []key.Binding{keys.Up, keys.Down, keys.ScrollUp, keys.ScrollDown, keys.Enter, keys.NewStory, keys.Delete, keys.Back},
	}
}

func storyDetailKeys() contextualKeyMap {
	return contextualKeyMap{
		bindings: []key.Binding{keys.ScrollUp, keys.ScrollDown, keys.Complete, keys.Back},
	}
}

func featureFormKeys() contextualKeyMap {
	return contextualKeyMap{
		bindings: []key.Binding{keys.Tab, keys.Space, keys.Enter, keys.Back},
	}
}

func storyFormKeys() contextualKeyMap {
	return contextualKeyMap{
		bindings: []key.Binding{keys.Tab, keys.Left, keys.Right, keys.Enter, keys.Back},
	}
}

func searchKeys() contextualKeyMap {
	return contextualKeyMap{
		bindings: []key.Binding{keys.Up, keys.Down, keys.Enter, keys.Back},
	}
}

func wideKeys() contextualKeyMap {
	return contextualKeyMap{
		bindings: []key.Binding{keys.Up, keys.Down, keys.ScrollUp, keys.ScrollDown, keys.Enter, keys.Space, keys.NewFeature, keys.NewStory, keys.Delete, keys.Search, keys.Complete, keys.Quit},
	}
}
