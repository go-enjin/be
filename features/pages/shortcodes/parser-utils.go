package shortcodes

import (
	"strings"

	"github.com/go-enjin/go-stdlib-text-scanner"
)

func unquote(input string) (output string) {
	if total := len(input); total > 2 {
		if input[0] == '"' || input[0] == '\'' {
			if input[total-1] == input[0] {
				output = input[1 : total-1]
				return
			}
		}
	}
	output = input
	return
}

func slurpTo(scan *scanner.Scanner, to ...string) (text string, stop string) {
	stoppers := map[string]struct{}{}
	for _, stopper := range to {
		stoppers[stopper] = struct{}{}
	}
	for tok := scan.Scan(); tok != scanner.EOF; tok = scan.Scan() {
		token := scan.TokenText()
		if _, stopped := stoppers[token]; stopped {
			stop = token
			return
		} else {
			text += token
		}
	}
	return
}

func slurpToClosingTag(scan *scanner.Scanner, name string) (text, raw string, closed bool) {
	var prev, token string

	for tok := scan.Scan(); tok != scanner.EOF; tok = scan.Scan() {
		token = scan.TokenText()

		if prev == "[" {

			if token == "/" {

				if value, stop := slurpTo(scan, "]"); stop != "]" {
					raw += "[/" + value
					text += "[/" + value
					return
				} else if normalized := strings.ToLower(strings.TrimSpace(value)); normalized == name {
					closed = true
					raw += "[/" + value + "]"
					return
				} else {
					raw += "[/" + value + "]"
					text += "[/" + value + "]"
					token = ""
					prev = ""
				}

			} else {

				raw += prev
				text += prev

			}

		} else {

			raw += prev
			text += prev

		}

		prev = token
	}

	raw += token
	text += token
	return
}

func parseArgumentString(input string) (argv map[string]string, keys []string) {
	argv = make(map[string]string)

	findingKey := true
	var openQuote uint8
	var key, value string
	for i := 0; i < len(input); i++ {
		if findingKey {
			if input[i] == ' ' {
				continue
			} else if input[i] == '=' {
				findingKey = false
			} else {
				key += string(input[i])
			}
		} else {
			if value == "" {
				if input[i] == '"' || input[i] == '\'' {
					openQuote = input[i]
				} else {
					value += string(input[i])
				}
			} else {
				if openQuote > 0 {
					if input[i] == openQuote {
						openQuote = 0
						findingKey = true
						argv[key] = value
						keys = append(keys, key)
						key = ""
						value = ""
					} else {
						value += string(input[i])
					}
				} else {
					if input[i] == ' ' {
						findingKey = true
						argv[key] = value
						keys = append(keys, key)
						key = ""
						value = ""
					} else {
						value += string(input[i])
					}
				}
			}
		}
	}

	if key != "" {
		argv[key] = value
		keys = append(keys, key)
	}

	return
}