package mvdparser

import (
	"strconv"
	"strings"
	"time"

	"github.com/osm/quake/common/args"
	"github.com/osm/quake/common/ascii"
	"github.com/osm/quake/common/context"
	"github.com/osm/quake/common/death"
	"github.com/osm/quake/common/infostring"
	"github.com/osm/quake/demo/mvd"
	"github.com/osm/quake/packet/command/ftespawnbaseline"
	"github.com/osm/quake/packet/command/modellist"
	"github.com/osm/quake/packet/command/print"
	"github.com/osm/quake/packet/command/spawnbaseline"
	"github.com/osm/quake/packet/command/stufftext"
	"github.com/osm/quake/packet/command/updateuserinfo"
	"github.com/osm/quake/packet/svc"

	"github.com/osm/mvdmm2/internal/format"
	"github.com/osm/mvdmm2/internal/spam"
)

type Type string

const (
	Lost       Type = "lost"
	LostReport Type = "lost-report"
	Took       Type = "took"
	TookReport Type = "took-report"
)

type Event struct {
	Timestamp float64
	Name      string
	Type      Type
	Item      string
}

type entity struct {
	coord [3]float32
	model byte
	skin  byte
}

type Parser struct {
	elapsed  float64
	entities map[uint16]*entity
	filter   *spam.Filter
	models   []string
	players  map[byte]string
	started  bool
	events   []Event
}

func New() *Parser {
	return &Parser{
		entities: make(map[uint16]*entity),
		filter:   spam.New(time.Second * 1),
		players:  make(map[byte]string),
	}
}

func (p *Parser) Parse(data []byte) ([]Event, error) {
	demo, err := mvd.Parse(context.New(), data)
	if err != nil {
		return nil, err
	}

	for _, d := range demo.Data {
		if p.started {
			p.elapsed += float64(d.Timestamp) * 0.001
		}

		if d.Read == nil {
			continue
		}

		gd, ok := d.Read.Packet.(*svc.GameData)
		if !ok {
			continue
		}

		var err error
		for _, cmd := range gd.Commands {
			switch c := cmd.(type) {
			case *modellist.Command:
				p.handleModellist(c)
			case *spawnbaseline.Command:
				p.handleSpawnBaseline(c)
			case *ftespawnbaseline.Command:
				p.handleFTESpawnBaseline(c)
			case *updateuserinfo.Command:
				p.handleUpdateUserinfo(c)
			case *stufftext.Command:
				err = p.handleStufftext(c)
			case *print.Command:
				p.handlePrint(c)
			}
		}

		if err != nil {
			return nil, err
		}
	}

	return p.events, nil
}

func (p *Parser) handleModellist(cmd *modellist.Command) {
	if len(p.models) == 0 {
		p.models = append(p.models, "")
	}

	for _, m := range cmd.Models {
		p.models = append(p.models, m)
	}
}

func (p *Parser) handleSpawnBaseline(cmd *spawnbaseline.Command) {
	if cmd.Baseline == nil {
		return
	}

	p.entities[cmd.Index] = &entity{
		coord: cmd.Baseline.Coord,
		model: cmd.Baseline.ModelIndex,
		skin:  cmd.Baseline.SkinNum,
	}
}

func (p *Parser) handleFTESpawnBaseline(cmd *ftespawnbaseline.Command) {
	if cmd.Delta == nil {
		return
	}

	p.entities[cmd.Delta.Number] = &entity{
		coord: cmd.Delta.Coord,
		model: cmd.Delta.ModelIndex,
		skin:  cmd.Delta.Skin,
	}
}

func (p *Parser) handleUpdateUserinfo(cmd *updateuserinfo.Command) {
	is := infostring.Parse(cmd.UserInfo)
	name := ascii.Parse(is.Get("name"))
	if name == "" {
		return
	}

	p.players[cmd.PlayerIndex+1] = name
}

func (p *Parser) handleStufftext(cmd *stufftext.Command) error {
	for _, a := range args.Parse(cmd.String) {
		if len(a.Args) < 1 {
			continue
		}

		switch a.Cmd {
		case "//ktx":
			return p.handleKTX(a.Args)
		}
	}

	return nil
}

func (p *Parser) handleKTX(args []string) error {
	switch args[0] {
	case "matchstart":
		p.started = true
	case "took":
		if len(args) != 4 {
			break
		}

		entID, err := strconv.Atoi(args[1])
		if err != nil {
			return err
		}

		playerID, err := strconv.Atoi(args[3])
		if err != nil {
			return err
		}

		ent := p.entities[uint16(entID)]
		item := format.Model(p.models[ent.model], ent.skin)

		p.events = append(p.events, Event{
			Timestamp: p.elapsed,
			Name:      p.players[byte(playerID)],
			Type:      Took,
			Item:      item,
		})
	}

	return nil
}

func (p *Parser) handlePrint(c *print.Command) {
	s := strings.TrimRight(ascii.Parse(c.String), "\r\n")

	isMM2 := strings.HasPrefix(s, "(")
	now := time.Unix(0, int64(p.elapsed*1e9))
	if !isMM2 {
		ob, ok := death.Parse(s)
		if ok {
			if ob.Victim != "" && p.isPlayer(ob.Victim) {
				p.events = append(p.events, Event{
					Timestamp: p.elapsed,
					Name:      ob.Victim,
					Type:      Lost,
				})
			}
		}
	} else if isMM2 && p.filter.Allow(now, s) {
		if strings.Contains(s, "drop") || strings.Contains(s, "lost") {
			p.events = append(p.events, Event{
				Timestamp: p.elapsed,
				Name:      parseMM2Name(s),
				Type:      LostReport,
			})
		} else if strings.Contains(s, "took") || strings.Contains(s, "team") {
			p.events = append(p.events, Event{
				Timestamp: p.elapsed,
				Name:      parseMM2Name(s),
				Type:      TookReport,
			})
		}
	}

}

func parseMM2Name(input string) string {
	start := -1
	for i := 0; i < len(input); i++ {
		if input[i] == '(' {
			start = i
			break
		}
	}
	if start == -1 {
		return ""
	}

	depth := 0
	for i := start; i < len(input); i++ {
		switch input[i] {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return strings.TrimSpace(input[start+1 : i])
			}
		}
	}

	return ""
}

func (p *Parser) isPlayer(name string) bool {
	for _, n := range p.players {
		if n == name {
			return true
		}
	}

	return false
}
