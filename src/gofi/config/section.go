package config

import (
	"errors"
	"strconv"
	"strings"
)

// ErrInvalid is returned if the config file is invalid.
var ErrInvalid = errors.New("Invalid config")

// Section represents a component of a unifi configuration file.
type Section struct {
	Value     string
	HasValue  bool
	NamedSubs map[string]*Section
}

func prefixJoin(prefix, sectionName string) string {
	if prefix == "" {
		return sectionName
	}
	return prefix + "." + sectionName
}

func (s *Section) generate(prefix string, out *[]string) error {
	if len(s.NamedSubs) > 0 {
		//Do numbers if you can
		notNumKeys := map[string]bool{}
		for sectionName, section := range s.NamedSubs {
			_, numErr := strconv.Atoi(sectionName)
			if numErr == nil {
				section.generate(prefixJoin(prefix, sectionName), out)
			} else {
				notNumKeys[sectionName] = true
			}
		}
		//Do non-number ones last
		for sectionName := range notNumKeys {
			s.NamedSubs[sectionName].generate(prefixJoin(prefix, sectionName), out)
		}
	}
	if s.HasValue {
		outp := append(*out, prefix+"="+s.Value)
		*out = outp
	}
	return nil
}

func newSect() *Section {
	return &Section{NamedSubs: map[string]*Section{}}
}

// Parse reads a unifi config file into Sections.
func Parse(in []byte) (*Section, error) {
	out := newSect()
	lines := strings.Split(string(in), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if !strings.Contains(line, "=") {
			return nil, ErrInvalid
		}

		spl := strings.Split(line, "=")
		cursor := out

		for _, section := range strings.Split(spl[0], ".") {
			if _, ok := cursor.NamedSubs[section]; !ok {
				cursor.NamedSubs[section] = newSect()
			}
			cursor = cursor.NamedSubs[section]

		}
		cursor.Value = strings.Join(spl[1:], "=")
		cursor.HasValue = true
	}

	return out, nil
}
