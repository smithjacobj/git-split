package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/awesome-gocui/gocui"
	"github.com/fatih/color"
	"github.com/smithjacobj/git-split/git"
)

const k_Debug = true

var g_TargetRef string
var g_StartRef string

func init() {
	flag.Parse()
	if flag.NArg() == 0 {
		g_TargetRef = "HEAD"
		g_StartRef = "HEAD^"
	} else if flag.NArg() == 1 {
		g_TargetRef = flag.Arg(0)
		// by default, select the first parent of this commit. In general, for merge commits, this
		// is the branch that another ref or topic branch was merged into, e.g. main
		g_StartRef = g_TargetRef + "^"
	} else if flag.NArg() == 2 {
		g_StartRef = flag.Arg(0)
		g_TargetRef = flag.Arg(1)
	}
}

func main() {
	finishUp := false

	// we can't operate on a repo with uncommitted changes, as we will need to move around the index.
	if c, err := git.HasChanges(); err != nil {
		color.Red(err.Error())
		os.Exit(1)
	} else if c {
		color.Red("Changes detected in tracked files. Please commit or stash changes before splitting.")
		os.Exit(1)
	}

	// get a hash so the reference is valid when we move around.
	var err error
	if g_StartRef, err = git.RevParse(g_StartRef); err != nil {
		log.Panicln(err)
	}
	if g_TargetRef, err = git.RevParse(g_TargetRef); err != nil {
		log.Panicln(err)
	}

	// TODO: we will move the branch to the new leaf
	// but for now we make an alternate branch and switch to it
	originalBranchName, err := git.GetCurrentBranchName()
	if err != nil {
		log.Panicln(err)
	}

	// this creates a branch that saves the original branch state
	backupBranchNameBase := "git-split-backups/" + originalBranchName
	backupBranchName := backupBranchNameBase
	backupBranchNameNum := 0
	for git.BranchExists(backupBranchName) {
		backupBranchNameNum++
		backupBranchName = fmt.Sprintf("%s.%d", backupBranchNameBase, backupBranchNameNum)
	}
	if err := git.CreateBranch(backupBranchName); err != nil {
		log.Panicln(err)
	}

	// move to the commit before the target commit
	// FIXME: this doesn't take into account a potential 3-way merge situation, which may cause
	// spurious results. Need to either detect and abort or find a way to manage it.
	if err := git.Checkout(g_StartRef); err != nil {
		log.Panicln(err)
	}

	for {
		// get a patch format of the diff described by the selected commit
		patch, err := git.Diff("HEAD", g_TargetRef)
		fmt.Println(patch.String())
		if err != nil {
			log.Println(err)
			log.Panicln(patch.String())
		}

		commit, err := ParseCommit(patch)
		if err != nil {
			log.Panicln(err)
		} else if len(commit.Files) == 0 {
			// no more changes, rebase and quit
			if err := git.Rebase("HEAD", originalBranchName); err != nil {
				log.Panicln(err)
			}
			os.Exit(0)
		}
		commit.Description, err = git.FormatShowRefDescription(
			g_TargetRef,
			`# Original commit: %H
# Author: %an <%ae>
# Date:   %ad
#
# The original commit message is below. You may edit it as you see fit.

%B`,
		)
		if err != nil {
			log.Panicln(err)
		}

		doOnConfirm := func() error {
			git.ApplyPatch(strings.NewReader(commit.AsPatchString()))

			files := commit.GetSelectedFiles()
			if err = git.Add(files...); err != nil {
				git.Checkout(originalBranchName)
				return err
			}
			if err = git.Commit(commit.Description); err != nil {
				git.Checkout(originalBranchName)
				return err
			}
			if err = git.Amend(); err != nil {
				git.Checkout(originalBranchName)
				return err
			}
			return nil
		}

		if !finishUp {
			g, err := gocui.NewGui(gocui.OutputNormal, false)
			if err != nil {
				log.Panicln(err)
			}

			g.SetManagerFunc(layoutFn(commit))
			g.Cursor = true
			g.FgColor = gocui.ColorWhite
			g.BgColor = gocui.ColorBlack
			g.SelBgColor = gocui.ColorWhite
			g.SelFgColor = gocui.ColorBlack

			if err := setGlobalKeybindings(g); err != nil {
				log.Panicln(err)
			}

			if err := g.MainLoop(); err != nil && err != gocui.ErrQuit && err != ErrConfirm {
				log.Panicln(err)
			} else if err == gocui.ErrQuit {
				g.Close()
				git.Checkout(originalBranchName)
				os.Exit(0)
			} else if err == ErrConfirm {
				g.Close()
				if err := doOnConfirm(); err != nil {
					log.Panicln(err)
				}

				if isDifferent, err := git.IsDifferent("HEAD", g_TargetRef); err != nil {
					log.Panicln(err)
				} else if isDifferent {
					fmt.Print("Do you want to continue splitting? [Y/n]")
					nextChar := []byte{0}
					for nextChar[0] != 'y' && nextChar[0] != 'n' && nextChar[0] != '\n' {
						if _, err := os.Stdin.Read(nextChar); err != nil {
							log.Panicln(err)
						}
					}
					if nextChar[0] == 'n' {
						finishUp = true
					}
				}
			}
		} else {
			if err := doOnConfirm(); err != nil {
				log.Panicln(err)
			}
		}
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
