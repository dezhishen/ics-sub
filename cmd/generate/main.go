package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"ics-sub/internal/subscriptions"
	_ "ics-sub/plugins/manual"
)

type calendarIndexItem struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Provider    string    `json:"provider"`
	UpdatedAt   time.Time `json:"updatedAt"`
	ICSPath     string    `json:"icsPath"`
	EventCount  int       `json:"eventCount"`
}

type groupPayload struct {
	Key       string              `json:"key"`
	Name      string              `json:"name"`
	Calendars []calendarIndexItem `json:"calendars"`
}

type payload struct {
	GeneratedAt time.Time      `json:"generatedAt"`
	Groups      []groupPayload `json:"groups"`
}

func main() {
	outPath := flag.String("out", "web/public/data/subscriptions.json", "output json path")
	icsDir := flag.String("ics-dir", "web/public/ics", "output directory for generated .ics files")
	flag.Parse()

	calendars, err := subscriptions.GenerateAll()
	if err != nil {
		panic(err)
	}
	if err := os.MkdirAll(*icsDir, 0o755); err != nil {
		panic(err)
	}

	groups := map[string]*groupPayload{}
	for _, c := range calendars {
		if strings.TrimSpace(c.ID) == "" {
			panic("calendar id cannot be empty")
		}
		if strings.TrimSpace(c.Name) == "" {
			panic(fmt.Sprintf("calendar name cannot be empty: %s", c.ID))
		}
		if strings.TrimSpace(c.Group) == "" {
			panic(fmt.Sprintf("calendar group cannot be empty: %s", c.ID))
		}

		icsFilename := c.ID + ".ics"
		icsPath := filepath.Join(*icsDir, icsFilename)
		if err := os.WriteFile(icsPath, buildICS(c), 0o644); err != nil {
			panic(err)
		}

		g, ok := groups[c.Group]
		if !ok {
			name := c.GroupName
			if name == "" {
				name = c.Group
			}
			g = &groupPayload{Key: c.Group, Name: name}
			groups[c.Group] = g
		}

		g.Calendars = append(g.Calendars, calendarIndexItem{
			ID:          c.ID,
			Name:        c.Name,
			Description: c.Description,
			Provider:    c.Provider,
			UpdatedAt:   c.UpdatedAt,
			ICSPath:     "ics/" + icsFilename,
			EventCount:  len(c.Events),
		})
	}

	groupKeys := make([]string, 0, len(groups))
	for key := range groups {
		groupKeys = append(groupKeys, key)
	}
	sort.Strings(groupKeys)

	outGroups := make([]groupPayload, 0, len(groupKeys))
	for _, key := range groupKeys {
		g := groups[key]
		sort.Slice(g.Calendars, func(i, j int) bool {
			return g.Calendars[i].Name < g.Calendars[j].Name
		})
		outGroups = append(outGroups, *g)
	}

	result := payload{
		GeneratedAt: time.Now().UTC(),
		Groups:      outGroups,
	}

	if err := os.MkdirAll(filepath.Dir(*outPath), 0o755); err != nil {
		panic(err)
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		panic(err)
	}

	if err := os.WriteFile(*outPath, append(data, '\n'), 0o644); err != nil {
		panic(err)
	}

	fmt.Printf("generated %d calendars -> %s and %s\n", len(calendars), *outPath, *icsDir)
}

func buildICS(c subscriptions.Calendar) []byte {
	now := time.Now().UTC().Format("20060102T150405Z")
	b := &bytes.Buffer{}
	writeLine(b, "BEGIN:VCALENDAR")
	writeLine(b, "VERSION:2.0")
	writeLine(b, "PRODID:-//ics-sub//calendar//CN")
	writeLine(b, "CALSCALE:GREGORIAN")
	writeLine(b, "X-WR-CALNAME:"+escapeText(c.Name))
	if c.Description != "" {
		writeLine(b, "X-WR-CALDESC:"+escapeText(c.Description))
	}

	for i, e := range c.Events {
		uid := e.UID
		if uid == "" {
			uid = fmt.Sprintf("%s-%d@ics-sub", c.ID, i+1)
		}
		writeLine(b, "BEGIN:VEVENT")
		writeLine(b, "UID:"+escapeText(uid))
		writeLine(b, "DTSTAMP:"+now)
		if e.AllDay {
			writeLine(b, "DTSTART;VALUE=DATE:"+e.StartAt.UTC().Format("20060102"))
			writeLine(b, "DTEND;VALUE=DATE:"+e.EndAt.UTC().Format("20060102"))
		} else {
			writeLine(b, "DTSTART:"+e.StartAt.UTC().Format("20060102T150405Z"))
			writeLine(b, "DTEND:"+e.EndAt.UTC().Format("20060102T150405Z"))
		}
		writeLine(b, "SUMMARY:"+escapeText(e.Summary))
		if e.Description != "" {
			writeLine(b, "DESCRIPTION:"+escapeText(e.Description))
		}
		if e.Location != "" {
			writeLine(b, "LOCATION:"+escapeText(e.Location))
		}
		writeLine(b, "END:VEVENT")
	}

	writeLine(b, "END:VCALENDAR")
	return b.Bytes()
}

func writeLine(b *bytes.Buffer, line string) {
	b.WriteString(line)
	b.WriteString("\r\n")
}

func escapeText(value string) string {
	v := strings.ReplaceAll(value, "\\", "\\\\")
	v = strings.ReplaceAll(v, ";", "\\;")
	v = strings.ReplaceAll(v, ",", "\\,")
	v = strings.ReplaceAll(v, "\n", "\\n")
	return v
}
