package conf

import (
	"reflect"
	"testing"
)

func TestNormalizeReplacesOutOfRangeValues(t *testing.T) {
	testCases := []struct {
		name         string
		wait         int
		waitSlow     int
		historyLimit int
		closeMatches int
		wantWait     int
		wantSlow     int
		wantHistory  int
		wantClose    int
		wantRecorded []string
	}{
		{
			name: "valid",
			wait: 1, waitSlow: 30, historyLimit: 100, closeMatches: 5,
			wantWait: 1, wantSlow: 30, wantHistory: 100, wantClose: 5,
		},
		{
			name: "zero wait",
			wait: 0, waitSlow: 15, closeMatches: 3,
			wantWait: 3, wantSlow: 15, wantClose: 3,
			wantRecorded: []string{"wait_command 0"},
		},
		{
			name: "negative wait",
			wait: -1, waitSlow: 15, closeMatches: 3,
			wantWait: 3, wantSlow: 15, wantClose: 3,
			wantRecorded: []string{"wait_command -1"},
		},
		{
			name: "wait overflow",
			wait: 10000000000, waitSlow: 15, closeMatches: 3,
			wantWait: 3, wantSlow: 15, wantClose: 3,
			wantRecorded: []string{"wait_command 10000000000"},
		},
		{
			name: "zero slow wait",
			wait: 3, waitSlow: 0, closeMatches: 3,
			wantWait: 3, wantSlow: 15, wantClose: 3,
			wantRecorded: []string{"wait_slow_command 0"},
		},
		{
			name: "max wait",
			wait: 3600, waitSlow: 3600, closeMatches: 3,
			wantWait: 3600, wantSlow: 3600, wantClose: 3,
		},
		{
			name: "negative history limit",
			wait: 3, waitSlow: 15, historyLimit: -5, closeMatches: 3,
			wantWait: 3, wantSlow: 15, wantHistory: 0, wantClose: 3,
			wantRecorded: []string{"history_limit -5"},
		},
		{
			name: "negative close matches",
			wait: 3, waitSlow: 15, closeMatches: -2,
			wantWait: 3, wantSlow: 15, wantClose: 3,
			wantRecorded: []string{"num_close_matches -2"},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			settings := Default()
			settings.WaitCommand = testCase.wait
			settings.WaitSlowCommand = testCase.waitSlow
			settings.HistoryLimit = testCase.historyLimit
			settings.NumCloseMatches = testCase.closeMatches

			settings.normalize()

			if settings.WaitCommand != testCase.wantWait {
				t.Errorf("WaitCommand = %d, want %d", settings.WaitCommand, testCase.wantWait)
			}
			if settings.WaitSlowCommand != testCase.wantSlow {
				t.Errorf("WaitSlowCommand = %d, want %d", settings.WaitSlowCommand, testCase.wantSlow)
			}
			if settings.HistoryLimit != testCase.wantHistory {
				t.Errorf("HistoryLimit = %d, want %d", settings.HistoryLimit, testCase.wantHistory)
			}
			if settings.NumCloseMatches != testCase.wantClose {
				t.Errorf("NumCloseMatches = %d, want %d", settings.NumCloseMatches, testCase.wantClose)
			}
			if !reflect.DeepEqual(settings.InvalidValues, testCase.wantRecorded) {
				t.Errorf("InvalidValues = %q, want %q", settings.InvalidValues, testCase.wantRecorded)
			}
		})
	}
}

func TestLoadRecordsInvalidEnvValues(t *testing.T) {
	isolateConfig(t)
	t.Setenv("GOFUCK_WAIT_COMMAND", "0")
	t.Setenv("THEFUCK_WAIT_SLOW_COMMAND", "-3")

	settings, err := Load(&Args{})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if settings.WaitCommand != 3 || settings.WaitSlowCommand != 15 {
		t.Fatalf("WaitCommand = %d, WaitSlowCommand = %d, want the defaults", settings.WaitCommand, settings.WaitSlowCommand)
	}
	if !reflect.DeepEqual(settings.InvalidValues,
		[]string{"wait_command 0", "wait_slow_command -3"}) {
		t.Fatalf("InvalidValues = %q", settings.InvalidValues)
	}
}
