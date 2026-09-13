package gitmod

import (
	"fmt"
	"strings"
)

// Pin is one submodule as .gitmodules declares it.
type Pin struct {
	Path string
	URL  string
	Rev  string
}

// Pins reads every submodule .gitmodules declares, with the url and the rev
// each one carries.
//
// A module zip carries .gitmodules, because it is a tracked file, and carries
// nothing of the submodule itself. So a consumer resolving this module from the
// proxy has these three fields and an empty directory. That is what lets the
// generate step fetch its own input. The rev is a key git does not read, and
// git leaves a key it does not know alone.
func Pins(root string) ([]Pin, error) {
	out, ok := git(root, "config", "--file", ".gitmodules", "--list")
	if !ok {
		return nil, fmt.Errorf("cannot read %s/.gitmodules", root)
	}
	byName := map[string]*Pin{}
	var order []string
	for _, line := range strings.Split(out, "\n") {
		key, value, found := strings.Cut(strings.TrimSpace(line), "=")
		if !found || !strings.HasPrefix(key, "submodule.") {
			continue
		}
		name, field, found := cutLast(strings.TrimPrefix(key, "submodule."), ".")
		if !found {
			continue
		}
		pin, seen := byName[name]
		if !seen {
			pin = &Pin{}
			byName[name] = pin
			order = append(order, name)
		}
		switch field {
		case "path":
			pin.Path = strings.TrimRight(value, "/")
		case "url":
			pin.URL = value
		case "rev":
			pin.Rev = value
		}
	}
	pins := make([]Pin, 0, len(order))
	for _, name := range order {
		pins = append(pins, *byName[name])
	}
	return pins, nil
}

// GitlinkRev reads the commit the index records for a submodule path. It
// answers "" where there is no gitlink, which is every consumer that resolved
// this module from the proxy.
func GitlinkRev(root, path string) string {
	staged, ok := git(root, "ls-files", "--stage", "--", path)
	if !ok {
		return ""
	}
	fields := strings.Fields(strings.TrimSpace(staged))
	if len(fields) < 2 || fields[0] != gitlinkMode {
		return ""
	}
	return fields[1]
}

// cutLast splits s at the LAST separator. A submodule name may hold a dot, so
// cutting at the first one reads part of the name as the field.
func cutLast(s, sep string) (before, after string, found bool) {
	i := strings.LastIndex(s, sep)
	if i < 0 {
		return s, "", false
	}
	return s[:i], s[i+len(sep):], true
}
