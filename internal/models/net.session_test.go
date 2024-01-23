package models

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/ryanfaerman/netctl/internal/events"
)

type netSessionTestCase struct {
	input    NetSession
	events   EventStream
	expected NetSession
}

type netSessionTestSuite map[string]netSessionTestCase

func (suite netSessionTestSuite) run(t *testing.T) {
	for scenario, example := range suite {
		scenario := scenario
		example := example
		t.Run(scenario, func(t *testing.T) {
			t.Parallel()
			net := NewNet(1, "test")
			net.Sessions["abc"] = &example.input
			net.replay(example.events)
			actual := *net.Sessions["abc"]
			if diff := cmp.Diff(example.expected, actual); diff != "" {
				t.Errorf("Events did not apply: (-want +got)\n%s", diff)
			}
		})
	}
}

func TestNetSessionReplayScheduling(t *testing.T) {
	futureTime := time.Now().Add(time.Hour)
	nowTime := time.Now()

	suite := netSessionTestSuite{
		"open empty net": {
			input: NetSession{ID: "abc"},
			events: EventStream{
				{
					StreamID: "abc",
					At:       nowTime,
					Event:    events.NetSessionOpened{},
				},
			},
			expected: NetSession{ID: "abc", Periods: Periods{{OpenedAt: nowTime}}},
		},

		"open an open net": {
			input: NetSession{ID: "abc", Periods: Periods{
				{OpenedAt: nowTime},
			}},
			events: EventStream{
				{
					StreamID: "abc",
					At:       nowTime.Add(time.Minute),
					Event:    events.NetSessionOpened{},
				},
			},
			expected: NetSession{ID: "abc", Periods: Periods{
				{OpenedAt: nowTime},
			}},
		},

		"open a closed net": {
			input: NetSession{ID: "abc", Periods: Periods{
				{OpenedAt: nowTime, ClosedAt: nowTime.Add(time.Minute)},
			}},
			events: EventStream{
				{
					StreamID: "abc",
					At:       nowTime.Add(2 * time.Minute),
					Event:    events.NetSessionOpened{},
				},
			},
			expected: NetSession{ID: "abc", Periods: Periods{
				{OpenedAt: nowTime, ClosedAt: nowTime.Add(time.Minute)},
				{OpenedAt: nowTime.Add(2 * time.Minute)},
			}},
		},

		"open a scheduled net": {
			input: NetSession{ID: "abc", Periods: Periods{
				{OpenedAt: nowTime.Add(time.Minute), Scheduled: true},
			}},
			events: EventStream{
				{
					StreamID: "abc",
					At:       nowTime.Add(2 * time.Minute),
					Event:    events.NetSessionOpened{},
				},
			},
			expected: NetSession{ID: "abc", Periods: Periods{
				{OpenedAt: nowTime.Add(time.Minute), ClosedAt: nowTime.Add(2 * time.Minute), Scheduled: true},
				{OpenedAt: nowTime.Add(2 * time.Minute)},
			}},
		},

		"schedule new net": {
			input: NetSession{ID: "abc"},
			events: EventStream{
				{
					StreamID: "abc",
					Event:    events.NetSessionScheduled{At: futureTime},
				},
			},
			expected: NetSession{ID: "abc", Periods: Periods{
				{OpenedAt: futureTime, Scheduled: true},
			}},
		},

		"schedule onto open net": {
			input: NetSession{ID: "abc", Periods: Periods{
				{OpenedAt: nowTime},
			}},
			events: EventStream{
				{
					StreamID: "abc",
					Event:    events.NetSessionScheduled{At: futureTime},
				},
			},
			expected: NetSession{ID: "abc", Periods: Periods{
				{OpenedAt: nowTime},
			}},
		},

		"schedule onto closed net": {
			input: NetSession{ID: "abc", Periods: Periods{
				{OpenedAt: nowTime, ClosedAt: nowTime.Add(time.Minute)},
			}},
			events: EventStream{
				{
					StreamID: "abc",
					Event:    events.NetSessionScheduled{At: futureTime},
				},
			},
			expected: NetSession{ID: "abc", Periods: Periods{
				{OpenedAt: nowTime, ClosedAt: nowTime.Add(time.Minute)},
			}},
		},

		"close an unopened net": {
			input: NetSession{ID: "abc", Periods: Periods{}},
			events: EventStream{
				{
					StreamID: "abc",
					Event:    events.NetSessionClosed{},
				},
			},
			expected: NetSession{ID: "abc", Periods: Periods{}},
		},

		"close an open net": {
			input: NetSession{ID: "abc", Periods: Periods{
				{OpenedAt: nowTime},
			}},
			events: EventStream{
				{
					StreamID: "abc",
					At:       nowTime.Add(time.Minute),
					Event:    events.NetSessionClosed{},
				},
			},
			expected: NetSession{ID: "abc", Periods: Periods{
				{OpenedAt: nowTime, ClosedAt: nowTime.Add(time.Minute)},
			}},
		},

		"close a closed net": {
			input: NetSession{ID: "abc", Periods: Periods{
				{OpenedAt: nowTime, ClosedAt: nowTime.Add(time.Minute)},
			}},
			events: EventStream{
				{
					StreamID: "abc",
					At:       nowTime.Add(2 * time.Minute),
					Event:    events.NetSessionClosed{},
				},
			},
			expected: NetSession{ID: "abc", Periods: Periods{
				{OpenedAt: nowTime, ClosedAt: nowTime.Add(time.Minute)},
			}},
		},

		"close a scheduled net": {
			input: NetSession{ID: "abc", Periods: Periods{
				{OpenedAt: nowTime, Scheduled: true},
			}},
			events: EventStream{
				{
					StreamID: "abc",
					At:       nowTime.Add(time.Minute),
					Event:    events.NetSessionClosed{},
				},
			},
			expected: NetSession{ID: "abc", Periods: Periods{
				{OpenedAt: nowTime, ClosedAt: nowTime.Add(time.Minute), Scheduled: true},
			}},
		},
	}

	suite.run(t)
}
