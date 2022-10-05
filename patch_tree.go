package main

import (
	"bytes"
	"io"
	"strings"
)

type ActionType int

const (
	None ActionType = iota
	Add
	Remove
)

type Root struct {
	Selected bool
	Files    []*File
}

type File struct {
	Selected bool
	OldName  string
	NewName  string
	Chunks   []*Chunk
}

type Chunk struct {
	Selected bool
	Lines    []*Line
}

type Line struct {
	Selected bool
	Action   ActionType
	Content  string
}

func BuildPatchTree(buf *bytes.Buffer) (rootNode *Root, err error) {
	rootNode = &Root{
		Selected: true,
		Files:    make([]*File, 0),
	}
	currentFile := (*File)(nil)
	currentChunk := (*Chunk)(nil)
	currentLine := (*Line)(nil)

	for err == nil {
		s := ""
		if s, err = buf.ReadString('\n'); err != nil {
			break
		}

		if strings.Index(s, "diff --git") == 0 {
			currentFile = &File{Selected: true}
		} else if strings.Index(s, "---") == 0 {
			currentFile.OldName = s[4:]
		}
	}

	if err != io.EOF {
		return rootNode, err
	}
	return rootNode, nil
}
