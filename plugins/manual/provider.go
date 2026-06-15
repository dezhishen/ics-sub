package manual

import (
	"time"

	"ics-sub/internal/subscriptions"
)

type provider struct{}

func (provider) Disabled() bool {
	return false
}

func (provider) Name() string {
	return "manual"
}

func (provider) Generate() ([]subscriptions.Calendar, error) {
	now := time.Now().UTC()
	return []subscriptions.Calendar{
		{
			ID:          "cn-holiday",
			Name:        "中国节假日",
			Group:       "holiday",
			GroupName:   "节假日",
			Description: "内置示例数据，可替换为真实节假日源。",
			UpdatedAt:   now,
			Events: []subscriptions.Event{
				{
					UID:         "cn-holiday-2026-new-year",
					Summary:     "元旦",
					Description: "法定节假日示例事件",
					StartAt:     time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
					EndAt:       time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
					AllDay:      true,
				},
			},
		},
		{
			ID:          "release-schedule",
			Name:        "版本发布计划",
			Group:       "work",
			GroupName:   "工作安排",
			Description: "示例发布节奏，可由插件动态维护。",
			UpdatedAt:   now,
			Events: []subscriptions.Event{
				{
					UID:         "release-plan-2026-q3",
					Summary:     "Q3 版本发布",
					Description: "示例发布窗口",
					StartAt:     time.Date(2026, 7, 15, 2, 0, 0, 0, time.UTC),
					EndAt:       time.Date(2026, 7, 15, 3, 0, 0, 0, time.UTC),
				},
			},
		},
	}, nil
}

func init() {
	subscriptions.Register(provider{})
}
