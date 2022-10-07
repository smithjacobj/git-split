package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/bluekeyes/go-gitdiff/gitdiff"
	"github.com/fatih/color"
)

const k_MissingSpacer = "   "
const k_DisplayTab = "    "
const k_StartExpanded = Expanded

var errBreak = fmt.Errorf("break out of callback loop")

type SelectionState int

const (
	Selected SelectionState = iota
	PartiallySelected
	Deselected
)

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
	// LineMap maps line numbers (cursor positions) to files.
	LineMap []interface{}
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
		if err := ffn(file); err != nil {
			if err == errBreak {
				break
			} else {
				return err
			}
		}
		for _, chunk := range file.Chunks {
			if err := cfn(file, chunk); err != nil {
				if err == errBreak {
					break
				} else {
					return err
				}
			}
			for _, line := range chunk.Lines {
				if err := lfn(file, chunk, line); err != nil {
					if err == errBreak {
						break
					} else {
						return err
					}
				}
			}
		}
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
			fmt.Fprint(sb, " ", f.Selected.String())
			if len(f.OldName) == 0 {
				fmt.Fprint(sb, " (NEW FILE)")
			} else {
				fmt.Fprint(sb, " ", f.OldName)
			}
			fmt.Fprint(sb, " => ")
			if len(f.NewName) == 0 {
				fmt.Fprint(sb, "(DELETED)")
			} else {
				fmt.Fprint(sb, f.NewName)
			}
			fmt.Fprintln(sb)
			return nil
		},
		func(f *File, c *Chunk) error {
			if f.Expanded == Collapsed {
				return errBreak
			}

			c.LineNumber = len(commit.LineMap)
			commit.LineMap = append(commit.LineMap, c)

			fmt.Fprint(sb, k_DisplayTab, c.Expanded.String())
			fmt.Fprintf(sb, " %s %s\n", c.Selected.String(), color.CyanString(c.Header()))
			return nil
		},
		func(f *File, c *Chunk, l *Line) error {
			if f.Expanded == Collapsed || c.Expanded == Collapsed {
				return errBreak
			}

			commit.LineMap = append(commit.LineMap, l)

			fmt.Fprint(sb, k_DisplayTab, k_DisplayTab)
			fmt.Fprintf(sb, "%s %s %s", k_MissingSpacer, l.Selected.String(), l.String())
			return nil
		},
	)
	return sb.String()
}

type File struct {
	*gitdiff.File
	Selected   SelectionState
	Expanded   ExpansionState
	LineNumber int
	Chunks     []*Chunk
}

type Chunk struct {
	*gitdiff.TextFragment
	Selected   SelectionState
	Expanded   ExpansionState
	LineNumber int
	Parent     *File
	Lines      []*Line
}

type Line struct {
	gitdiff.Line
	Selected SelectionState
	Parent   *Chunk
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
			}
		}
	}
	return commit, nil
}