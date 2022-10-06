package main

import (
	"flag"
	"log"

	"github.com/awesome-gocui/gocui"
	"github.com/smithjacobj/git-split/git"
)

const k_Debug = true

var g_TargetRef string

func init() {
	flag.Parse()
	if flag.NArg() == 0 {
		g_TargetRef = "HEAD"
	} else {
		g_TargetRef = flag.Arg(0)
	}
}

func main() {
	patch, err := git.ShowRef(g_TargetRef)
	if err != nil {
		log.Println(err)
		log.Panicln(patch.String())
	}

	commit, err := ParseCommit(patch)
	if err != nil {
		log.Panicln(err)
	}

	g, err := gocui.NewGui(gocui.OutputNormal, false)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.SetManagerFunc(layoutFn(commit))
	g.Cursor = true
	g.FgColor = gocui.ColorWhite
	g.BgColor = gocui.ColorBlack
	g.SelBgColor = gocui.ColorWhite
	g.SelFgColor = gocui.ColorBlack

	if err := setGlobalKeybindings(g); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}

func layoutFn(c *Commit) func(g *gocui.Gui) error {
	return func(g *gocui.Gui) error {
		if _, err := LayoutHelpView(g); err != nil {
			return err
		}

		if mainView, isInit, err := LayoutMainView(g); err != nil {
			return err
		} else if isInit {
			mainView.SetCommit(c)
			g.SetCurrentView(mainView.Name())
		}

		if k_Debug {
			if _, err := LayoutDebugView(g); err != nil {
				return err
			}
		}

		return nil
	}
}

func setGlobalKeybindings(g *gocui.Gui) error {
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		return err
	}
	if err := g.SetKeybinding("", 'q', gocui.ModNone, quit); err != nil {
		return err
	}
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}
