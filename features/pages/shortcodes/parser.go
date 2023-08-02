package shortcodes

import (
	"bytes"
	"strings"

	"github.com/go-enjin/go-stdlib-text-scanner"
)

type parser struct {
	feature    *CFeature
	errHandler func(s *scanner.Scanner, msg string)
}

func (p *parser) process(input string) (nodes Nodes) {

	buf := bytes.NewBuffer([]byte(input))
	var scan scanner.Scanner
	scan.Filename = "content"
	scan.Init(buf)
	scan.Mode |= scanner.ScanInts
	scan.Mode |= scanner.ScanChars
	scan.Mode |= scanner.ScanIdents
	scan.Mode |= scanner.ScanFloats
	scan.Mode |= scanner.ScanStrings
	scan.Mode |= scanner.ScanRawStrings
	scan.Mode |= scanner.KeepComments
	scan.Whitespace ^= 1<<' ' | 1<<'\t' | 1<<'\n'
	scan.Error = p.errHandler

	var current *Node

	keepAndResetCurrent := func() {
		if current != nil {
			nodes = append(nodes, current)
			current = nil
		}
	}

	for tok := scan.Scan(); tok != scanner.EOF; tok = scan.Scan() {
		tokenText := scan.TokenText()

		switch tokenText {

		case "[":

			keepAndResetCurrent()

			if value, stopped := slurpTo(&scan, "]"); stopped != "]" {
				// stray opening square-bracket, not stopped and so is at EOF for main scan
				current = newNode(p, "", "")
				current.Append(newNode(p, "", "["))
				current.Append(p.process(value)...)
				keepAndResetCurrent()
				continue

			} else if rxNameOnly.MatchString(value) {
				// [name]
				current = newNode(p, strings.ToLower(value), "")
				current.Raw += "[" + value + "]"

			} else if rxNameValue.MatchString(value) {
				// [name=value]
				m := rxNameValue.FindAllStringSubmatch(value, 1)
				current = newNode(p, strings.ToLower(m[0][1]), "")
				current.Raw += "[" + value + "]"
				current.Attributes.Set(current.Name, unquote(m[0][2]))

			} else if rxNameKeyValues.MatchString(value) {
				// [name key=value...]
				m := rxNameKeyValues.FindAllStringSubmatch(value, 1)
				current = newNode(p, m[0][1], "")
				current.Raw += "[" + value + "]"
				if argv, keys := parseArgumentString(m[0][2]); len(keys) > 0 {
					for _, key := range keys {
						current.Attributes.Set(strings.ToLower(key), argv[key])
					}
				}

			} else {
				// none of the above formats, keep whatever it is, as-is
				current = newNode(p, "", "["+value+"]")
				keepAndResetCurrent()
				continue
			}

			if sc, ok := p.feature.LookupShortcode(current.Name); ok && sc.Inline {
				// is an inline shortcode, resume scanning
				keepAndResetCurrent()
				continue
			}

			// block shortcodes... (current is not nil)

			if content, raw, closed := slurpToClosingTag(&scan, current.Name); closed {
				// closed, parse inner content
				current.Raw += raw
				current.Append(p.process(content)...)
				keepAndResetCurrent()

			} else {
				// not closed, process further content
				nodes = append(nodes, newNode(p, "", current.Raw))
				current = nil
				for _, node := range p.process(raw) {
					nodes = append(nodes, node)
				}
			}

		default:
			if current != nil && current.Name != "" {
				keepAndResetCurrent()
			}
			if current == nil {
				current = newNode(p, "", "")
			}
			current.Raw += tokenText
			current.Content += tokenText
		}
	}

	if current != nil {
		nodes = append(nodes, current)
	}

	return
}