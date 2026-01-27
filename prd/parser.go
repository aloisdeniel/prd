package prd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	east "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// Parse parses a PRD markdown document into a Feature.
func Parse(input string) (Feature, error) {
	source := []byte(input)

	md := goldmark.New(
		goldmark.WithExtensions(extension.TaskList),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
	)
	doc := md.Parser().Parse(text.NewReader(source))

	var f Feature

	// Collect all top-level nodes into a slice for indexed access.
	var nodes []ast.Node
	for n := doc.FirstChild(); n != nil; n = n.NextSibling() {
		nodes = append(nodes, n)
	}

	// leafStart returns the first byte offset of a node by scanning its leaf segments.
	var leafStart func(ast.Node) int
	leafStart = func(n ast.Node) int {
		if lines := n.Lines(); lines.Len() > 0 {
			return lines.At(0).Start
		}
		for c := n.FirstChild(); c != nil; c = c.NextSibling() {
			if s := leafStart(c); s >= 0 {
				return s
			}
		}
		return -1
	}

	// headingLineStart returns the byte offset of the beginning of the line
	// containing the heading (i.e. before the '#' markers).
	headingLineStart := func(h *ast.Heading) int {
		s := leafStart(h)
		if s < 0 {
			return -1
		}
		// Scan backwards past the "### " prefix (space + '#' markers)
		pos := s - 1
		// Skip spaces between '#' and text
		for pos >= 0 && source[pos] == ' ' {
			pos--
		}
		// Skip '#' markers
		for pos >= 0 && source[pos] == '#' {
			pos--
		}
		return pos + 1
	}

	// nodeByteStart returns the start byte for a node — for headings, the line
	// start (before '#' markers); for other nodes, the leaf start.
	nodeByteStart := func(n ast.Node) int {
		if h, ok := n.(*ast.Heading); ok {
			return headingLineStart(h)
		}
		return leafStart(n)
	}

	// rawSection extracts raw markdown text between after a heading node and
	// before the next sibling heading of level <= the given level.
	rawSection := func(idx int, level int) string {
		startByte := -1
		endByte := len(source)

		for j := idx + 1; j < len(nodes); j++ {
			s := nodeByteStart(nodes[j])
			if s >= 0 {
				startByte = s
				break
			}
		}
		if startByte < 0 {
			return ""
		}

		for j := idx + 1; j < len(nodes); j++ {
			if h, ok := nodes[j].(*ast.Heading); ok && h.Level <= level {
				s := nodeByteStart(nodes[j])
				if s >= 0 {
					endByte = s
				}
				break
			}
		}

		return strings.TrimSpace(string(source[startByte:endByte]))
	}

	// contentAfterHeading extracts text from after a heading until the next heading of any kind.
	contentAfterHeading := func(idx int) string {
		startByte := -1
		endByte := len(source)

		for j := idx + 1; j < len(nodes); j++ {
			if startByte < 0 {
				if _, ok := nodes[j].(*ast.Heading); ok {
					return ""
				}
				s := nodeByteStart(nodes[j])
				if s >= 0 {
					startByte = s
				}
				continue
			}
			if _, ok := nodes[j].(*ast.Heading); ok {
				s := nodeByteStart(nodes[j])
				if s >= 0 {
					endByte = s
				}
				break
			}
		}
		if startByte < 0 {
			return ""
		}
		return strings.TrimSpace(string(source[startByte:endByte]))
	}

	// headingText extracts the text content of a heading node.
	headingText := func(n *ast.Heading) string {
		var sb strings.Builder
		for c := n.FirstChild(); c != nil; c = c.NextSibling() {
			if t, ok := c.(*ast.Text); ok {
				sb.Write(t.Segment.Value(source))
			} else {
				for gc := c.FirstChild(); gc != nil; gc = gc.NextSibling() {
					if t2, ok := gc.(*ast.Text); ok {
						sb.Write(t2.Segment.Value(source))
					}
				}
			}
		}
		return sb.String()
	}

	type h2Section int
	const (
		h2None h2Section = iota
		h2Description
		h2UserStories
		h2Notes
	)

	type h4Sub int
	const (
		h4None h4Sub = iota
		h4AcceptanceCriteria
	)

	curH2 := h2None
	curH4 := h4None
	var currentStory *UserStory
	var currentNote *Note

	flushStory := func() {
		if currentStory != nil {
			f.UserStories = append(f.UserStories, *currentStory)
			currentStory = nil
		}
		curH4 = h4None
	}

	flushNote := func() {
		if currentNote != nil {
			f.Notes = append(f.Notes, *currentNote)
			currentNote = nil
		}
	}

	for i, n := range nodes {
		switch node := n.(type) {
		case *ast.Heading:
			switch node.Level {
			case 1:
				f.Name = strings.TrimSpace(headingText(node))
				f.Description = contentAfterHeading(i)
				curH2 = h2Description

			case 2:
				if curH2 == h2UserStories {
					flushStory()
				} else if curH2 == h2Notes {
					flushNote()
				}
				curH4 = h4None

				heading := strings.TrimSpace(headingText(node))
				switch heading {
				case "Goals":
					curH2 = h2None
					f.Goals = rawSection(i, 2)
				case "User Stories":
					curH2 = h2UserStories
				case "Functional Requirements":
					curH2 = h2None
					f.FunctionalRequirements = rawSection(i, 2)
				case "Non-Goals":
					curH2 = h2None
					f.NonGoals = rawSection(i, 2)
				case "Technical Considerations":
					curH2 = h2None
					f.TechnicalConsiderations = rawSection(i, 2)
				case "Analytics & Instrumentation":
					curH2 = h2None
					f.Analytics = rawSection(i, 2)
				case "Notes":
					curH2 = h2Notes
				case "Risks & Mitigations":
					curH2 = h2None
					f.Risks = rawSection(i, 2)
				case "Success Metrics":
					curH2 = h2None
					f.SuccessMetrics = rawSection(i, 2)
				case "Open Questions":
					curH2 = h2None
					f.OpenQuestions = rawSection(i, 2)
				case "Introduction":
					curH2 = h2Description
					f.Description = rawSection(i, 2)
				default:
					curH2 = h2None
				}

			case 3:
				switch curH2 {
				case h2UserStories:
					flushStory()
					hText := strings.TrimSpace(headingText(node))
					story, err := parseUserStoryHeading(hText)
					if err != nil {
						return f, err
					}
					currentStory = &story
					curH4 = h4None
					currentStory.Description = contentAfterHeading(i)
				case h2Notes:
					flushNote()
					date := strings.TrimSpace(headingText(node))
					currentNote = &Note{Date: date}
					currentNote.Content = contentAfterHeading(i)
				}

			case 4:
				if curH2 == h2UserStories && currentStory != nil {
					h4Text := strings.TrimSpace(headingText(node))
					switch h4Text {
					case "Acceptance Criteria":
						curH4 = h4AcceptanceCriteria
						currentStory.Description = strings.TrimSpace(currentStory.Description)
					case "Technical Considerations":
						curH4 = h4None
						currentStory.TechnicalConsiderations = contentAfterHeading(i)
					default:
						curH4 = h4None
					}
				}
			}

		case *ast.List:
			if curH2 == h2UserStories && curH4 == h4AcceptanceCriteria && currentStory != nil {
				for li := node.FirstChild(); li != nil; li = li.NextSibling() {
					listItem, ok := li.(*ast.ListItem)
					if !ok {
						continue
					}
					completed := false
					hasCheckbox := false
					// First block child (Paragraph or TextBlock) contains the checkbox
					if blk := listItem.FirstChild(); blk != nil {
						if fc := blk.FirstChild(); fc != nil {
							if cb, ok2 := fc.(*east.TaskCheckBox); ok2 {
								hasCheckbox = true
								completed = cb.IsChecked
							}
						}
					}
					if !hasCheckbox {
						continue
					}
					var textParts []string
					for blk := listItem.FirstChild(); blk != nil; blk = blk.NextSibling() {
						for inline := blk.FirstChild(); inline != nil; inline = inline.NextSibling() {
							if _, ok2 := inline.(*east.TaskCheckBox); ok2 {
								continue
							}
							if t, ok2 := inline.(*ast.Text); ok2 {
								textParts = append(textParts, string(t.Segment.Value(source)))
							}
						}
					}
					currentStory.AcceptanceCriteria = append(currentStory.AcceptanceCriteria, AcceptanceCriterion{
						Text:      strings.TrimSpace(strings.Join(textParts, "")),
						Completed: completed,
					})
				}
			}

		case *ast.Paragraph:
			for c := node.FirstChild(); c != nil; c = c.NextSibling() {
				if emph, ok := c.(*ast.Emphasis); ok {
					var sb strings.Builder
					for ec := emph.FirstChild(); ec != nil; ec = ec.NextSibling() {
						if t, ok := ec.(*ast.Text); ok {
							sb.Write(t.Segment.Value(source))
						}
					}
					txt := sb.String()
					if strings.HasPrefix(txt, "Document Version:") {
						f.Version = strings.TrimSpace(strings.TrimPrefix(txt, "Document Version:"))
					} else if strings.HasPrefix(txt, "Last Updated:") {
						f.LastUpdated = strings.TrimSpace(strings.TrimPrefix(txt, "Last Updated:"))
					}
				}
			}
		}
	}

	// Final flush
	switch curH2 {
	case h2UserStories:
		flushStory()
	case h2Notes:
		flushNote()
	}

	return f, nil
}

func parseUserStoryHeading(heading string) (UserStory, error) {
	heading = strings.TrimSpace(heading)
	parts := strings.SplitN(heading, "|", 3)

	story := UserStory{Priority: 3}

	if len(parts) < 2 {
		return story, fmt.Errorf("invalid user story heading: %q", heading)
	}

	// Parse US-X ID
	idPart := strings.TrimSpace(parts[0])
	if !strings.HasPrefix(idPart, "US-") {
		return story, fmt.Errorf("invalid user story ID format: %q", idPart)
	}
	id, err := strconv.Atoi(strings.TrimLeft(idPart[3:], "0"))
	if err != nil {
		// Handle US-0 edge case
		if idPart[3:] == "0" || strings.TrimLeft(idPart[3:], "0") == "" {
			id = 0
		} else {
			return story, fmt.Errorf("invalid user story ID number: %q", idPart)
		}
	}
	story.ID = id

	// Determine if second part is priority or name
	secondPart := strings.TrimSpace(parts[1])

	if len(parts) == 3 {
		// US-X | PX | Name
		if !strings.HasPrefix(secondPart, "P") {
			return story, fmt.Errorf("invalid priority format: %q", secondPart)
		}
		p, err := strconv.Atoi(secondPart[1:])
		if err != nil {
			return story, fmt.Errorf("invalid priority number: %q", secondPart)
		}
		story.Priority = p
		story.Name = strings.TrimSpace(parts[2])
	} else {
		// US-X | Name (no priority)
		story.Name = secondPart
	}

	return story, nil
}
