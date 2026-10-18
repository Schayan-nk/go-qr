package qr

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
)

// testdata/segno.json is the same texts drawn by segno, an independent,
// widely used Python QR encoder (level M, byte mode, the mask forced or its
// own choice). Generated once with testdata/segno_gen.py; every forced-mask grid
// must match module for module -- which checks the data encoding, the error
// correction, the interleaving, the placement, the format and version bits
// and every mask pattern against somebody else's implementation.
type reference struct {
	Text    string   `json:"text"`
	Mask    int      `json:"mask"`
	Chosen  int      `json:"chosen"`
	Version int      `json:"version"`
	Rows    []string `json:"rows"`
}

func loadReferences(t *testing.T) []reference {
	t.Helper()
	raw, err := os.ReadFile("testdata/segno.json")
	if err != nil {
		t.Fatal(err)
	}
	var refs []reference
	if err := json.Unmarshal(raw, &refs); err != nil {
		t.Fatal(err)
	}
	return refs
}

func grid(c *Code) []string {
	rows := make([]string, c.Size)
	for y := 0; y < c.Size; y++ {
		var b strings.Builder
		for x := 0; x < c.Size; x++ {
			if c.Dark(x, y) {
				b.WriteByte('1')
			} else {
				b.WriteByte('0')
			}
		}
		rows[y] = b.String()
	}
	return rows
}

func TestEveryGridMatchesAnIndependentEncoder(t *testing.T) {
	refs := loadReferences(t)
	versions := map[int]bool{}
	for _, ref := range refs {
		mask := ref.Mask
		if mask < 0 {
			mask = ref.Chosen // compared on segno's own choice; see below
		}
		c, err := encode(ref.Text, mask)
		if err != nil {
			t.Fatalf("%d bytes: %v", len(ref.Text), err)
		}
		if want := ref.Version*4 + 17; c.Size != want {
			t.Errorf("%d bytes: size %d, reference is version %d (%d)", len(ref.Text), c.Size, ref.Version, want)
			continue
		}
		got := grid(c)
		for y := range got {
			if got[y] != ref.Rows[y] {
				t.Errorf("%d bytes, mask %d: row %d differs\n got  %s\n want %s",
					len(ref.Text), mask, y, got[y], ref.Rows[y])
				break
			}
		}
		versions[ref.Version] = true
	}
	for _, v := range []int{1, 5, 6, 7, 8, 9, 10} {
		if !versions[v] {
			t.Errorf("no reference grid at version %d", v)
		}
	}
}

// The mask picked by penalty. Any mask decodes; the penalty only avoids
// patterns a camera finds hard. Logged rather than failed when it differs,
// because the standard's own penalty rules are read differently between
// implementations.
func TestTheChosenMaskIsReported(t *testing.T) {
	for _, ref := range loadReferences(t) {
		if ref.Mask >= 0 {
			continue
		}
		c, err := Encode(ref.Text)
		if err != nil {
			t.Fatal(err)
		}
		mine := -1
		for m := 0; m < 8; m++ {
			if forced, _ := encode(ref.Text, m); strings.Join(grid(forced), "") == strings.Join(grid(c), "") {
				mine = m
			}
		}
		if mine != ref.Chosen {
			t.Logf("%d bytes: chose mask %d, segno chose %d", len(ref.Text), mine, ref.Chosen)
		}
	}
}

func TestTooLongIsRefused(t *testing.T) {
	if _, err := Encode(strings.Repeat("a", 214)); !errors.Is(err, ErrTooLong) {
		t.Errorf("214 bytes: err = %v, want ErrTooLong", err)
	}
	if _, err := Encode(strings.Repeat("a", 213)); err != nil {
		t.Errorf("213 bytes, the most version 10 holds: %v", err)
	}
}

func TestTheSVGIsOnlyAPicture(t *testing.T) {
	c, err := Encode("otpauth://totp/docfind:ali?secret=JBSWY3DPEHPK3PXP")
	if err != nil {
		t.Fatal(err)
	}
	svg := c.SVG()
	if !strings.HasPrefix(svg, "<svg ") || !strings.HasSuffix(svg, "</svg>") {
		t.Errorf("not an SVG document: %.60s...", svg)
	}
	// Modules and nothing else: no script, no link, no handler, and not the
	// secret itself in readable form.
	for _, bad := range []string{"<script", "href", " on", "otpauth", "JBSWY3DP"} {
		if strings.Contains(svg, bad) {
			t.Errorf("the SVG carries %q", bad)
		}
	}
}
