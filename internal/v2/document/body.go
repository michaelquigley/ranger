package document

import (
	"net/url"
	"path"
	"regexp"
	"strings"

	"github.com/michaelquigley/ranger/internal/v2/model"
)

const backtick = '\x60'

var (
	promptEnd  = regexp.MustCompile(`^ {0,3}##(?:[ \t]|$)`)
	checkbox   = regexp.MustCompile(`^[ \t]*(?:[-+*]|[0-9]+[.)])[ \t]+\[([ xX])\](?:[ \t]|$)`)
	inlineLink = regexp.MustCompile(`\[[^\[\]]*\]\([ \t]*(<[^<>]+>|[^() \t]+?)(?:[ \t]+(?:"[^"]*"|'[^']*'))?[ \t]*\)`)
)

func scanBody(body, slug string) (model.Checklist, []string) {
	var count model.Checklist
	var targets []string
	var fence byte
	fenceLength := 0
	inPrompt := true
	for line := range strings.SplitSeq(body, "\n") {
		line = strings.TrimSuffix(line, "\r")
		trimmed := strings.TrimSpace(line)
		if fence != 0 {
			length := runLength(trimmed, fence)
			if length >= fenceLength && length == len(trimmed) {
				fence = 0
			}
			continue
		}
		if trimmed != "" && (trimmed[0] == backtick || trimmed[0] == '~') {
			length := runLength(trimmed, trimmed[0])
			if length >= 3 && (trimmed[0] == '~' || !strings.ContainsRune(trimmed[length:], backtick)) {
				fence, fenceLength = trimmed[0], length
				continue
			}
		}
		if promptEnd.MatchString(line) {
			inPrompt = false
		}
		if inPrompt {
			if match := checkbox.FindStringSubmatch(line); match != nil {
				count.Total++
				if match[1] == " " {
					count.Open++
				}
			}
		}
		line = maskCodeSpans(line)
		for _, match := range inlineLink.FindAllStringSubmatchIndex(line, -1) {
			start := match[0]
			if escaped(line, start) || start > 0 && line[start-1] == '!' && !escaped(line, start-1) {
				continue
			}
			if target := resolveBodyTarget(slug, line[match[2]:match[3]]); target != "" {
				targets = append(targets, target)
			}
		}
	}
	return count, targets
}

// maskCodeSpans blanks matched backtick runs so code cannot introduce links.
// it keeps the label of an ordinary link parseable when it contains code.
func maskCodeSpans(line string) string {
	masked := []byte(line)
	for i := 0; i < len(line); {
		if line[i] != backtick || escaped(line, i) {
			i++
			continue
		}
		length := runLength(line[i:], backtick)
		start := i
		i += length
		for end := i; end < len(line); {
			offset := strings.IndexByte(line[end:], backtick)
			if offset == -1 {
				break
			}
			end += offset
			run := runLength(line[end:], backtick)
			end += run
			if run == length {
				for j := start; j < end; j++ {
					masked[j] = ' '
				}
				i = end
				break
			}
		}
	}
	return string(masked)
}

func runLength(text string, marker byte) int {
	i := 0
	for i < len(text) && text[i] == marker {
		i++
	}
	return i
}

func escaped(line string, index int) bool {
	backslashes := 0
	for i := index - 1; i >= 0 && line[i] == '\\'; i-- {
		backslashes++
	}
	return backslashes%2 == 1
}

func resolveBodyTarget(slug, destination string) string {
	destination = strings.TrimSuffix(strings.TrimPrefix(destination, "<"), ">")
	target, err := url.Parse(destination)
	if err != nil || target.IsAbs() || target.Host != "" || target.Path == "" || strings.HasPrefix(target.Path, "/") {
		return ""
	}
	const future = "/docs/future/"
	resolved := path.Join(future, slug, target.Path)
	relative, withinFuture := strings.CutPrefix(resolved, future)
	name, _, _ := strings.Cut(relative, "/")
	if !withinFuture || name == slug || strings.HasPrefix(name, ".") {
		return ""
	}
	return name
}
