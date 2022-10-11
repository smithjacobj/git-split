const k_StartExpanded = Collapsed
var errContinue = fmt.Errorf("continue to next iteration of callback loop")
				} else if err == errContinue {
					continue
			if f.IsNew {
			if f.IsDelete {
func (c *Commit) AsPatchString() string {
	sb := &strings.Builder{}
	endsWithNewline := false

	c.ForEachNode(
		func(f *File) error {
			if f.Selected == Deselected {
				return errContinue
			}

			fmt.Fprint(sb, f.Header())
			return nil
		},
		func(_ *File, c *Chunk) error {
			if c.Selected == Deselected {
				return errContinue
			}

			// we will use the outdated chunk line counts and use git-apply --recount
			fmt.Fprintln(sb, c.Header())
			return nil
		},
		func(_ *File, _ *Chunk, l *Line) error {
			if l.Selected == Deselected {
				return errContinue
			}

			s := l.String()
			endsWithNewline = strings.HasSuffix(s, "\n")
			fmt.Fprint(sb, s)
			return nil
		},
	)

	if !endsWithNewline {
		fmt.Fprint(sb, "\n\\ No newline at end of file")
	}

	return sb.String()
}

func (c *Commit) GetSelectedFiles() []string {
	ss := make([]string, 0, len(c.Files))
	for _, file := range c.Files {
		if file.Selected != Deselected {
			ss = append(ss, file.NewName)
		}
	}
	return ss
}

				} else if err == errContinue {
					continue
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

				} else if err == errContinue {
					continue