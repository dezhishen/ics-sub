package lpl

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"ics-sub/internal/subscriptions"
)

const (
	gameListURL   = "https://apps.game.qq.com/lol/match/apis/searchMatchGameInfo.php?r1=gamelist&p1=5"
	matchListTmpl = "https://lpl.qq.com/web201612/data/LOL_MATCH2_MATCH_HOMEPAGE_BMATCH_LIST_%s.js"
)

type provider struct{}

type seasonListResponse struct {
	Status string       `json:"status"`
	Msg    []seasonItem `json:"msg"`
}

type seasonItem struct {
	GameID   string `json:"GameId"`
	GameName string `json:"GameName"`
	GameYear string `json:"GameYear"`
}

type matchListResponse struct {
	Status string      `json:"status"`
	Msg    []matchItem `json:"msg"`
}

type matchItem struct {
	BMatchID       string `json:"bMatchId"`
	BMatchName     string `json:"bMatchName"`
	TeamShortNameA string `json:"TeamShortNameA"`
	TeamShortNameB string `json:"TeamShortNameB"`
	GameModeName   string `json:"GameModeName"`
	GameTypeName   string `json:"GameTypeName"`
	GameProcName   string `json:"GameProcName"`
	MatchDate      string `json:"MatchDate"`
	GamePlaceName  string `json:"GamePlaceName"`
	MatchStatus    string `json:"MatchStatus"`
	ScoreA         string `json:"ScoreA"`
	ScoreB         string `json:"ScoreB"`
}

func (provider) Name() string {
	return "lpl"
}
func (provider) Disabled() bool {
	return false
}

func (provider) Generate() ([]subscriptions.Calendar, error) {
	client := &http.Client{Timeout: 20 * time.Second}

	season, err := fetchLatestSeason(client)
	if err != nil {
		return nil, err
	}

	matches, err := fetchMatches(client, season.GameID)
	if err != nil {
		return nil, err
	}

	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return nil, fmt.Errorf("load timezone: %w", err)
	}

	events := make([]subscriptions.Event, 0, len(matches))
	seen := map[string]struct{}{}
	for _, m := range matches {
		if strings.TrimSpace(m.MatchDate) == "" {
			continue
		}
		startAt, err := time.ParseInLocation("2006-01-02 15:04:05", m.MatchDate, loc)
		if err != nil {
			continue
		}

		uid := strings.TrimSpace(m.BMatchID)
		if uid == "" {
			uid = fmt.Sprintf("%d", startAt.Unix())
		}
		uid = "lpl-" + uid
		if _, exists := seen[uid]; exists {
			continue
		}
		seen[uid] = struct{}{}

		summary := strings.TrimSpace(m.BMatchName)
		if summary == "" {
			summary = strings.TrimSpace(m.TeamShortNameA) + " vs " + strings.TrimSpace(m.TeamShortNameB)
		}
		if summary == "vs" {
			summary = "LPL 比赛"
		}

		desc := strings.TrimSpace(m.GameTypeName)
		if stage := strings.TrimSpace(m.GameProcName); stage != "" {
			if desc != "" {
				desc += " / "
			}
			desc += stage
		}
		if mode := strings.TrimSpace(m.GameModeName); mode != "" {
			if desc != "" {
				desc += " / "
			}
			desc += mode
		}
		if (m.ScoreA != "" || m.ScoreB != "") && m.MatchStatus != "1" {
			if desc != "" {
				desc += " / "
			}
			desc += fmt.Sprintf("比分 %s:%s", strings.TrimSpace(m.ScoreA), strings.TrimSpace(m.ScoreB))
		}

		events = append(events, subscriptions.Event{
			UID:         uid,
			Summary:     summary,
			Description: desc,
			Location:    strings.TrimSpace(m.GamePlaceName),
			StartAt:     startAt,
			EndAt:       startAt.Add(estimateDuration(m.GameModeName)),
		})
	}

	sort.Slice(events, func(i, j int) bool {
		return events[i].StartAt.Before(events[j].StartAt)
	})

	now := time.Now().UTC()
	return []subscriptions.Calendar{
		{
			ID:          "lpl-schedule",
			Name:        season.GameName,
			Group:       "esports",
			GroupName:   "电竞赛事",
			Description: "LPL 官方赛程，来源 lpl.qq.com。",
			UpdatedAt:   now,
			Events:      events,
		},
	}, nil
}

func fetchLatestSeason(client *http.Client) (seasonItem, error) {
	var resp seasonListResponse
	if err := fetchJSON(client, gameListURL, &resp); err != nil {
		return seasonItem{}, fmt.Errorf("fetch lpl season list: %w", err)
	}
	if resp.Status != "0" || len(resp.Msg) == 0 {
		return seasonItem{}, fmt.Errorf("lpl season list is empty")
	}

	sort.Slice(resp.Msg, func(i, j int) bool {
		yi, _ := strconv.Atoi(resp.Msg[i].GameYear)
		yj, _ := strconv.Atoi(resp.Msg[j].GameYear)
		if yi != yj {
			return yi > yj
		}
		idi, _ := strconv.Atoi(resp.Msg[i].GameID)
		idj, _ := strconv.Atoi(resp.Msg[j].GameID)
		return idi > idj
	})

	return resp.Msg[0], nil
}

func fetchMatches(client *http.Client, seasonID string) ([]matchItem, error) {
	url := fmt.Sprintf(matchListTmpl, seasonID)
	var resp matchListResponse
	if err := fetchJSON(client, url, &resp); err != nil {
		return nil, fmt.Errorf("fetch lpl matches: %w", err)
	}
	if resp.Status != "0" {
		return nil, fmt.Errorf("lpl matches response status: %s", resp.Status)
	}
	return resp.Msg, nil
}

func fetchJSON(client *http.Client, url string, out any) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "ics-sub/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("http %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	payload := extractJSONObject(body)
	if len(payload) == 0 {
		return fmt.Errorf("empty response payload")
	}

	if err := json.Unmarshal(payload, out); err != nil {
		return err
	}
	return nil
}

func extractJSONObject(raw []byte) []byte {
	s := strings.TrimSpace(string(raw))
	start := strings.IndexByte(s, '{')
	end := strings.LastIndexByte(s, '}')
	if start < 0 || end <= start {
		return nil
	}
	return []byte(s[start : end+1])
}

func estimateDuration(mode string) time.Duration {
	m := strings.ToUpper(strings.TrimSpace(mode))
	switch m {
	case "BO1":
		return 90 * time.Minute
	case "BO3":
		return 3 * time.Hour
	case "BO5":
		return 4 * time.Hour
	default:
		return 3 * time.Hour
	}
}

func init() {
	subscriptions.Register(provider{})
}
