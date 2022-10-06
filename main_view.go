package main

import (
	"fmt"
	"strings"

	"github.com/awesome-gocui/gocui"
)

const k_MainView = "main"

type MainView struct {
	*gocui.Gui
	*gocui.View

	commit *Commit
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

func (v *MainView) SetCommit(c *Commit) {
	v.commit = c
	v.printContent()
	v.View.SetCursor(0, 0)
}

func (v *MainView) setKeybindings() error {
	if err := v.Gui.SetKeybinding(v.View.Name(), gocui.KeyArrowUp, gocui.ModNone, moveCursor(0, -1)); err != nil {
		return err
	}
	if err := v.Gui.SetKeybinding(v.View.Name(), gocui.KeyArrowUp, gocui.ModShift, moveCursor(0, -15)); err != nil {
		return err
	}
	if err := v.Gui.SetKeybinding(v.View.Name(), gocui.KeyArrowDown, gocui.ModNone, moveCursor(0, 1)); err != nil {
		return err
	}
	if err := v.Gui.SetKeybinding(v.View.Name(), gocui.KeyArrowDown, gocui.ModShift, moveCursor(0, 15)); err != nil {
		return err
	}
	return nil
}

func (v *MainView) printContent() {
	v.View.Clear()
	commitString := strings.TrimSpace(v.commit.String())
	fmt.Fprint(v.View, commitString)

}

func moveCursor(x, y int) func(g *gocui.Gui, v *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		v.MoveCursor(x, y)
		return nil
	}
}
