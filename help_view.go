package main

import (
	"fmt"

	"github.com/awesome-gocui/gocui"
)

const k_HelpView = "help"
const k_HelpViewHeight = 3

type HelpView struct {
	*gocui.Gui
	*gocui.View
}

func LayoutHelpView(g *gocui.Gui) (v *HelpView, err error) {
	v = &HelpView{Gui: g}
	maxX, _ := g.Size()
	v.View, err = g.SetView(k_HelpView, 0, 0, maxX-1, k_HelpViewHeight-1, 0)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return nil, err
		}
		v.printContent()
	}

	v.View.FgColor = gocui.ColorMagenta

	return v, nil
}

func (v *HelpView) printContent() {
	fmt.Fprintln(v.View, " a: select all  A: select none  q: abort  c: confirm  up/down: navigate changes  left/right: collapse/expand")
}
