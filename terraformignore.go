package slug

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/scanner"
)

func parseIgnoreFile(rootPath string) []rule {
	// Do the actual file opening
	file, err := os.Open(filepath.Join(rootPath, ".terraformignore"))
	defer file.Close()

	// If there's any kind of file error, punt and use the default ignore patterns
	if err != nil {
		return defaultExclusions
	}
	return readRules(file)
}

func readRules(input io.Reader) []rule {
	rules := defaultExclusions
	scanner := bufio.NewScanner(input)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		pattern := scanner.Text()
		// Ignore blank lines
		if len(pattern) == 0 {
			continue
		}
		// Trim spaces
		pattern = strings.TrimSpace(pattern)
		// Ignore comments
		if pattern[0] == '#' {
			continue
		}
		rule := rule{}
		if pattern[0] == '!' {
			rule.excluded = true
			pattern = pattern[1:]
			// Should I add back the ! later?
		}
		// Tidy things up
		// pattern = filepath.Clean(pattern)
		// pattern = filepath.ToSlash(pattern)
		if len(pattern) > 1 && pattern[0] == '/' && !rule.excluded {
			pattern = pattern[1:]
		}
		rule.pattern = pattern
		rule.dirs = strings.Split(pattern, string(os.PathSeparator))
		rules = append(rules, rule)
	}

	if err := scanner.Err(); err != nil {
		// return nil, fmt.Errorf("Error reading .terraformignore: %v", err)
		fmt.Println("Handle error")
	}
	return rules
}

func matchIgnorePattern(path string, patterns []rule) bool {
	matched := false
	path = filepath.FromSlash(path)
	dir, filename := filepath.Split(path)
	dirSplit := strings.Split(dir, string(os.PathSeparator))

	for _, pattern := range patterns {
		negative := false

		if pattern.excluded {
			negative = true
		}
		match, err := pattern.match(path)
		if err != nil {
			return false
		}

		// If no match, try the filename alone
		if !match {
			match, err = pattern.match(filename)
		}

		if !match {
			// Filename check for current directory
			if pattern.pattern[0:1] == "/" && dir == "" {
				pattern.pattern = pattern.pattern[1:]
				pattern.compile()
				match, _ = pattern.match(filename)
			}
		}

		// Check to see if the pattern matches one of our parent dirs.
		if !match {
			// Is our rule for a directory? i.e. ends in /
			if pattern.pattern[len(pattern.pattern)-1] == os.PathSeparator {
				// does some combination of its parents match our rule?
				// Start at 1 to skip the .
				for i := 1; i < len(dirSplit); i++ {
					// From the left
					match, _ = pattern.match(strings.Join(dirSplit[:i], string(os.PathSeparator)) + string(os.PathSeparator))
					// We found a match! stop whilst ahead
					if match {
						break
					}
					// From the right
					match, _ = pattern.match(strings.Join(dirSplit[i:], string(os.PathSeparator)))
					if match {
						break
					}
				}

				// Something special if our pattern is the current directory
				// This is a case of say, ignoring terraform.d but NOT ./terraform.d/
				if pattern.pattern[0] == '/' {
					pattern.pattern = pattern.pattern[1:]
					pattern.compile()
					match, _ = pattern.match(dir)
				}
			}
		}

		if match {
			matched = !negative
		}
	}

	if matched {
		fmt.Printf("Skipping excluded path: %s \n", path)
	}

	return matched
}

type rule struct {
	pattern  string
	excluded bool
	dirs     []string
	regex    *regexp.Regexp
}

func (r *rule) match(path string) (bool, error) {
	if r.regex == nil {
		if err := r.compile(); err != nil {
			return false, filepath.ErrBadPattern
		}
	}

	b := r.regex.MatchString(path)
	// fmt.Println(path, r.pattern, r.regex, b)
	return b, nil
}

func (r *rule) compile() error {
	regStr := "^"
	pattern := r.pattern
	// Go through the pattern and convert it to a regexp.
	// Use a scanner to support utf-8 chars.
	var scan scanner.Scanner
	scan.Init(strings.NewReader(pattern))

	sl := string(os.PathSeparator)
	escSL := sl
	if sl == `\` {
		escSL += `\`
	}

	for scan.Peek() != scanner.EOF {
		ch := scan.Next()
		if ch == '*' {
			if scan.Peek() == '*' {
				// is some flavor of "**"
				scan.Next()

				// Treat **/ as ** so eat the "/"
				if string(scan.Peek()) == sl {
					scan.Next()
				}

				if scan.Peek() == scanner.EOF {
					// is "**EOF" - to align with .gitignore just accept all
					regStr += ".*"
				} else {
					// is "**"
					// Note that this allows for any # of /'s (even 0) because
					// the .* will eat everything, even /'s
					regStr += "(.*" + escSL + ")?"
				}
			} else {
				// is "*" so map it to anything but "/"
				regStr += "[^" + escSL + "]*"
			}
		} else if ch == '?' {
			// "?" is any char except "/"
			regStr += "[^" + escSL + "]"
		} else if ch == '.' || ch == '$' {
			// Escape some regexp special chars that have no meaning
			// in golang's filepath.Match
			regStr += `\` + string(ch)
		} else if ch == '\\' {
			// escape next char. Note that a trailing \ in the pattern
			// will be left alone (but need to escape it)
			if sl == `\` {
				// On windows map "\" to "\\", meaning an escaped backslash,
				// and then just continue because filepath.Match on
				// Windows doesn't allow escaping at all
				regStr += escSL
				continue
			}
			if scan.Peek() != scanner.EOF {
				regStr += `\` + string(scan.Next())
			} else {
				regStr += `\`
			}
		} else {
			regStr += string(ch)
		}
	}

	regStr += "$"
	re, err := regexp.Compile(regStr)
	if err != nil {
		return err
	}

	r.regex = re
	return nil
}

/*
	".git/",
	".terraform/",
	"!.terraform/modules/",
*/

var defaultExclusions = []rule{
	{
		pattern:  ".git/",
		excluded: false,
	},
	{
		pattern:  ".terraform/",
		excluded: false,
	},
	{
		pattern:  ".terraform/modules/",
		excluded: true,
	},
}
