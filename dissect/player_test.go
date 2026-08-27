package dissect

import (
	"bytes"
	"testing"
)

func TestPlayerIndicators(t *testing.T) {
	tests := []struct {
		name        string
		codeVersion int
		wantID      []byte
		wantUI      []byte
	}{
		{
			name:        "Y11S2 Alpha03 markers",
			codeVersion: Y11S2Alpha3,
			wantID:      []byte{0x8C, 0x61, 0x1A, 0x75, 0x23},
			wantUI:      []byte{0x70, 0xFA, 0xAF, 0x28},
		},
		{
			name:        "Y11S2 Alpha04 restores legacy markers",
			codeVersion: 9838928,
			wantID:      []byte{0x33, 0xD8, 0x3D, 0x4F, 0x23},
			wantUI:      []byte{0x38, 0xDF, 0xEE, 0x88},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gotID, gotUI := playerIndicators(test.codeVersion)
			if !bytes.Equal(gotID, test.wantID) {
				t.Fatalf("id marker = % X, want % X", gotID, test.wantID)
			}
			if !bytes.Equal(gotUI, test.wantUI) {
				t.Fatalf("ui marker = % X, want % X", gotUI, test.wantUI)
			}
		})
	}
}
