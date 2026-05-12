package rafay

import (
	"testing"

	"github.com/RafaySystems/rafay-common/proto/types/hub/infrapb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Expand tests
// ---------------------------------------------------------------------------

func TestExpandToV3GkeReleaseChannel(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		expected *infrapb.GkeReleaseChannel
		wantErr  bool
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: nil,
		},
		{
			name:     "empty input",
			input:    []interface{}{},
			expected: nil,
		},
		{
			name:     "nil element",
			input:    []interface{}{nil},
			expected: nil,
		},
		{
			name: "STABLE channel",
			input: []interface{}{
				map[string]interface{}{
					"channel": "STABLE",
				},
			},
			expected: &infrapb.GkeReleaseChannel{Channel: "STABLE"},
		},
		{
			name: "REGULAR channel",
			input: []interface{}{
				map[string]interface{}{
					"channel": "REGULAR",
				},
			},
			expected: &infrapb.GkeReleaseChannel{Channel: "REGULAR"},
		},
		{
			name: "RAPID channel",
			input: []interface{}{
				map[string]interface{}{
					"channel": "RAPID",
				},
			},
			expected: &infrapb.GkeReleaseChannel{Channel: "RAPID"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := expandToV3GkeReleaseChannel(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				assert.Equal(t, tt.expected.Channel, result.Channel)
			}
		})
	}
}

func TestExpandToV3GkeMaintenancePolicy(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		expected *infrapb.GkeMaintenancePolicy
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: nil,
		},
		{
			name:     "empty input",
			input:    []interface{}{},
			expected: nil,
		},
		{
			name: "daily window only",
			input: []interface{}{
				map[string]interface{}{
					"daily_maintenance_window": []interface{}{
						map[string]interface{}{
							"start_time": "03:00",
						},
					},
				},
			},
			expected: &infrapb.GkeMaintenancePolicy{
				DailyMaintenanceWindow: &infrapb.GkeDailyMaintenanceWindow{StartTime: "03:00"},
			},
		},
		{
			name: "recurring window only",
			input: []interface{}{
				map[string]interface{}{
					"recurring_window": []interface{}{
						map[string]interface{}{
							"start_time": "2026-01-03T09:00:00Z",
							"end_time":   "2026-01-03T17:00:00Z",
							"recurrence": "FREQ=WEEKLY;BYDAY=SA,SU",
						},
					},
				},
			},
			expected: &infrapb.GkeMaintenancePolicy{
				RecurringWindow: &infrapb.GkeRecurringWindow{
					StartTime:  "2026-01-03T09:00:00Z",
					EndTime:    "2026-01-03T17:00:00Z",
					Recurrence: "FREQ=WEEKLY;BYDAY=SA,SU",
				},
			},
		},
		{
			name: "exclusions only",
			input: []interface{}{
				map[string]interface{}{
					"maintenance_exclusions": []interface{}{
						map[string]interface{}{
							"name":       "holidays",
							"start_time": "2026-12-24T00:00:00Z",
							"end_time":   "2026-12-26T00:00:00Z",
						},
					},
				},
			},
			expected: &infrapb.GkeMaintenancePolicy{
				MaintenanceExclusions: []*infrapb.GkeMaintenanceExclusion{
					{
						Name:      "holidays",
						StartTime: "2026-12-24T00:00:00Z",
						EndTime:   "2026-12-26T00:00:00Z",
					},
				},
			},
		},
		{
			name: "recurring window with exclusions and options",
			input: []interface{}{
				map[string]interface{}{
					"recurring_window": []interface{}{
						map[string]interface{}{
							"start_time": "2026-01-03T09:00:00Z",
							"end_time":   "2026-01-03T17:00:00Z",
							"recurrence": "FREQ=WEEKLY;BYDAY=SA,SU",
						},
					},
					"maintenance_exclusions": []interface{}{
						map[string]interface{}{
							"name":       "deploy-freeze",
							"start_time": "2026-12-20T00:00:00Z",
							"end_time":   "2027-01-02T00:00:00Z",
							"exclusion_options": []interface{}{
								map[string]interface{}{
									"scope": "NO_UPGRADES",
								},
							},
						},
						map[string]interface{}{
							"name":       "q1-freeze",
							"start_time": "2027-03-01T00:00:00Z",
							"end_time":   "2027-03-15T00:00:00Z",
						},
					},
				},
			},
			expected: &infrapb.GkeMaintenancePolicy{
				RecurringWindow: &infrapb.GkeRecurringWindow{
					StartTime:  "2026-01-03T09:00:00Z",
					EndTime:    "2026-01-03T17:00:00Z",
					Recurrence: "FREQ=WEEKLY;BYDAY=SA,SU",
				},
				MaintenanceExclusions: []*infrapb.GkeMaintenanceExclusion{
					{
						Name:             "deploy-freeze",
						StartTime:        "2026-12-20T00:00:00Z",
						EndTime:          "2027-01-02T00:00:00Z",
						ExclusionOptions: &infrapb.GkeMaintenanceExclusionOptions{Scope: "NO_UPGRADES"},
					},
					{
						Name:      "q1-freeze",
						StartTime: "2027-03-01T00:00:00Z",
						EndTime:   "2027-03-15T00:00:00Z",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := expandToV3GkeMaintenancePolicy(tt.input)
			assert.NoError(t, err)
			if tt.expected == nil {
				assert.Nil(t, result)
				return
			}
			require.NotNil(t, result)

			// daily window
			if tt.expected.DailyMaintenanceWindow != nil {
				require.NotNil(t, result.DailyMaintenanceWindow)
				assert.Equal(t, tt.expected.DailyMaintenanceWindow.StartTime, result.DailyMaintenanceWindow.StartTime)
			} else {
				assert.Nil(t, result.DailyMaintenanceWindow)
			}

			// recurring window
			if tt.expected.RecurringWindow != nil {
				require.NotNil(t, result.RecurringWindow)
				assert.Equal(t, tt.expected.RecurringWindow.StartTime, result.RecurringWindow.StartTime)
				assert.Equal(t, tt.expected.RecurringWindow.EndTime, result.RecurringWindow.EndTime)
				assert.Equal(t, tt.expected.RecurringWindow.Recurrence, result.RecurringWindow.Recurrence)
			} else {
				assert.Nil(t, result.RecurringWindow)
			}

			// exclusions
			require.Len(t, result.MaintenanceExclusions, len(tt.expected.MaintenanceExclusions))
			for i, exp := range tt.expected.MaintenanceExclusions {
				got := result.MaintenanceExclusions[i]
				assert.Equal(t, exp.Name, got.Name)
				assert.Equal(t, exp.StartTime, got.StartTime)
				assert.Equal(t, exp.EndTime, got.EndTime)
				if exp.ExclusionOptions != nil {
					require.NotNil(t, got.ExclusionOptions)
					assert.Equal(t, exp.ExclusionOptions.Scope, got.ExclusionOptions.Scope)
				} else {
					assert.Nil(t, got.ExclusionOptions)
				}
			}
		})
	}
}

func TestExpandToV3GkeDailyMaintenanceWindow(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		expected *infrapb.GkeDailyMaintenanceWindow
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: nil,
		},
		{
			name: "valid start_time",
			input: []interface{}{
				map[string]interface{}{
					"start_time": "05:00",
				},
			},
			expected: &infrapb.GkeDailyMaintenanceWindow{StartTime: "05:00"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := expandToV3GkeDailyMaintenanceWindow(tt.input)
			assert.NoError(t, err)
			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				assert.Equal(t, tt.expected.StartTime, result.StartTime)
			}
		})
	}
}

func TestExpandToV3GkeRecurringWindow(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		expected *infrapb.GkeRecurringWindow
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: nil,
		},
		{
			name: "full fields",
			input: []interface{}{
				map[string]interface{}{
					"start_time": "2026-01-01T00:00:00Z",
					"end_time":   "2026-01-01T08:00:00Z",
					"recurrence": "FREQ=DAILY",
				},
			},
			expected: &infrapb.GkeRecurringWindow{
				StartTime:  "2026-01-01T00:00:00Z",
				EndTime:    "2026-01-01T08:00:00Z",
				Recurrence: "FREQ=DAILY",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := expandToV3GkeRecurringWindow(tt.input)
			assert.NoError(t, err)
			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				assert.Equal(t, tt.expected.StartTime, result.StartTime)
				assert.Equal(t, tt.expected.EndTime, result.EndTime)
				assert.Equal(t, tt.expected.Recurrence, result.Recurrence)
			}
		})
	}
}

func TestExpandToV3GkeMaintenanceExclusions(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		expected []*infrapb.GkeMaintenanceExclusion
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: nil,
		},
		{
			name: "single exclusion without options",
			input: []interface{}{
				map[string]interface{}{
					"name":       "maintenance-break",
					"start_time": "2026-06-01T00:00:00Z",
					"end_time":   "2026-06-02T00:00:00Z",
				},
			},
			expected: []*infrapb.GkeMaintenanceExclusion{
				{
					Name:      "maintenance-break",
					StartTime: "2026-06-01T00:00:00Z",
					EndTime:   "2026-06-02T00:00:00Z",
				},
			},
		},
		{
			name: "multiple exclusions with options",
			input: []interface{}{
				map[string]interface{}{
					"name":       "excl-1",
					"start_time": "2026-06-01T00:00:00Z",
					"end_time":   "2026-06-02T00:00:00Z",
					"exclusion_options": []interface{}{
						map[string]interface{}{
							"scope": "NO_MINOR_UPGRADES",
						},
					},
				},
				map[string]interface{}{
					"name":       "excl-2",
					"start_time": "2026-07-01T00:00:00Z",
					"end_time":   "2026-07-05T00:00:00Z",
				},
			},
			expected: []*infrapb.GkeMaintenanceExclusion{
				{
					Name:             "excl-1",
					StartTime:        "2026-06-01T00:00:00Z",
					EndTime:          "2026-06-02T00:00:00Z",
					ExclusionOptions: &infrapb.GkeMaintenanceExclusionOptions{Scope: "NO_MINOR_UPGRADES"},
				},
				{
					Name:      "excl-2",
					StartTime: "2026-07-01T00:00:00Z",
					EndTime:   "2026-07-05T00:00:00Z",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := expandToV3GkeMaintenanceExclusions(tt.input)
			assert.NoError(t, err)
			if tt.expected == nil {
				assert.Nil(t, result)
				return
			}
			require.Len(t, result, len(tt.expected))
			for i, exp := range tt.expected {
				got := result[i]
				assert.Equal(t, exp.Name, got.Name)
				assert.Equal(t, exp.StartTime, got.StartTime)
				assert.Equal(t, exp.EndTime, got.EndTime)
				if exp.ExclusionOptions != nil {
					require.NotNil(t, got.ExclusionOptions)
					assert.Equal(t, exp.ExclusionOptions.Scope, got.ExclusionOptions.Scope)
				} else {
					assert.Nil(t, got.ExclusionOptions)
				}
			}
		})
	}
}

func TestExpandToV3GkeMaintenanceExclusionOptions(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		expected *infrapb.GkeMaintenanceExclusionOptions
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: nil,
		},
		{
			name: "valid scope",
			input: []interface{}{
				map[string]interface{}{
					"scope": "NO_UPGRADES",
				},
			},
			expected: &infrapb.GkeMaintenanceExclusionOptions{Scope: "NO_UPGRADES"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := expandToV3GkeMaintenanceExclusionOptions(tt.input)
			assert.NoError(t, err)
			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				assert.Equal(t, tt.expected.Scope, result.Scope)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Validation tests
// ---------------------------------------------------------------------------

func TestValidateAutoUpgradeWithReleaseChannel(t *testing.T) {
	tests := []struct {
		name    string
		gke     *infrapb.GkeV3ConfigObject
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil gke config",
			gke:     nil,
			wantErr: false,
		},
		{
			name:    "no release channel",
			gke:     &infrapb.GkeV3ConfigObject{},
			wantErr: false,
		},
		{
			name: "release channel with empty channel string",
			gke: &infrapb.GkeV3ConfigObject{
				ReleaseChannel: &infrapb.GkeReleaseChannel{Channel: ""},
			},
			wantErr: false,
		},
		{
			name: "UNSPECIFIED channel skips validation",
			gke: &infrapb.GkeV3ConfigObject{
				ReleaseChannel: &infrapb.GkeReleaseChannel{Channel: "UNSPECIFIED"},
				NodePools: []*infrapb.GkeNodePool{
					{Name: "pool-1"},
				},
			},
			wantErr: false,
		},
		{
			name: "REGULAR channel with auto_upgrade enabled",
			gke: &infrapb.GkeV3ConfigObject{
				ReleaseChannel: &infrapb.GkeReleaseChannel{Channel: "REGULAR"},
				NodePools: []*infrapb.GkeNodePool{
					{Name: "pool-1", Management: &infrapb.GkeNodeManagement{AutoUpgrade: true}},
				},
			},
			wantErr: false,
		},
		{
			name: "RAPID channel with auto_upgrade enabled on all pools",
			gke: &infrapb.GkeV3ConfigObject{
				ReleaseChannel: &infrapb.GkeReleaseChannel{Channel: "RAPID"},
				NodePools: []*infrapb.GkeNodePool{
					{Name: "pool-1", Management: &infrapb.GkeNodeManagement{AutoUpgrade: true}},
					{Name: "pool-2", Management: &infrapb.GkeNodeManagement{AutoUpgrade: true}},
				},
			},
			wantErr: false,
		},
		{
			name: "STABLE channel with auto_upgrade disabled",
			gke: &infrapb.GkeV3ConfigObject{
				ReleaseChannel: &infrapb.GkeReleaseChannel{Channel: "STABLE"},
				NodePools: []*infrapb.GkeNodePool{
					{Name: "pool-1", Management: &infrapb.GkeNodeManagement{AutoUpgrade: false}},
				},
			},
			wantErr: true,
			errMsg:  `node pool "pool-1" must have management.auto_upgrade enabled when release channel "STABLE" is set`,
		},
		{
			name: "EXTENDED channel with nil management",
			gke: &infrapb.GkeV3ConfigObject{
				ReleaseChannel: &infrapb.GkeReleaseChannel{Channel: "EXTENDED"},
				NodePools: []*infrapb.GkeNodePool{
					{Name: "pool-no-mgmt"},
				},
			},
			wantErr: true,
			errMsg:  `node pool "pool-no-mgmt" must have management.auto_upgrade enabled when release channel "EXTENDED" is set`,
		},
		{
			name: "REGULAR channel mixed pools - second pool missing auto_upgrade",
			gke: &infrapb.GkeV3ConfigObject{
				ReleaseChannel: &infrapb.GkeReleaseChannel{Channel: "REGULAR"},
				NodePools: []*infrapb.GkeNodePool{
					{Name: "pool-ok", Management: &infrapb.GkeNodeManagement{AutoUpgrade: true}},
					{Name: "pool-bad", Management: &infrapb.GkeNodeManagement{AutoUpgrade: false}},
				},
			},
			wantErr: true,
			errMsg:  `node pool "pool-bad" must have management.auto_upgrade enabled when release channel "REGULAR" is set`,
		},
		{
			name: "case-insensitive channel check",
			gke: &infrapb.GkeV3ConfigObject{
				ReleaseChannel: &infrapb.GkeReleaseChannel{Channel: "Stable"},
				NodePools: []*infrapb.GkeNodePool{
					{Name: "pool-1", Management: &infrapb.GkeNodeManagement{AutoUpgrade: false}},
				},
			},
			wantErr: true,
			errMsg:  `node pool "pool-1" must have management.auto_upgrade enabled when release channel "Stable" is set`,
		},
		{
			name: "no node pools with release channel set",
			gke: &infrapb.GkeV3ConfigObject{
				ReleaseChannel: &infrapb.GkeReleaseChannel{Channel: "REGULAR"},
				NodePools:      []*infrapb.GkeNodePool{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAutoUpgradeWithReleaseChannel(tt.gke)
			if tt.wantErr {
				require.Error(t, err)
				assert.Equal(t, tt.errMsg, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Flatten tests
// ---------------------------------------------------------------------------

func TestFlattenGKEV3ReleaseChannel(t *testing.T) {
	tests := []struct {
		name     string
		input    *infrapb.GkeReleaseChannel
		p        []interface{}
		expected []interface{}
	}{
		{
			name:     "nil input",
			input:    nil,
			p:        []interface{}{},
			expected: nil,
		},
		{
			name:  "STABLE channel",
			input: &infrapb.GkeReleaseChannel{Channel: "STABLE"},
			p:     []interface{}{},
			expected: []interface{}{
				map[string]interface{}{
					"channel": "STABLE",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := flattenGKEV3ReleaseChannel(tt.input, tt.p)
			if tt.expected == nil {
				assert.Nil(t, result)
				return
			}
			require.Len(t, result, 1)
			resultMap := result[0].(map[string]interface{})
			expectedMap := tt.expected[0].(map[string]interface{})
			assert.Equal(t, expectedMap["channel"], resultMap["channel"])
		})
	}
}

func TestFlattenGKEV3MaintenancePolicy(t *testing.T) {
	tests := []struct {
		name     string
		input    *infrapb.GkeMaintenancePolicy
		p        []interface{}
		checkFn  func(t *testing.T, result []interface{})
		expectNl bool
	}{
		{
			name:     "nil input",
			input:    nil,
			p:        []interface{}{},
			expectNl: true,
		},
		{
			name: "daily window only",
			input: &infrapb.GkeMaintenancePolicy{
				DailyMaintenanceWindow: &infrapb.GkeDailyMaintenanceWindow{StartTime: "03:00"},
			},
			p: []interface{}{},
			checkFn: func(t *testing.T, result []interface{}) {
				m := result[0].(map[string]interface{})
				dw := m["daily_maintenance_window"].([]interface{})
				require.Len(t, dw, 1)
				dwMap := dw[0].(map[string]interface{})
				assert.Equal(t, "03:00", dwMap["start_time"])
			},
		},
		{
			name: "recurring window only",
			input: &infrapb.GkeMaintenancePolicy{
				RecurringWindow: &infrapb.GkeRecurringWindow{
					StartTime:  "2026-01-03T09:00:00Z",
					EndTime:    "2026-01-03T17:00:00Z",
					Recurrence: "FREQ=WEEKLY;BYDAY=SA,SU",
				},
			},
			p: []interface{}{},
			checkFn: func(t *testing.T, result []interface{}) {
				m := result[0].(map[string]interface{})
				rw := m["recurring_window"].([]interface{})
				require.Len(t, rw, 1)
				rwMap := rw[0].(map[string]interface{})
				assert.Equal(t, "2026-01-03T09:00:00Z", rwMap["start_time"])
				assert.Equal(t, "2026-01-03T17:00:00Z", rwMap["end_time"])
				assert.Equal(t, "FREQ=WEEKLY;BYDAY=SA,SU", rwMap["recurrence"])
			},
		},
		{
			name: "exclusions with options",
			input: &infrapb.GkeMaintenancePolicy{
				MaintenanceExclusions: []*infrapb.GkeMaintenanceExclusion{
					{
						Name:             "freeze",
						StartTime:        "2026-12-20T00:00:00Z",
						EndTime:          "2027-01-02T00:00:00Z",
						ExclusionOptions: &infrapb.GkeMaintenanceExclusionOptions{Scope: "NO_UPGRADES"},
					},
				},
			},
			p: []interface{}{},
			checkFn: func(t *testing.T, result []interface{}) {
				m := result[0].(map[string]interface{})
				excls := m["maintenance_exclusions"].([]interface{})
				require.Len(t, excls, 1)
				exclMap := excls[0].(map[string]interface{})
				assert.Equal(t, "freeze", exclMap["name"])
				assert.Equal(t, "2026-12-20T00:00:00Z", exclMap["start_time"])
				assert.Equal(t, "2027-01-02T00:00:00Z", exclMap["end_time"])
				opts := exclMap["exclusion_options"].([]interface{})
				require.Len(t, opts, 1)
				optsMap := opts[0].(map[string]interface{})
				assert.Equal(t, "NO_UPGRADES", optsMap["scope"])
			},
		},
		{
			name: "full combo",
			input: &infrapb.GkeMaintenancePolicy{
				RecurringWindow: &infrapb.GkeRecurringWindow{
					StartTime:  "2026-01-03T09:00:00Z",
					EndTime:    "2026-01-03T17:00:00Z",
					Recurrence: "FREQ=WEEKLY;BYDAY=SA,SU",
				},
				MaintenanceExclusions: []*infrapb.GkeMaintenanceExclusion{
					{Name: "e1", StartTime: "2026-12-20T00:00:00Z", EndTime: "2027-01-02T00:00:00Z"},
				},
			},
			p: []interface{}{},
			checkFn: func(t *testing.T, result []interface{}) {
				m := result[0].(map[string]interface{})
				rw := m["recurring_window"].([]interface{})
				require.Len(t, rw, 1)
				excls := m["maintenance_exclusions"].([]interface{})
				require.Len(t, excls, 1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := flattenGKEV3MaintenancePolicy(tt.input, tt.p)
			if tt.expectNl {
				assert.Nil(t, result)
				return
			}
			require.Len(t, result, 1)
			if tt.checkFn != nil {
				tt.checkFn(t, result)
			}
		})
	}
}

func TestFlattenGKEV3DailyMaintenanceWindow(t *testing.T) {
	tests := []struct {
		name     string
		input    *infrapb.GkeDailyMaintenanceWindow
		expected []interface{}
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: nil,
		},
		{
			name:  "valid",
			input: &infrapb.GkeDailyMaintenanceWindow{StartTime: "04:00"},
			expected: []interface{}{
				map[string]interface{}{"start_time": "04:00"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := flattenGKEV3DailyMaintenanceWindow(tt.input, []interface{}{})
			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				require.Len(t, result, 1)
				assert.Equal(t, tt.expected[0].(map[string]interface{})["start_time"],
					result[0].(map[string]interface{})["start_time"])
			}
		})
	}
}

func TestFlattenGKEV3RecurringWindow(t *testing.T) {
	tests := []struct {
		name     string
		input    *infrapb.GkeRecurringWindow
		expected []interface{}
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: nil,
		},
		{
			name: "valid",
			input: &infrapb.GkeRecurringWindow{
				StartTime:  "2026-01-01T00:00:00Z",
				EndTime:    "2026-01-01T08:00:00Z",
				Recurrence: "FREQ=DAILY",
			},
			expected: []interface{}{
				map[string]interface{}{
					"start_time": "2026-01-01T00:00:00Z",
					"end_time":   "2026-01-01T08:00:00Z",
					"recurrence": "FREQ=DAILY",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := flattenGKEV3RecurringWindow(tt.input, []interface{}{})
			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				require.Len(t, result, 1)
				rm := result[0].(map[string]interface{})
				em := tt.expected[0].(map[string]interface{})
				assert.Equal(t, em["start_time"], rm["start_time"])
				assert.Equal(t, em["end_time"], rm["end_time"])
				assert.Equal(t, em["recurrence"], rm["recurrence"])
			}
		})
	}
}

func TestFlattenGKEV3MaintenanceExclusions(t *testing.T) {
	tests := []struct {
		name    string
		input   []*infrapb.GkeMaintenanceExclusion
		checkFn func(t *testing.T, result []interface{})
		isNil   bool
	}{
		{
			name:  "nil input",
			input: nil,
			isNil: true,
		},
		{
			name: "single exclusion",
			input: []*infrapb.GkeMaintenanceExclusion{
				{Name: "e1", StartTime: "2026-06-01T00:00:00Z", EndTime: "2026-06-02T00:00:00Z"},
			},
			checkFn: func(t *testing.T, result []interface{}) {
				require.Len(t, result, 1)
				m := result[0].(map[string]interface{})
				assert.Equal(t, "e1", m["name"])
				assert.Equal(t, "2026-06-01T00:00:00Z", m["start_time"])
				assert.Equal(t, "2026-06-02T00:00:00Z", m["end_time"])
			},
		},
		{
			name: "multiple with options",
			input: []*infrapb.GkeMaintenanceExclusion{
				{
					Name:             "e1",
					StartTime:        "2026-06-01T00:00:00Z",
					EndTime:          "2026-06-02T00:00:00Z",
					ExclusionOptions: &infrapb.GkeMaintenanceExclusionOptions{Scope: "NO_MINOR_UPGRADES"},
				},
				{
					Name:      "e2",
					StartTime: "2026-07-01T00:00:00Z",
					EndTime:   "2026-07-05T00:00:00Z",
				},
			},
			checkFn: func(t *testing.T, result []interface{}) {
				require.Len(t, result, 2)
				m0 := result[0].(map[string]interface{})
				assert.Equal(t, "e1", m0["name"])
				opts := m0["exclusion_options"].([]interface{})
				require.Len(t, opts, 1)
				assert.Equal(t, "NO_MINOR_UPGRADES", opts[0].(map[string]interface{})["scope"])

				m1 := result[1].(map[string]interface{})
				assert.Equal(t, "e2", m1["name"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := flattenGKEV3MaintenanceExclusions(tt.input, []interface{}{})
			if tt.isNil {
				assert.Nil(t, result)
				return
			}
			if tt.checkFn != nil {
				tt.checkFn(t, result)
			}
		})
	}
}

func TestFlattenGKEV3MaintenanceExclusionOptions(t *testing.T) {
	tests := []struct {
		name     string
		input    *infrapb.GkeMaintenanceExclusionOptions
		expected []interface{}
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: nil,
		},
		{
			name:  "valid scope",
			input: &infrapb.GkeMaintenanceExclusionOptions{Scope: "NO_UPGRADES"},
			expected: []interface{}{
				map[string]interface{}{"scope": "NO_UPGRADES"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := flattenGKEV3MaintenanceExclusionOptions(tt.input, []interface{}{})
			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				require.Len(t, result, 1)
				assert.Equal(t, tt.expected[0].(map[string]interface{})["scope"],
					result[0].(map[string]interface{})["scope"])
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Round-trip tests
// ---------------------------------------------------------------------------

func TestRoundTripReleaseChannel(t *testing.T) {
	tests := []struct {
		name  string
		input []interface{}
	}{
		{
			name: "STABLE",
			input: []interface{}{
				map[string]interface{}{"channel": "STABLE"},
			},
		},
		{
			name: "RAPID",
			input: []interface{}{
				map[string]interface{}{"channel": "RAPID"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proto, err := expandToV3GkeReleaseChannel(tt.input)
			require.NoError(t, err)
			require.NotNil(t, proto)

			flat := flattenGKEV3ReleaseChannel(proto, []interface{}{})
			require.Len(t, flat, 1)

			originalMap := tt.input[0].(map[string]interface{})
			roundTripMap := flat[0].(map[string]interface{})
			assert.Equal(t, originalMap["channel"], roundTripMap["channel"])
		})
	}
}

func TestRoundTripMaintenancePolicy(t *testing.T) {
	tests := []struct {
		name    string
		input   []interface{}
		checkFn func(t *testing.T, original map[string]interface{}, roundTrip map[string]interface{})
	}{
		{
			name: "daily window",
			input: []interface{}{
				map[string]interface{}{
					"daily_maintenance_window": []interface{}{
						map[string]interface{}{"start_time": "03:00"},
					},
				},
			},
			checkFn: func(t *testing.T, original, roundTrip map[string]interface{}) {
				origDW := original["daily_maintenance_window"].([]interface{})[0].(map[string]interface{})
				rtDW := roundTrip["daily_maintenance_window"].([]interface{})[0].(map[string]interface{})
				assert.Equal(t, origDW["start_time"], rtDW["start_time"])
			},
		},
		{
			name: "recurring window with exclusions",
			input: []interface{}{
				map[string]interface{}{
					"recurring_window": []interface{}{
						map[string]interface{}{
							"start_time": "2026-01-03T09:00:00Z",
							"end_time":   "2026-01-03T17:00:00Z",
							"recurrence": "FREQ=WEEKLY;BYDAY=SA,SU",
						},
					},
					"maintenance_exclusions": []interface{}{
						map[string]interface{}{
							"name":       "freeze",
							"start_time": "2026-12-20T00:00:00Z",
							"end_time":   "2027-01-02T00:00:00Z",
							"exclusion_options": []interface{}{
								map[string]interface{}{"scope": "NO_UPGRADES"},
							},
						},
					},
				},
			},
			checkFn: func(t *testing.T, original, roundTrip map[string]interface{}) {
				// recurring window
				origRW := original["recurring_window"].([]interface{})[0].(map[string]interface{})
				rtRW := roundTrip["recurring_window"].([]interface{})[0].(map[string]interface{})
				assert.Equal(t, origRW["start_time"], rtRW["start_time"])
				assert.Equal(t, origRW["end_time"], rtRW["end_time"])
				assert.Equal(t, origRW["recurrence"], rtRW["recurrence"])

				// exclusions
				origExcls := original["maintenance_exclusions"].([]interface{})
				rtExcls := roundTrip["maintenance_exclusions"].([]interface{})
				require.Len(t, rtExcls, len(origExcls))
				origE := origExcls[0].(map[string]interface{})
				rtE := rtExcls[0].(map[string]interface{})
				assert.Equal(t, origE["name"], rtE["name"])
				assert.Equal(t, origE["start_time"], rtE["start_time"])
				assert.Equal(t, origE["end_time"], rtE["end_time"])

				origOpts := origE["exclusion_options"].([]interface{})[0].(map[string]interface{})
				rtOpts := rtE["exclusion_options"].([]interface{})[0].(map[string]interface{})
				assert.Equal(t, origOpts["scope"], rtOpts["scope"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proto, err := expandToV3GkeMaintenancePolicy(tt.input)
			require.NoError(t, err)
			require.NotNil(t, proto)

			flat := flattenGKEV3MaintenancePolicy(proto, []interface{}{})
			require.Len(t, flat, 1)

			originalMap := tt.input[0].(map[string]interface{})
			roundTripMap := flat[0].(map[string]interface{})
			tt.checkFn(t, originalMap, roundTripMap)
		})
	}
}
