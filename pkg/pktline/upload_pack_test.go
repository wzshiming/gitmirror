package pktline

import (
	"bytes"
	"strings"
	"testing"
)

func TestParseUploadPackRequest(t *testing.T) {
	// Test with properly formatted pktline data
	// 0032 = 50 bytes (including the 4-byte length)
	// "want " (5) + 40-char hash + "\n" (1) = 46 bytes + 4 = 50 = 0x32
	tests := []struct {
		name      string
		input     string
		wantWants []string
		wantHaves []string
		wantDone  bool
	}{
		{
			name: "simple want and done",
			// Build the pktline manually: want line (46 chars) + 4 = 50 = 0x32
			input: "0032want 0123456789abcdef0123456789abcdef01234567\n" +
				"0000" +
				"0009done\n" +
				"0000",
			wantWants: []string{"0123456789abcdef0123456789abcdef01234567"},
			wantDone:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := ParseUploadPackRequest(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("ParseUploadPackRequest() error = %v", err)
			}

			if len(req.Wants) != len(tt.wantWants) {
				t.Errorf("got %d wants, want %d", len(req.Wants), len(tt.wantWants))
			} else {
				for i := range req.Wants {
					if req.Wants[i] != tt.wantWants[i] {
						t.Errorf("Wants[%d] = %q, want %q", i, req.Wants[i], tt.wantWants[i])
					}
				}
			}

			if req.Done != tt.wantDone {
				t.Errorf("Done = %v, want %v", req.Done, tt.wantDone)
			}
		})
	}
}

func TestEncodeUploadPackRequest(t *testing.T) {
	req := &UploadPackRequest{
		Wants:        []string{"0123456789abcdef0123456789abcdef01234567"},
		Capabilities: []string{"multi_ack"},
		Done:         true,
	}

	encoded := EncodeUploadPackRequest(req)

	// The encoded result should contain the want line with capabilities
	if !bytes.Contains(encoded, []byte("want 0123456789abcdef0123456789abcdef01234567 multi_ack")) {
		t.Errorf("Encoded request doesn't contain expected want line with capabilities")
	}

	// Should contain done
	if !bytes.Contains(encoded, []byte("done")) {
		t.Errorf("Encoded request doesn't contain done")
	}
}
