package main

import (
	"fmt"

	"github.com/jroimartin/gocui"
)

const k_MainView = "main"

type MainView struct {
	*gocui.Gui
	*gocui.View
}

func NewMainView(g *gocui.Gui) (v *MainView, err error) {
	v = &MainView{Gui: g}
	maxX, maxY := g.Size()
	v.View, err = g.SetView(k_MainView, 0, k_HelpViewHeight-1, maxX-1, maxY-1)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return nil, err
		}
		v.printContent()
	}

	v.View.SelBgColor = gocui.ColorWhite
	v.View.SelFgColor = gocui.ColorBlack
	v.View.Frame = false
	v.View.Highlight = true

	if err := v.setKeybindings(); err != nil {
		return nil, err
	}

	return v, nil
}

func (v *MainView) setKeybindings() error {
	if err := v.Gui.SetKeybinding(k_MainView, gocui.KeyArrowUp, gocui.ModNone, moveCursorUp); err != nil {
		return err
	}
	if err := v.Gui.SetKeybinding(k_MainView, gocui.KeyArrowDown, gocui.ModNone, moveCursorDown); err != nil {
		return err
	}
	return nil
}

func (v *MainView) printContent() {
	fmt.Fprintln(v.View, "MAIN")

	// TODO:
}

func moveCursorUp(g *gocui.Gui, v *gocui.View) error {
	if _, y := v.Cursor(); y > 0 {
		v.MoveCursor(0, -1, false)
	}
	return nil
}

func moveCursorDown(g *gocui.Gui, v *gocui.View) error {
	_, maxY := v.Size()
	if _, y := v.Cursor(); y < maxY-1 {
		v.MoveCursor(0, 1, false)
	}
	return nil
}
