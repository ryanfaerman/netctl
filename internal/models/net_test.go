package models

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/ryanfaerman/netctl/internal/events"
)

func TestNetReplay(t *testing.T) {
	futureTime := time.Now().Add(time.Hour)
	nowTime := time.Now()

	examples := []struct {
		scenario string
		input    NetSession
		events   EventStream
		expected NetSession
	}{
		{
			scenario: "schedule new net",
			input:    NetSession{ID: "abc"},
			events: EventStream{
				Event{
					StreamID: "abc",
					Event:    events.NetSessionScheduled{At: futureTime},
				},
			},
			expected: NetSession{ID: "abc", Periods: Periods{{OpenedAt: futureTime}}},
		},
		{
			scenario: "schedule onto open net",
			input:    NetSession{ID: "abc"},
			events: EventStream{
				Event{
					StreamID: "abc",
					At:       nowTime,
					Event:    events.NetSessionOpened{},
				},
				Event{
					StreamID: "abc",
					Event:    events.NetSessionScheduled{At: futureTime},
				},
			},
			expected: NetSession{ID: "abc", Periods: Periods{{OpenedAt: nowTime}}},
		},
	}

	for _, example := range examples {
		example := example
		t.Run(example.scenario, func(t *testing.T) {
			net := NewNet(1, example.scenario)
			net.Sessions["abc"] = &example.input

			net.replay(example.events)

			actual := *net.Sessions["abc"]

			// changelog, err := diff.Diff(actual, example.expected)
			// if err != nil {
			// 	t.Fatal(err)
			// }
			// if len(changelog) > 0 {
			// 	for _, change := range changelog {
			// 		t.Errorf(
			// 			"Events did not apply: \n\tValue:    %s\n\tExpected: %s\n\tActual:   %s",
			// 			strings.Join(change.Path, "."),
			// 			change.From, change.To,
			// 		)
			// 	}
			// }
			if diff := cmp.Diff(example.expected, actual); diff != "" {
				t.Errorf("Events did not apply: (-want +got)\n%s", diff)
			}
		})
	}
}
