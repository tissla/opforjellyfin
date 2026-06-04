package cmd

import "testing"

func TestPositionalArcName(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    string
		wantArc bool
	}{
		{name: "download key", args: []string{"15"}, wantArc: false},
		{name: "download key range", args: []string{"15-17"}, wantArc: false},
		{name: "quoted arc", args: []string{"Long Ring Long Land"}, want: "Long Ring Long Land", wantArc: true},
		{name: "unquoted arc", args: []string{"Long", "Ring", "Long", "Land"}, want: "Long Ring Long Land", wantArc: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotArc := positionalArcName(tt.args)
			if gotArc != tt.wantArc {
				t.Fatalf("gotArc = %v, want %v", gotArc, tt.wantArc)
			}
			if got != tt.want {
				t.Fatalf("got = %q, want %q", got, tt.want)
			}
		})
	}
}
