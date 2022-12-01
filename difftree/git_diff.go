package difftree

import (
	"fmt"
	"io"
	"strings"

	"github.com/bluekeyes/go-gitdiff/gitdiff"
	"github.com/fatih/color"
)

const k_MissingSpacer = "   "
const k_DisplayTab = "    "
const k_StartExpanded = Collapsed
const k_PanicPartialSelection = "PartiallySelected is not a valid manual selection"
const k_NoEOL = "\\ No newline at end of file\n"

var ErrBreak = fmt.Errorf("break out of callback loop")
var ErrContinue = fmt.Errorf("continue to next iteration of callback loop")

type SelectionState int

const (
	Selected SelectionState = iota
	PartiallySelected
	Deselected
)

type Selectable interface {
	ToggleSelection()
	SetSelection(SelectionState)
}

func (s SelectionState) String() string {
	switch s {
	case Selected:
		return "[*]"
	case PartiallySelected:
		return "[~]"
	case Deselected:
		return "[ ]"
	}
	return k_MissingSpacer
}

func (s *SelectionState) Toggle() {
	if *s == Selected || *s == PartiallySelected {
		*s = Deselected
	} else {
		*s = Selected
	}
}

type ExpansionState int

const (
	Collapsed ExpansionState = iota
	Expanded
)

func (e ExpansionState) String() string {
	switch e {
	case Collapsed:
		return "(+)"
	case Expanded:
		return "(-)"
	}
	return k_MissingSpacer
}

type Commit struct {
	Files []*File
	// LineMap maps line numbers (cursor positions) to selection nodes (files, chunks, lines).
	LineMap []Selectable
	// Description includes the commit details, like commit message, etc.
	Description string
}

// FileFunc is a callback for ForEachNode. Return an error to break out of the loop.
type FileFunc func(*File) error

// ChunkFunc is a callback for ForEachNode. Return an error to break out of the sub-loop.
type ChunkFunc func(*File, *Chunk) error

// LineFunc is a callback for ForEachNode. Return an error to break out of the sub-loop.
type LineFunc func(*File, *Chunk, *Line) error

func (c *Commit) ForEachNode(ffn FileFunc, cfn ChunkFunc, lfn LineFunc) error {
	for _, file := range c.Files {
		if ffn != nil {
			if err := ffn(file); err != nil {
				if err == ErrBreak {
					break
				} else if err == ErrContinue {
					continue
				} else {
					return err
				}
			}
		}
		file.ForEachNode(cfn, lfn)
	}
	return nil
}

func (commit *Commit) String() string {
	sb := &strings.Builder{}
	commit.LineMap = commit.LineMap[:0]
	commit.ForEachNode(
		func(f *File) error {
			f.LineNumber = len(commit.LineMap)
			commit.LineMap = append(commit.LineMap, f)

			fmt.Fprint(sb, f.Expanded.String())
			fmt.Fprint(sb, " ", f.selection.String())
			if f.IsNew {
				fmt.Fprint(sb, " (NEW FILE)")
			} else {
				fmt.Fprint(sb, " ", f.OldName)
			}
			fmt.Fprint(sb, " => ")
			if f.IsDelete {
				fmt.Fprint(sb, "(DELETED)")
			} else {
				fmt.Fprint(sb, f.NewName)
			}
			fmt.Fprintln(sb)
			return nil
		},
		func(f *File, c *Chunk) error {
			if f.Expanded == Collapsed {
				return ErrBreak
			}

			c.LineNumber = len(commit.LineMap)
			commit.LineMap = append(commit.LineMap, c)

			fmt.Fprint(sb, k_DisplayTab, c.Expanded.String())
			fmt.Fprintf(sb, " %s %s\n", c.selection.String(), color.CyanString(c.Header()))
			return nil
		},
		func(f *File, c *Chunk, l *Line) error {
			if f.Expanded == Collapsed || c.Expanded == Collapsed {
				return ErrBreak
			}

			commit.LineMap = append(commit.LineMap, l)

			fmt.Fprint(sb, k_DisplayTab, k_DisplayTab)
			lineColor := color.FgWhite
			if l.Op == gitdiff.OpAdd {
				lineColor = color.FgGreen
			} else if l.Op == gitdiff.OpDelete {
				lineColor = color.FgRed
			}
			// aligns as there's no collapse/expand on lines
			fmt.Fprint(sb, k_MissingSpacer, " ")
			if l.Op == gitdiff.OpContext {
				// selecting or deselecting context lines is pointless
				fmt.Fprint(sb, k_MissingSpacer)
			} else {
				fmt.Fprint(sb, l.selection.String())
			}

			fmt.Fprintf(sb, " \u001b[%dm%s\u001b[%dm", lineColor, l.String(), color.FgWhite)

			if l.NoEOL() {
				// we make sure that the line map remains normalized even with this added virtual line.
				commit.LineMap = append(commit.LineMap, l)
				fmt.Fprint(sb, "\n", k_DisplayTab, k_DisplayTab, k_DisplayTab, k_DisplayTab)
				fmt.Fprintf(sb, "\u001b[%dm%s%s\u001b[%dm", lineColor, l.Op.String(), k_NoEOL, color.FgWhite)
			}
			return nil
		},
	)
	return sb.String()
}

func (c *Commit) AsPatchString() string {
	sb := &strings.Builder{}

	c.ForEachNode(
		func(f *File) error {
			if f.selection == Deselected {
				return ErrContinue
			}

			fmt.Fprint(sb, f.Header())
			return nil
		},
		func(_ *File, c *Chunk) error {
			if c.selection == Deselected {
				return ErrContinue
			}

			// we will use the outdated chunk line counts and use git-apply --recount
			fmt.Fprintln(sb, c.Header())
			return nil
		},
		func(_ *File, _ *Chunk, l *Line) error {
			s := l.String()

			if l.selection == Deselected {
				if l.Op == gitdiff.OpAdd {
					return ErrContinue
				} else if l.Op == gitdiff.OpDelete {
					// removing OpDeletes makes the patch fail, so we change them into context lines
					// for patches.
					s = gitdiff.OpContext.String() + l.Line.Line
				}
			}

			fmt.Fprint(sb, s)
			if l.NoEOL() {
				fmt.Fprint(sb, "\n", k_NoEOL)
			}
			return nil
		},
	)

	return sb.String()
}

func (c *Commit) GetSelectedFiles() []string {
	ss := make([]string, 0, len(c.Files))
	for _, file := range c.Files {
		if file.selection != Deselected {
			ss = append(ss, file.NewName)
		}
	}
	return ss
}

type File struct {
	*gitdiff.File
	selection  SelectionState
	Expanded   ExpansionState
	LineNumber int
	Chunks     []*Chunk
}

func (f *File) ToggleSelection() {
	f.selection.Toggle()
	f.afterSelection()
}

func (f *File) SetSelection(state SelectionState) {
	f.selection = state
	f.afterSelection()
}

func (f *File) afterSelection() {
	f.ForEachNode(
		func(_ *File, c *Chunk) error {
			c.selection = f.selection
			return nil
		},
		func(_ *File, c *Chunk, l *Line) error {
			l.selection = f.selection
			return nil
		},
	)
}

func (file *File) ForEachNode(cfn ChunkFunc, lfn LineFunc) error {
	for _, chunk := range file.Chunks {
		if cfn != nil {
			if err := cfn(file, chunk); err != nil {
				if err == ErrBreak {
					break
				} else if err == ErrContinue {
					continue
				} else {
					return err
				}
			}
		}
		chunk.ForEachNode(lfn)
	}
	return nil
}

func (file *File) UpdateSelection() {
	selectedChunkCount := 0
	partiallySelectedChunkCount := 0
	file.ForEachNode(
		func(f *File, c *Chunk) error {
			if c.selection == Selected {
				selectedChunkCount++
			} else if c.selection == PartiallySelected {
				partiallySelectedChunkCount++
			}
			return nil
		},
		nil,
	)
	if selectedChunkCount == len(file.Chunks) {
		file.selection = Selected
	} else if selectedChunkCount > 0 || partiallySelectedChunkCount > 0 {
		file.selection = PartiallySelected
	} else {
		file.selection = Deselected
	}
}

func (file *File) Header() string {
	sb := &strings.Builder{}

	if file.IsNew {
		fmt.Fprintf(sb, "diff --git a/%s b/%s\n", file.NewName, file.NewName)
	} else if file.IsDelete {
		fmt.Fprintf(sb, "diff --git a/%s b/%s\n", file.OldName, file.OldName)
	} else {
		fmt.Fprintf(sb, "diff --git a/%s b/%s\n", file.OldName, file.NewName)
	}

	if file.IsCopy {
		fmt.Fprintf(sb, "copy from %s\n", file.OldName)
		fmt.Fprintf(sb, "copy to %s\n", file.NewName)
	} else if file.IsRename {
		fmt.Fprintf(sb, "rename from %s\n", file.OldName)
		fmt.Fprintf(sb, "rename to %s\n", file.NewName)
	} else if file.IsNew {
		fmt.Fprintf(sb, "new file mode %06o\n", file.NewMode)
		fmt.Fprint(sb, "--- /dev/null\n")
		fmt.Fprintf(sb, "+++ b/%s\n", file.NewName)
	} else if file.IsDelete {
		fmt.Fprintf(sb, "deleted file mode %06o\n", file.OldMode)
		fmt.Fprintf(sb, "--- a/%s\n", file.OldName)
		fmt.Fprint(sb, "+++ /dev/null\n")
	} else {
		if file.NewMode != 0 && file.OldMode != file.NewMode {
			fmt.Fprintf(sb, "old mode %06o\n", file.OldMode)
			fmt.Fprintf(sb, "new mode %06o\n", file.NewMode)
		}
		fmt.Fprintf(sb, "--- a/%s\n", file.OldName)
		fmt.Fprintf(sb, "+++ b/%s\n", file.NewName)
		// we leave out object IDs as splits should never need to 3-way merge and the new OID
		// will be invalid until we create the new commit.
	}
	return sb.String()
}

type Chunk struct {
	*gitdiff.TextFragment
	selection           SelectionState
	Expanded            ExpansionState
	LineNumber          int
	Parent              *File
	Lines               []*Line
	NonContextLineCount int
}

func (c *Chunk) ToggleSelection() {
	c.selection.Toggle()
	c.afterSelection()
}

func (c *Chunk) SetSelection(state SelectionState) {
	if state == PartiallySelected {
		panic(k_PanicPartialSelection)
	}
	c.selection = state
	c.afterSelection()
}

func (c *Chunk) afterSelection() {
	c.ForEachNode(func(_ *File, _ *Chunk, l *Line) error {
		l.selection = c.selection
		return nil
	})
	c.Parent.UpdateSelection()
}

func (chunk *Chunk) ForEachNode(lfn LineFunc) error {
	for _, line := range chunk.Lines {
		if lfn != nil {
			if err := lfn(chunk.Parent, chunk, line); err != nil {
				if err == ErrBreak {
					break
				} else if err == ErrContinue {
					continue
				} else {
					return err
				}
			}
		}
	}
	return nil
}

func (chunk *Chunk) UpdateSelection() {
	selectedLineCount := 0
	partiallySelectedLineCount := 0
	chunk.ForEachNode(
		func(f *File, c *Chunk, l *Line) error {
			if l.Op != gitdiff.OpContext {
				if l.selection == Selected {
					selectedLineCount++
				} else if l.selection == PartiallySelected {
					partiallySelectedLineCount++
				}
			}
			return nil
		},
	)
	if selectedLineCount == chunk.NonContextLineCount {
		chunk.selection = Selected
	} else if selectedLineCount > 0 || partiallySelectedLineCount > 0 {
		chunk.selection = PartiallySelected
	} else {
		chunk.selection = Deselected
	}
	chunk.Parent.UpdateSelection()
}

type Line struct {
	gitdiff.Line
	selection SelectionState
	Parent    *Chunk
}

func (l *Line) ToggleSelection() {
	l.selection.Toggle()
	l.Parent.UpdateSelection()
}

func (l *Line) SetSelection(state SelectionState) {
	if state == PartiallySelected {
		panic(k_PanicPartialSelection)
	}
	l.selection = state
	l.Parent.UpdateSelection()
}

func ParseCommit(r io.Reader) (commit *Commit, err error) {
	files, desc, err := gitdiff.Parse(r)
	if err != nil {
		return nil, err
	}

	commit = &Commit{Files: make([]*File, 0, len(files)), Description: desc}
	for _, file := range files {
		outFile := &File{File: file, Expanded: k_StartExpanded}
		commit.Files = append(commit.Files, outFile)
		for _, chunk := range file.TextFragments {
			outChunk := &Chunk{TextFragment: chunk, Expanded: k_StartExpanded, Parent: outFile}
			outFile.Chunks = append(outFile.Chunks, outChunk)
			for _, line := range chunk.Lines {
				outLine := &Line{Line: line, Parent: outChunk}
				outChunk.Lines = append(outChunk.Lines, outLine)
				if line.Op != gitdiff.OpContext {
					outChunk.NonContextLineCount++
				}
			}
		}
	}
	return commit, nil
}
