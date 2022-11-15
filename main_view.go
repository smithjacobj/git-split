package main

import (
	"fmt"
	"strings"

	"github.com/awesome-gocui/gocui"
	"github.com/bluekeyes/go-gitdiff/gitdiff"
	ir "github.com/smithjacobj/git-split/difftree"
)

const k_MainView = "main"

var ErrConfirm = fmt.Errorf("confirm changes and quit")

type MainView struct {
	*gocui.Gui
	*gocui.View

	commit *ir.Commit
}

func LayoutMainView(g *gocui.Gui) (v *MainView, isInit bool, err error) {
	v = &MainView{Gui: g}
	maxX, maxY := g.Size()
	v.View, err = g.SetView(k_MainView, 0, k_HelpViewHeight-1, maxX-1, maxY-1, 0)
	if err != nil {
		if err == gocui.ErrUnknownView {
			isInit = true
		} else {
			return nil, false, err
		}
	}

	if isInit {
		if err := v.setKeybindings(); err != nil {
			return nil, false, err
		}

		v.View.Frame = false
		v.View.Highlight = true
	}

	return v, isInit, nil
}

func (v *MainView) SetCommit(c *ir.Commit) {
	v.commit = c
	v.printContent()
	v.View.SetCursor(0, 0)
}

func (v *MainView) setKeybindings() error {
	if err := v.Gui.SetKeybinding(v.View.Name(), gocui.KeyArrowUp, gocui.ModNone, moveCursor(v, -1)); err != nil {
		return err
	}
	if err := v.Gui.SetKeybinding(v.View.Name(), gocui.KeyArrowUp, gocui.ModShift, moveCursor(v, -15)); err != nil {
		return err
	}
	if err := v.Gui.SetKeybinding(v.View.Name(), gocui.KeyPgup, gocui.ModNone, moveCursor(v, -15)); err != nil {
		return err
	}
	if err := v.Gui.SetKeybinding(v.View.Name(), gocui.KeyArrowDown, gocui.ModNone, moveCursor(v, 1)); err != nil {
		return err
	}
	if err := v.Gui.SetKeybinding(v.View.Name(), gocui.KeyArrowDown, gocui.ModShift, moveCursor(v, 15)); err != nil {
		return err
	}
	if err := v.Gui.SetKeybinding(v.View.Name(), gocui.KeyPgdn, gocui.ModNone, moveCursor(v, 15)); err != nil {
		return err
	}
	if err := v.Gui.SetKeybinding(v.View.Name(), gocui.KeyArrowLeft, gocui.ModNone, setExpansionState(v, ir.Collapsed)); err != nil {
		return err
	}
	if err := v.Gui.SetKeybinding(v.View.Name(), gocui.KeyArrowLeft, gocui.ModShift, setExpansionAll(v, ir.Collapsed)); err != nil {
		return err
	}
	if err := v.Gui.SetKeybinding(v.View.Name(), gocui.KeyArrowRight, gocui.ModNone, setExpansionState(v, ir.Expanded)); err != nil {
		return err
	}
	if err := v.Gui.SetKeybinding(v.View.Name(), gocui.KeyArrowRight, gocui.ModShift, setExpansionAll(v, ir.Expanded)); err != nil {
		return err
	}
	if err := v.Gui.SetKeybinding(v.View.Name(), gocui.KeySpace, gocui.ModNone, toggleSelection(v)); err != nil {
		return err
	}
	if err := v.Gui.SetKeybinding(v.View.Name(), 'c', gocui.ModNone, confirm(v)); err != nil {
		return err
	}
	if err := v.Gui.SetKeybinding(v.View.Name(), 'a', gocui.ModNone, selectAll(v, ir.Selected)); err != nil {
		return err
	}
	if err := v.Gui.SetKeybinding(v.View.Name(), 'A', gocui.ModNone, selectAll(v, ir.Deselected)); err != nil {
		return err
	}
	return nil
}

func (v *MainView) printContent() {
	x, y := v.View.Cursor()
	_, oY := v.View.Origin()

	v.View.Clear()
	commitString := strings.TrimSpace(v.commit.String())
	fmt.Fprint(v.View, commitString)

	v.View.SetCursor(x, y)
	v.View.SetOrigin(0, oY)
}

func fixScroll(v *gocui.View) {
	_, sy := v.Size()
	_, cy := v.Cursor()
	ox, oy := v.Origin()
	if cy < oy {
		v.SetOrigin(ox, cy)
	} else if cy > oy+sy-2 {
		v.SetOrigin(ox, cy-sy+1)
	}
}

func moveCursor(v *MainView, dy int) func(*gocui.Gui, *gocui.View) error {
	return func(g *gocui.Gui, _ *gocui.View) error {
		_, cy := v.View.Cursor()
		cy += dy
		d1y := 1
		if dy < 0 {
			d1y = -1
		}
	loop:
		for {
			if cy >= len(v.commit.LineMap) || cy < 0 {
				return nil
			}
			switch val := v.commit.LineMap[cy].(type) {
			case *ir.Line:
				if val.Op == gitdiff.OpContext {
					cy += d1y
				} else {
					break loop
				}
			default:
				break loop
			}
		}
		v.View.SetCursor(0, cy)
		fixScroll(v.View)
		return nil
	}
}

func setExpansionState(v *MainView, state ir.ExpansionState) func(*gocui.Gui, *gocui.View) error {
	return func(g *gocui.Gui, _ *gocui.View) error {
		x, y := v.View.Cursor()
		i := v.commit.LineMap[y]
		switch node := i.(type) {
		case *ir.File:
			if state == ir.Expanded && node.Expanded == ir.Expanded {
				// if already expanded, go to first child
				y++
			} else {
				node.Expanded = state
			}
		case *ir.Chunk:
			if state == ir.Collapsed && node.Expanded == ir.Collapsed {
				// if already collapsed, go up to parent
				y = node.Parent.LineNumber
			} else if state == ir.Expanded && node.Expanded == ir.Expanded {
				// if already expanded, go to first child
				y++
			} else {
				node.Expanded = state
			}
		case *ir.Line:
			if state == ir.Collapsed {
				y = node.Parent.LineNumber
			}
		}
		v.printContent()
		v.View.SetCursor(x, y)
		fixScroll(v.View)
		return nil
	}
}

func setExpansionAll(v *MainView, state ir.ExpansionState) func(*gocui.Gui, *gocui.View) error {
	return func(_ *gocui.Gui, _ *gocui.View) error {
		x, y := v.View.Cursor()
		i := v.commit.LineMap[y]
		v.commit.ForEachNode(
			func(f *ir.File) error {
				f.Expanded = state
				return nil
			},
			func(_ *ir.File, c *ir.Chunk) error {
				c.Expanded = state
				return nil
			},
			nil,
		)
		v.printContent()

		if state == ir.Collapsed {
			// in this case we want to jump to the file that contained the previously-selected line
			switch node := i.(type) {
			case *ir.Chunk:
				y = node.Parent.LineNumber
			case *ir.Line:
				y = node.Parent.Parent.LineNumber
			}
			v.View.SetCursor(x, y)
		}
		fixScroll(v.View)
		return nil
	}
}

func toggleSelection(v *MainView) func(*gocui.Gui, *gocui.View) error {
	return func(g *gocui.Gui, _ *gocui.View) error {
		_, y := v.View.Cursor()
		i := v.commit.LineMap[y]
		i.ToggleSelection()
		v.printContent()
		return nil
	}
}

func confirm(v *MainView) func(*gocui.Gui, *gocui.View) error {
	return func(_ *gocui.Gui, _ *gocui.View) error {
		return ErrConfirm
	}
}

func selectAll(v *MainView, state ir.SelectionState) func(*gocui.Gui, *gocui.View) error {
	return func(_ *gocui.Gui, _ *gocui.View) error {
		v.commit.ForEachNode(
			func(f *ir.File) error {
				f.SetSelection(state)
				return nil
			},
			func(_ *ir.File, c *ir.Chunk) error {
				c.SetSelection(state)
				return nil
			},
			func(_ *ir.File, _ *ir.Chunk, l *ir.Line) error {
				l.SetSelection(state)
				return nil
			},
		)
		v.printContent()
		return nil
	}
}
