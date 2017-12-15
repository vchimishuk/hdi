package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/vchimishuk/hdi/config"
	"github.com/vchimishuk/hdi/diskstats"
	"github.com/vchimishuk/hdi/logger"
	"github.com/vchimishuk/opt"
)

const (
	Version       = "0.1.0"
	DefaultConfig = "/etc/hdi.conf"
	DefaultLog    = "/var/log/hdi.log"
)

func fatal(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "hdi: ")
	fmt.Fprintf(os.Stderr, format, args...)
	fmt.Fprintf(os.Stderr, "\n")
	os.Exit(1)
}

func initLog(file string) error {
	l, err := logger.New(file)
	if err != nil {
		return err
	}

	log.SetOutput(l)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGHUP)
	go func() {
		for _ = range sigs {
			err := l.Reopen()
			if err != nil {
				fmt.Fprintf(os.Stderr,
					"failed to reopen log file: %s", err)
			}
		}
	}()

	return nil
}

func minDelay(cfg map[string]config.Device) time.Duration {
	min := time.Duration(0)

	for _, c := range cfg {
		t := time.Duration(int64(c.Time.Seconds())/10) * time.Second
		if min == 0 || t < min {
			min = t
		}
	}
	if min < time.Second {
		min = time.Second
	}

	return min
}

func execCmd(cmd config.Command) error {
	c := exec.Command(cmd.Name, cmd.Args...)
	stderr, err := c.StderrPipe()
	if err != nil {
		return err
	}
	if err := c.Start(); err != nil {
		stderr.Close()
		return err
	}

	out, err := ioutil.ReadAll(stderr)
	if err != nil {
		return err
	}

	if err := c.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			ws := exiterr.Sys().(syscall.WaitStatus)
			code := ws.ExitStatus()

			return fmt.Errorf("%s exited with code %d, stderr: %s",
				cmd.Name, code, out)
		} else {
			return fmt.Errorf("%s failed, stderr: %s",
				cmd.Name, out)
		}
	}

	return nil
}

func mainLoop(cfg map[string]config.Device) {
	type State struct {
		Read     uint64
		Write    uint64
		Time     time.Time
		SpunDown bool
	}

	delay := minDelay(cfg)
	states := make(map[string]*State)

	for d, _ := range cfg {
		states[d] = &State{}
	}

	for {
		stats, err := diskstats.ParseFile("/proc/diskstats")
		if err != nil {
			log.Printf("Failed to parse diskstats: %s", err)
			stats = make(map[string]diskstats.DiskStat)
		}
		for d, s := range stats {
			if t, ok := states[d]; ok {
				if s.Read != t.Read || s.Write != t.Write {
					if t.SpunDown {
						log.Printf("%s has been spun up", d)
					}
					t.Read = s.Read
					t.Write = s.Write
					t.Time = time.Now()
					t.SpunDown = false
				}

				minIdle := cfg[d].Time
				idle := time.Now().Sub(t.Time)

				if !t.SpunDown && minIdle < idle {
					log.Printf("Spining down %s", d)
					err := execCmd(cfg[d].Cmd)
					if err != nil {
						log.Printf("Failed to spin down %s: %s", d, err)
					}
					t.SpunDown = true
				}
			}
		}

		time.Sleep(delay)
	}
}

func main() {
	optDescs := []*opt.Desc{
		{"c", "config", opt.ArgString, "FILE",
			"configuration file name"},
		{"h", "help", opt.ArgNone, "",
			"display this help and exit"},
		{"l", "log", opt.ArgString, "FILE",
			"logfile name"},
		{"v", "version", opt.ArgNone, "",
			"output version information and exit"},
	}

	opts, _, err := opt.Parse(os.Args[1:], optDescs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

	if opts.Bool("help") {
		fmt.Print(opt.Usage(optDescs))
		os.Exit(0)
	}
	if opts.Bool("version") {
		fmt.Printf("%s %s", os.Args[0], Version)
		os.Exit(1)
	}

	logFile := opts.StringOr("log", DefaultLog)
	confFile := opts.StringOr("config", DefaultConfig)

	err = initLog(logFile)
	if err != nil {
		fatal("failed to open log file %s", err)
	}

	cfg, err := config.ParseFile(confFile)
	if err != nil {
		log.Fatalf("Failed to parse config: %s", err)
	}

	mainLoop(cfg)
}
