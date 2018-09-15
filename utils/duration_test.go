package utils

import (
	"testing"
	"time"
)

func Test_ParseDuration(t *testing.T) {
	type args struct {
		format string
	}
	tests := []struct {
		name    string
		format  string
		want    time.Duration
		wantErr bool
	}{
		{
			name:    "Single unit",
			format:  "10sec",
			want:    time.Second * 10,
			wantErr: false,
		},
		{
			name:    "Parse error",
			format:  "10mday",
			want:    time.Second * 0,
			wantErr: true,
		},
		{
			name:    "Parse error",
			format:  "21lol",
			want:    time.Second * 0,
			wantErr: true,
		},
		{
			name:    "Multiple units",
			format:  "44hour2m10sec",
			want:    time.Second*10 + time.Minute*2 + time.Hour*44,
			wantErr: false,
		},
		{
			name:    "Random",
			format:  "arex",
			want:    time.Second * 0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDuration(tt.format)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseDuration() title = %s error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseDuration() = %v, title = %s want %v", tt.name, got, tt.want)
			}
		})
	}
}
