package main

import (
	"fmt"

	"github.com/awesome-gocui/gocui"
)

const k_DebugView = "debug"

type DebugView struct {
	*gocui.Gui
	*gocui.View
}

func LayoutDebugView(g *gocui.Gui) (v *DebugView, err error) {
	v = &DebugView{Gui: g}

	maxX, _ := g.Size()
	if v.View, err = g.SetView(k_DebugView, maxX-15, 0, maxX-1, k_HelpViewHeight-1, 0); err != nil {
		if err != gocui.ErrUnknownView {
			return nil, err
		}
	}

	v.printDebugValues()

	return v, nil
}

func (v *DebugView) printDebugValues() error {
	v.View.Clear()

	mainView, err := v.Gui.View(k_MainView)
	if err != nil {
		return err
	}

	curX, curY := mainView.Cursor()
	fmt.Fprintf(v.View, "(%d,%d) %d lines", curX, curY, len(mainView.BufferLines()))
	return nil
}
