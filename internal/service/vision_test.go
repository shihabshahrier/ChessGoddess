package service

import "testing"

func TestDetectImageMIME(t *testing.T) {
	cases := []struct {
		name    string
		data    []byte
		want    string
		wantErr bool
	}{
		{
			name: "png",
			data: []byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a, 0, 0, 0, 0},
			want: "image/png",
		},
		{
			name: "jpeg",
			data: append([]byte{0xff, 0xd8, 0xff, 0xe0}, make([]byte, 16)...),
			want: "image/jpeg",
		},
		{
			name: "webp",
			data: []byte("RIFF\x00\x00\x00\x00WEBPVP8 "),
			want: "image/webp",
		},
		{
			name:    "gif rejected",
			data:    []byte("GIF89a\x00\x00\x00\x00"),
			wantErr: true,
		},
		{
			name:    "empty rejected",
			data:    nil,
			wantErr: true,
		},
		{
			name:    "non-image rejected",
			data:    []byte("just some plain text, definitely not an image file"),
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mime, err := detectImageMIME(tc.data)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error, got mime %q", mime)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if mime != tc.want {
				t.Errorf("mime = %q, want %q", mime, tc.want)
			}
		})
	}
}

func TestNormalizeFEN(t *testing.T) {
	// A board-only field gets padded to a full, legal FEN.
	got := normalizeFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR")
	if !isValidFEN(got) {
		t.Errorf("normalizeFEN produced an invalid FEN: %q", got)
	}

	// An already-complete FEN is left intact.
	full := "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq - 0 1"
	if normalizeFEN(full) != full {
		t.Errorf("normalizeFEN altered a complete FEN: %q", normalizeFEN(full))
	}
}

func TestIsValidFEN(t *testing.T) {
	if !isValidFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1") {
		t.Error("start position should be valid")
	}
	if isValidFEN("not a fen at all") {
		t.Error("garbage should be invalid")
	}
}
