package config

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestParseOk(t *testing.T) {
	s := "" +
		"# sda disk configuration.\n" +
		"sda {\n" +
		"    time = 60\n" +
		"    command = /sbin/hdparm -y /dev/sda\n" +
		"}\n" +
		"\n" +
		"sdb {\n" +
		"    time = 90\n" +
		"    command = /sbin/hdparm -y /dev/sdb\n" +
		"}\n"
	devs, err := Parse(strings.NewReader(s))
	if err != nil {
		t.Fatal(err)
	}

	assertDev(t, devs, "sda", 60*time.Minute,
		Command{"/sbin/hdparm", []string{"-y", "/dev/sda"}})
	assertDev(t, devs, "sdb", 90*time.Minute,
		Command{"/sbin/hdparm", []string{"-y", "/dev/sdb"}})
}

func TestParseEmpty(t *testing.T) {
	s := "# Test empty config."
	devs, err := Parse(strings.NewReader(s))
	if err != nil {
		t.Fatal(err)
	}
	if len(devs) != 0 {
		t.Fatal("empty config expected")
	}
}

func TestParseErr(t *testing.T) {
	s1 := "sda"
	exp1 := errors.New("1: device block start expected")
	_, err := Parse(strings.NewReader(s1))
	assertErr(t, exp1, err)

	s2 := "sda{\n foo\n }"
	exp2 := errors.New("2: key=value format expected")
	_, err = Parse(strings.NewReader(s2))
	assertErr(t, exp2, err)

	s3 := "sda{\n foo=bar\n }"
	exp3 := errors.New("2: unexpected parameter 'foo'")
	_, err = Parse(strings.NewReader(s3))
	assertErr(t, exp3, err)

	s4 := "sda{\n time=foo\n }"
	exp4 := errors.New("2: invalid time value")
	_, err = Parse(strings.NewReader(s4))
	assertErr(t, exp4, err)

	s5 := "sda{\n}"
	exp5 := errors.New("2: 'time' parameter expected")
	_, err = Parse(strings.NewReader(s5))
	assertErr(t, exp5, err)

	s6 := "sda{\ntime = 60\n}"
	exp6 := errors.New("3: 'command' parameter expected")
	_, err = Parse(strings.NewReader(s6))
	assertErr(t, exp6, err)
}

func assertDev(t *testing.T, devs map[string]Device, name string, time time.Duration, cmd Command) {
	if dev, ok := devs[name]; ok {
		if dev.Name != name {
			t.Fatalf("name '%s' expected but '%s' found",
				name, dev.Name)
		}
		if dev.Time != time {
			t.Fatalf("time %d expected but %d found",
				time, dev.Time)
		}
		if !cmdEq(dev.Cmd, cmd) {
			t.Fatalf("command '%s' expected but '%s' found",
				cmd, dev.Cmd)
		}
	} else {
		t.Fatalf("%s configuration expected", name)
	}
}

func assertErr(t *testing.T, expected error, actual error) {
	if expected == nil || actual == nil {
		t.Fatalf("'%s' error expected but '%s' found", expected, actual)
	} else {
		if expected.Error() != actual.Error() {
			t.Fatalf("'%s' error expected but '%s' found", expected, actual)
		}
	}
}

func cmdEq(a Command, b Command) bool {
	if a.Name != b.Name {
		return false
	}
	if len(a.Args) != len(b.Args) {
		return false
	}
	for i := 0; i < len(a.Args); i++ {
		if a.Args[i] != b.Args[i] {
			return false
		}
	}

	return true
}
