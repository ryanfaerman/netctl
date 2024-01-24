package models

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/ryanfaerman/netctl/hamdb"
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
			if diff := cmp.Diff(example.expected, actual, cmpopts.EquateErrors()); diff != "" {
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

func TestNetCheckinReplay(t *testing.T) {
	// futureTime := time.Now().Add(time.Hour)
	nowTime := time.Now()

	suite := netSessionTestSuite{
		"checkins heard": {
			input: NetSession{ID: "abc", Checkins: []NetCheckin{
				{
					ID:       "checkin-123",
					At:       nowTime,
					Callsign: Hearable{AsHeard: "CLSGN-1"},
					Name:     Hearable{AsHeard: "NAME-1"},
					Location: Hearable{AsHeard: "LOC-1"},
					Kind:     NetCheckinKindRoutine,
					Traffic:  1,
					Acked:    true,
					Verified: true,
					Valid:    nil,
				},
			}},
			events: EventStream{
				{
					StreamID: "abc",
					At:       nowTime,
					Event: events.NetCheckinHeard{
						ID:       "checkin-123",
						Callsign: "CLSGN-1",
						Name:     "NAME-1",
						Location: "LOC-1",
						Kind:     "ROUTINE",
						Traffic:  1,
					},
				},
				{
					StreamID: "abc",
					At:       nowTime,
					Event: events.NetCheckinHeard{
						ID:       "checkin-789",
						Callsign: "CLSGN-1",
						Name:     "NAME-1",
						Location: "LOC-1",
						Kind:     "ROUTINE",
						Traffic:  1,
					},
				},
				{
					StreamID: "abc",
					At:       nowTime,
					Event: events.NetCheckinHeard{
						ID:       "checkin-456",
						Callsign: "CLSGN-2",
						Name:     "NAME-2",
						Location: "LOC-2",
						Kind:     "WELFARE",
						Traffic:  3,
					},
				},
			},
			expected: NetSession{ID: "abc", Checkins: []NetCheckin{
				{
					ID:       "checkin-123",
					At:       nowTime,
					Callsign: Hearable{AsHeard: "CLSGN-1"},
					Name:     Hearable{AsHeard: "NAME-1"},
					Location: Hearable{AsHeard: "LOC-1"},
					Kind:     NetCheckinKindRoutine,
					Traffic:  1,
					Acked:    false,
					Verified: false,
					Valid:    nil,
				},
				{
					ID:       "checkin-456",
					At:       nowTime,
					Callsign: Hearable{AsHeard: "CLSGN-2"},
					Name:     Hearable{AsHeard: "NAME-2"},
					Location: Hearable{AsHeard: "LOC-2"},
					Kind:     NetCheckinKindWelfare,
					Traffic:  3,
					Acked:    false,
					Verified: false,
					Valid:    nil,
				},
			}},
		},
		"checkins verified": {
			input: NetSession{ID: "abc", Checkins: []NetCheckin{
				{
					ID:       "checkin-123",
					At:       nowTime,
					Callsign: Hearable{AsHeard: "CLSGN-1"},
					Name:     Hearable{AsHeard: "NAME-1"},
					Location: Hearable{AsHeard: "LOC-1"},
					Verified: false,
					Valid:    nil,
				},
				{
					ID:       "checkin-456",
					At:       nowTime,
					Callsign: Hearable{AsHeard: "CLSGN-2"},
					Name:     Hearable{},
					Location: Hearable{AsHeard: "LOC-2"},
					Verified: false,
					Valid:    nil,
				},
			}},
			events: EventStream{
				{
					StreamID: "abc",
					At:       nowTime,
					Event: events.NetCheckinVerified{
						ID:       "checkin-123",
						Callsign: "CLSGN-1-VERIFIED",
						Name:     "NAME-1-VERIFIED",
						Location: "LOC-1-VERIFIED",
					},
				},
				{
					StreamID: "abc",
					At:       nowTime,
					Event: events.NetCheckinVerified{
						ID:        "checkin-456",
						ErrorType: "hamdb.ErrNotFound",
					},
				},
			},
			expected: NetSession{ID: "abc", Checkins: []NetCheckin{
				{
					ID:       "checkin-123",
					At:       nowTime,
					Callsign: Hearable{AsHeard: "CLSGN-1", AsLicensed: "CLSGN-1-VERIFIED"},
					Name:     Hearable{AsHeard: "NAME-1", AsLicensed: "NAME-1-VERIFIED"},
					Location: Hearable{AsHeard: "LOC-1", AsLicensed: "LOC-1-VERIFIED"},
					Verified: true,
					Valid:    nil,
				},
				{
					ID:       "checkin-456",
					At:       nowTime,
					Callsign: Hearable{AsHeard: "CLSGN-2"},
					Name:     Hearable{},
					Location: Hearable{AsHeard: "LOC-2"},
					Verified: true,
					Valid:    hamdb.ErrNotFound,
				},
			}},
		},
		"checkin acked": {
			input: NetSession{ID: "abc", Checkins: []NetCheckin{
				{
					ID:       "checkin-123",
					At:       nowTime,
					Callsign: Hearable{AsHeard: "CLSGN-1", AsLicensed: "CLSGN-1-VERIFIED"},
					Name:     Hearable{AsHeard: "NAME-1", AsLicensed: "NAME-1-VERIFIED"},
					Location: Hearable{AsHeard: "LOC-1", AsLicensed: "LOC-1-VERIFIED"},
					Verified: true,
					Valid:    nil,
				},
				{
					ID:       "checkin-456",
					At:       nowTime,
					Callsign: Hearable{AsHeard: "CLSGN-2"},
					Name:     Hearable{},
					Location: Hearable{AsHeard: "LOC-2"},
					Verified: true,
					Valid:    nil,
				},
			}},
			events: EventStream{
				{
					StreamID: "abc",
					At:       nowTime,
					Event: events.NetCheckinAcked{
						ID: "checkin-123",
					},
				},
			},
			expected: NetSession{ID: "abc", Checkins: []NetCheckin{
				{
					ID:       "checkin-123",
					At:       nowTime,
					Callsign: Hearable{AsHeard: "CLSGN-1", AsLicensed: "CLSGN-1-VERIFIED"},
					Name:     Hearable{AsHeard: "NAME-1", AsLicensed: "NAME-1-VERIFIED"},
					Location: Hearable{AsHeard: "LOC-1", AsLicensed: "LOC-1-VERIFIED"},
					Verified: true,
					Valid:    nil,
					Acked:    true,
				},
				{
					ID:       "checkin-456",
					At:       nowTime,
					Callsign: Hearable{AsHeard: "CLSGN-2"},
					Name:     Hearable{},
					Location: Hearable{AsHeard: "LOC-2"},
					Verified: true,
					Valid:    nil,
				},
			}},
		},
		"checkin corrected": {
			input: NetSession{ID: "abc", Checkins: []NetCheckin{
				{
					ID:       "checkin-123",
					At:       nowTime,
					Callsign: Hearable{AsHeard: "CLSGN-1", AsLicensed: "CLSGN-1-VERIFIED"},
					Name:     Hearable{AsHeard: "NAME-1", AsLicensed: "NAME-1-VERIFIED"},
					Location: Hearable{AsHeard: "LOC-1", AsLicensed: "LOC-1-VERIFIED"},
					Kind:     NetCheckinKindRoutine,
					Traffic:  7,
					Verified: true,
					Valid:    nil,
					Acked:    true,
				},
				{
					ID:       "checkin-456",
					At:       nowTime,
					Callsign: Hearable{AsHeard: "CLSGN-2"},
					Name:     Hearable{},
					Location: Hearable{AsHeard: "LOC-2"},
					Verified: true,
					Valid:    nil,
				},
			}},
			events: EventStream{
				{
					StreamID: "abc",
					At:       nowTime,
					Event: events.NetCheckinCorrected{
						ID:       "checkin-123",
						Callsign: "CLSGN-1-CORRECTED",
						Name:     "NAME-1-CORRECTED",
						Location: "LOC-1-CORRECTED",
						Kind:     "WELFARE",
						Traffic:  3,
					},
				},
			},
			expected: NetSession{ID: "abc", Checkins: []NetCheckin{
				{
					ID:       "checkin-123",
					At:       nowTime,
					Callsign: Hearable{AsHeard: "CLSGN-1-CORRECTED"},
					Name:     Hearable{AsHeard: "NAME-1-CORRECTED"},
					Location: Hearable{AsHeard: "LOC-1-CORRECTED"},
					Kind:     NetCheckinKindWelfare,
					Traffic:  3,
					Verified: false,
					Valid:    nil,
					Acked:    true,
				},
				{
					ID:       "checkin-456",
					At:       nowTime,
					Callsign: Hearable{AsHeard: "CLSGN-2"},
					Name:     Hearable{},
					Location: Hearable{AsHeard: "LOC-2"},
					Verified: true,
					Valid:    nil,
				},
			}},
		},
		"checkin revoked": {
			input: NetSession{ID: "abc", Checkins: []NetCheckin{
				{
					ID:       "checkin-123",
					At:       nowTime,
					Callsign: Hearable{AsHeard: "CLSGN-1", AsLicensed: "CLSGN-1-VERIFIED"},
					Name:     Hearable{AsHeard: "NAME-1", AsLicensed: "NAME-1-VERIFIED"},
					Location: Hearable{AsHeard: "LOC-1", AsLicensed: "LOC-1-VERIFIED"},
					Kind:     NetCheckinKindRoutine,
					Traffic:  7,
					Verified: true,
					Valid:    nil,
					Acked:    true,
				},
				{
					ID:       "checkin-456",
					At:       nowTime,
					Callsign: Hearable{AsHeard: "CLSGN-2"},
					Name:     Hearable{},
					Location: Hearable{AsHeard: "LOC-2"},
					Verified: true,
					Valid:    nil,
				},
			}},
			events: EventStream{
				{
					StreamID: "abc",
					At:       nowTime,
					Event: events.NetCheckinRevoked{
						ID: "checkin-123",
					},
				},
			},
			expected: NetSession{ID: "abc", Checkins: []NetCheckin{
				{
					ID:       "checkin-456",
					At:       nowTime,
					Callsign: Hearable{AsHeard: "CLSGN-2"},
					Name:     Hearable{},
					Location: Hearable{AsHeard: "LOC-2"},
					Verified: true,
					Valid:    nil,
				},
			}},
		},
	}
	suite.run(t)
}
