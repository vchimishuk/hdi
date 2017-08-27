package config

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"
)

type Command struct {
	Name string
	Args []string
}

type Device struct {
	Name string
	Time time.Duration
	Cmd  Command
}

func Parse(r io.Reader) (map[string]Device, error) {
	scnr := bufio.NewScanner(r)
	devs := make(map[string]Device)
	open := false
	var dev Device

	i := -1
	for scnr.Scan() {
		i += 1
		line := stripComment(strings.TrimSpace(scnr.Text()))
		if line == "" {
			continue
		}

		if !open {
			if isDevStart(line) {
				name := strings.TrimSpace(strings.TrimSuffix(line,
					"{"))
				dev = Device{Name: name}
				open = true
			} else {
				return nil, newError(i+1,
					"device block start expected")
			}
		} else {
			if isDevEnd(line) {
				if dev.Time == 0 {
					return nil, newError(i+1,
						"'time' parameter expected")
				}
				if dev.Cmd.Name == "" {
					return nil, newError(i+1,
						"'command' parameter expected")
				}

				devs[dev.Name] = dev
				dev = Device{}
				open = false
			} else {
				name, val, err := parseParam(line)
				if err != nil {
					return nil, newError(i+1, err.Error())
				}
				switch name {
				case "time":
					n, err := strconv.Atoi(val)
					if err != nil || n <= 0 {
						return nil, newError(i+1,
							"invalid time value")
					}
					dev.Time = time.Duration(n) * time.Minute
				case "command":
					dev.Cmd = parseCommand(val)
				default:
					return nil, newError(i+1,
						fmt.Sprintf("unexpected parameter '%s'", name))
				}
			}
		}
	}
	if err := scnr.Err(); err != nil {
		return nil, newError(i+1, err.Error())
	}
	if open {
		return nil, newError(i+1, "device block end expected")
	}

	return devs, nil
}

func ParseFile(file string) (map[string]Device, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return Parse(f)
}

func isDevStart(s string) bool {
	if strings.HasSuffix(s, "{") {
		return isLower(strings.TrimSpace(strings.TrimSuffix(s, "{")))
	} else {
		return false
	}
}

func isDevEnd(s string) bool {
	return strings.TrimSuffix(s, "}") == ""
}

func isLower(s string) bool {
	l := true

	for _, r := range s {
		l = l && unicode.IsLower(r)
	}

	return l
}

func stripComment(s string) string {
	if i := strings.Index(s, "#"); i != -1 {
		return s[:i]
	} else {
		return s
	}
}

func parseParam(s string) (string, string, error) {
	parts := strings.SplitN(s, "=", 2)
	if len(parts) != 2 {
		return "", "", errors.New("key=value format expected")
	}

	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]), nil
}

func parseCommand(s string) Command {
	ss := strings.Split(s, " ")
	if len(ss) == 0 {
		return Command{}
	} else {
		return Command{Name: ss[0], Args: ss[1:]}
	}
}
