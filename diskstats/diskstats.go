package diskstats

import (
	"bufio"
	"io"
	"os"
	"strconv"
	"strings"
)

type DiskStat struct {
	Dev   string
	Read  uint64
	Write uint64
}

func Parse(r io.Reader) (map[string]DiskStat, error) {
	stats := make(map[string]DiskStat)
	scnr := bufio.NewScanner(r)

	for scnr.Scan() {
		line := scnr.Text()
		cols := splitWhitespace(line)
		dev := cols[2]
		read, err := strconv.ParseUint(cols[5], 10, 64)
		if err != nil {
			return nil, err
		}
		write, err := strconv.ParseUint(cols[9], 10, 64)
		if err != nil {
			return nil, err
		}

		stats[cols[2]] = DiskStat{Dev: dev, Read: read, Write: write}
	}

	return stats, scnr.Err()
}

func ParseFile(file string) (map[string]DiskStat, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return Parse(f)
}

func splitWhitespace(s string) []string {
	parts := strings.Split(s, " ")
	res := make([]string, 0, len(parts))

	for _, s := range parts {
		if s != "" {
			res = append(res, s)
		}
	}

	return res
}
