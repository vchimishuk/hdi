package diskstats

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	s := `8       0 sda 6310435 1585557 316924398 3380743 16056390 21528818 2851868384 13581986 0 7809906 16951530
8       1 sda1 1031191 1535871 20540552 347973 76303 5970882 48380560 511590 0 318246 859013
8       2 sda2 2244639 28539 97722642 1391006 4051802 3424031 89235784 1891173 0 1956150 3304383
8       3 sda3 2469809 19768 87637802 1268910 11363308 11804351 2134539768 9460193 0 4750533 10765693
8       4 sda4 564751 1379 111019266 372826 348901 329554 579712272 1392143 0 748940 1761963`

	stats, err := Parse(strings.NewReader(s))
	if err != nil {
		t.Fatal(err)
	}

	expected := []DiskStat{
		DiskStat{"sda", 316924398, 2851868384},
		DiskStat{"sda1", 20540552, 48380560},
		DiskStat{"sda2", 97722642, 89235784},
		DiskStat{"sda3", 87637802, 2134539768},
		DiskStat{"sda4", 111019266, 579712272},
	}

	if len(expected) != len(stats) {
		t.Fatalf("Expected %d disks but found %d", len(expected), len(stats))
	}
	for i := 0; i < len(expected); i++ {
		exp := expected[i]
		act := stats[exp.Dev]

		if exp != act {
			t.Fatalf("Expected %v but found %v", exp, act)
		}
	}
}
