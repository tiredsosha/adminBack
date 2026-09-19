package protocols

import "testing"

func TestPjlinkPowerStatus(t *testing.T) {
	tests := []struct {
		name   string
		values []string
		want   int
	}{
		{name: "on", values: []string{"1"}, want: 200},
		{name: "warming", values: []string{"3"}, want: 200},
		{name: "standby", values: []string{"0"}, want: 521},
		{name: "cooling", values: []string{"2"}, want: 521},
		{name: "nul padded", values: []string{"\x00 1\r\n"}, want: 200},
		{name: "empty", values: nil, want: 520},
		{name: "unknown", values: []string{"4"}, want: 520},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := pjlinkPowerStatus(test.values); got != test.want {
				t.Fatalf("pjlinkPowerStatus(%q) = %d, want %d", test.values, got, test.want)
			}
		})
	}
}
