package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/huh"
)

type GroupChoice struct {
	Name   string
	New    bool
	Ignore bool
	Stop   bool
	Open   bool
}

func SelectGroup(groups []string) (GroupChoice, error) {
	options := groupOptions(groups)
	options = append(options, huh.NewOption("+ Create new group", "__new__"))
	options = append(options, huh.NewOption("Do not add (blacklist)", "__ignore__"))
	options = append(options, huh.NewOption("Stop scanning", "__stop__"))

	var selected string
	form := newForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select a group").
				Options(options...).
				Value(&selected),
		),
	)

	if err := form.Run(); err != nil {
		return GroupChoice{}, err
	}

	if selected == "__new__" {
		var name string
		input := newForm(
			huh.NewGroup(
				huh.NewInput().
					Title("New group name").
					Value(&name),
			),
		)

		if err := input.Run(); err != nil {
			return GroupChoice{}, err
		}

		name = strings.TrimSpace(name)
		if name == "" {
			return GroupChoice{}, fmt.Errorf("group name is empty")
		}

		return GroupChoice{Name: name, New: true}, nil
	}

	if selected == "__ignore__" {
		return GroupChoice{Ignore: true}, nil
	}

	if selected == "__stop__" {
		return GroupChoice{Stop: true}, nil
	}

	return GroupChoice{Name: selected, New: false}, nil
}

func SelectGroupForPath(groups []string, path string, projectRoot string) (GroupChoice, error) {
	title := "Select a group"
	if path != "" {
		title = fmt.Sprintf("Select a group for: %s", path)
		if projectRoot != "" {
			title = fmt.Sprintf("Select a group for: %s (%s)", path, projectRoot)
		}
	}

	options := groupOptions(groups)
	options = append(options, huh.NewOption("+ Create new group", "__new__"))
	options = append(options, huh.NewOption("Open (read-only)", "__open__"))
	options = append(options, huh.NewOption("Do not add (blacklist)", "__ignore__"))
	options = append(options, huh.NewOption("Stop scanning", "__stop__"))

	var selected string
	form := newForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(title).
				Options(options...).
				Value(&selected),
		),
	)

	if err := form.Run(); err != nil {
		return GroupChoice{}, err
	}

	if selected == "__open__" {
		return GroupChoice{Open: true}, nil
	}

	if selected == "__new__" {
		var name string
		input := newForm(
			huh.NewGroup(
				huh.NewInput().
					Title("New group name").
					Value(&name),
			),
		)

		if err := input.Run(); err != nil {
			return GroupChoice{}, err
		}

		name = strings.TrimSpace(name)
		if name == "" {
			return GroupChoice{}, fmt.Errorf("group name is empty")
		}

		return GroupChoice{Name: name, New: true}, nil
	}

	if selected == "__ignore__" {
		return GroupChoice{Ignore: true}, nil
	}

	if selected == "__stop__" {
		return GroupChoice{Stop: true}, nil
	}

	return GroupChoice{Name: selected, New: false}, nil
}

type GroupSelection struct {
	GroupName string
	Paths     []string
}

func SelectGroupToEdit(groups []string) (string, bool, error) {
	options := groupOptions(groups)
	options = append(options, huh.NewOption("Done", "__done__"))

	var selected string
	form := newForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select a group to pick workflows").
				Options(options...).
				Value(&selected),
		),
	)

	if err := form.Run(); err != nil {
		return "", false, err
	}

	if selected == "__done__" {
		return "", true, nil
	}

	return selected, false, nil
}

func SelectGroupOrNew(groups []string) (GroupChoice, error) {
	options := groupOptions(groups)
	options = append(options, huh.NewOption("+ Create new group", "__new__"))

	var selected string
	form := newForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select destination group").
				Options(options...).
				Value(&selected),
		),
	)

	if err := form.Run(); err != nil {
		return GroupChoice{}, err
	}

	if selected == "__new__" {
		var name string
		input := newForm(
			huh.NewGroup(
				huh.NewInput().
					Title("New group name").
					Value(&name),
			),
		)

		if err := input.Run(); err != nil {
			return GroupChoice{}, err
		}

		name = strings.TrimSpace(name)
		if name == "" {
			return GroupChoice{}, fmt.Errorf("group name is empty")
		}

		return GroupChoice{Name: name, New: true}, nil
	}

	return GroupChoice{Name: selected, New: false}, nil
}

type WorkflowOption struct {
	Label string
	Value string
}

func SelectWorkflows(group string, options []WorkflowOption, preselected []string) ([]string, error) {
	opts := []huh.Option[string]{}
	for _, o := range options {
		opts = append(opts, huh.NewOption(o.Label, o.Value))
	}

	selected := append([]string{}, preselected...)

	form := newForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title(fmt.Sprintf("Select workflows in group: %s", group)).
				Options(opts...).
				Value(&selected),
		),
	)

	if err := form.Run(); err != nil {
		return nil, err
	}

	return selected, nil
}

func SelectIgnoredToRestore(paths []string) ([]string, error) {
	opts := []huh.Option[string]{}
	for _, p := range paths {
		opts = append(opts, huh.NewOption(p, p))
	}

	var selected []string
	form := newForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Select items to restore").
				Options(opts...).
				Value(&selected),
		),
	)

	if err := form.Run(); err != nil {
		return nil, err
	}

	return selected, nil
}

func groupOptions(groups []string) []huh.Option[string] {
	options := make([]huh.Option[string], 0, len(groups))
	for _, g := range groups {
		options = append(options, huh.NewOption("· "+g, g))
	}
	return options
}

func newForm(groups ...*huh.Group) *huh.Form {
	form := huh.NewForm(groups...)
	return form.WithKeyMap(vimKeyMap())
}

func vimKeyMap() *huh.KeyMap {
	km := huh.NewDefaultKeyMap()
	km.Select.Up = key.NewBinding(
		key.WithKeys("k", "up", "ctrl+k", "ctrl+p"),
		key.WithHelp("k/up", "up"),
	)
	km.Select.Down = key.NewBinding(
		key.WithKeys("j", "down", "ctrl+j", "ctrl+n"),
		key.WithHelp("j/down", "down"),
	)
	km.MultiSelect.Up = key.NewBinding(
		key.WithKeys("k", "up", "ctrl+p"),
		key.WithHelp("k/up", "up"),
	)
	km.MultiSelect.Down = key.NewBinding(
		key.WithKeys("j", "down", "ctrl+n"),
		key.WithHelp("j/down", "down"),
	)
	return km
}
