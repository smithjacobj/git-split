package main

import (
	"fmt"

	"github.com/awesome-gocui/gocui"
	"github.com/fatih/color"
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
	v.printKeybind("space", "toggle selection")
	v.printKeybind("a", "select all")
	v.printKeybind("A", "select none")
	v.printKeybind("q", "abort")
	v.printKeybind("c", "confirm")
	v.printKeybind("up/down", "navigate")
	v.printKeybind("left/right", "collapse/expand")
}

func (v *HelpView) printKeybind(key, usage string) {
	fmt.Fprint(v.View, color.CyanString(key))
	fmt.Fprintf(v.View, ": %s ", usage)
}
