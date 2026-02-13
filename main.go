package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/osm/mvdmm2/internal/fileutil"
	"github.com/osm/mvdmm2/internal/format"
	"github.com/osm/mvdmm2/internal/mvdparser"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("usage: %s <demo.mvd> [all|sum|log]\n", os.Args[0])
		os.Exit(1)
	}

	mvdPath := os.Args[1]
	mvdData, err := fileutil.ReadMVD(mvdPath)
	if err != nil {
		fmt.Printf("error: unable to open %v, %v\n", mvdPath, err)
		os.Exit(1)
	}

	p := mvdparser.New()
	stats, err := p.Parse(mvdData)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: unable to parse %v, %v\n", mvdPath, err)
		os.Exit(1)
	}

	if len(os.Args) == 3 && os.Args[2] == "all" {
		allEvents(stats.Events)
	} else if len(os.Args) == 3 && os.Args[2] == "sum" {
		sumEvents(stats.Events)
	} else {
		printMessages(stats.Messages)
	}
}

func allEvents(events []mvdparser.Event) {
	for _, ev := range events {
		if ev.Item == "" {
			fmt.Printf("%s: %s %s\n", format.Time(ev.Timestamp), ev.Name, ev.Type)
		} else {
			fmt.Printf("%s: %s %s %s\n", format.Time(ev.Timestamp), ev.Name, ev.Type, ev.Item)
		}
	}
}

func sumEvents(events []mvdparser.Event) {
	type Stats struct {
		Lost       int
		LostReport int
		Took       int
		TookReport int
	}

	stats := make(map[string]*Stats, 8)

	for _, ev := range events {
		st := stats[ev.Name]
		if st == nil {
			st = &Stats{}
			stats[ev.Name] = st
		}

		switch ev.Type {
		case mvdparser.Lost:
			st.Lost++
		case mvdparser.LostReport:
			st.LostReport++
		case mvdparser.Took:
			st.Took++
		case mvdparser.TookReport:
			st.TookReport++
		}
	}

	names := make([]string, 0, len(stats))
	for name := range stats {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		fmt.Printf(
			"name: %-16s death: %3d/%-3d took: %3d/%-3d\n",
			name,
			stats[name].LostReport, stats[name].Lost,
			stats[name].TookReport, stats[name].Took,
		)
	}
}

func printMessages(messages []mvdparser.Message) {
	for _, m := range messages {
		fmt.Printf("[%s] %s\n", format.Time(m.Timestamp), format.TrimColor(m.String))
	}
}
