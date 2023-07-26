package duration

import (
	"encoding/json"
	"testing"
	"time"
)

const dateLayout = "Jan 2, 2006 at 03:04:05"

func makeTime(t *testing.T, s string) time.Time {
	result, err := time.Parse(dateLayout, s)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func TestCanShift(t *testing.T) {
	cases := []struct {
		from     string
		duration Duration
		want     string
	}{
		{"Jan 1, 2018 at 00:00:00", Duration{}, "Jan 1, 2018 at 00:00:00"},
		{"Jan 1, 2018 at 00:00:00", Duration{Y: 1}, "Jan 1, 2019 at 00:00:00"},
		{"Jan 1, 2018 at 00:00:00", Duration{M: 1}, "Feb 1, 2018 at 00:00:00"},
		{"Jan 1, 2018 at 00:00:00", Duration{M: 2}, "Mar 1, 2018 at 00:00:00"},
		{"Jan 1, 2018 at 00:00:00", Duration{W: 1}, "Jan 8, 2018 at 00:00:00"},
		{"Jan 1, 2018 at 00:00:00", Duration{D: 1}, "Jan 2, 2018 at 00:00:00"},
		{"Jan 1, 2018 at 00:00:00", Duration{TH: 1}, "Jan 1, 2018 at 01:00:00"},
		{"Jan 1, 2018 at 00:00:00", Duration{TM: 1}, "Jan 1, 2018 at 00:01:00"},
		{"Jan 1, 2018 at 00:00:00", Duration{TS: 1}, "Jan 1, 2018 at 00:00:01"},
		{"Jan 1, 2018 at 00:00:00", Duration{
			Y:  10,
			M:  5,
			D:  8,
			TH: 5,
			TM: 10,
			TS: 6,
			//T: 5*time.Hour + 10*time.Minute + 6*time.Second,
		},
			"Jun 9, 2028 at 05:10:06",
		},
	}

	for k, c := range cases {
		from := makeTime(t, c.from)
		want := makeTime(t, c.want)

		got := c.duration.Shift(from)
		if !want.Equal(got) {
			t.Fatalf("Case %d: want=%s, got=%s", k, want, got)
		}
	}
}

func TestCanMaintainHourThroughDST(t *testing.T) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}

	current, err := time.ParseInLocation(dateLayout, "Jan 1, 2018 at 00:00:00", loc)
	if err != nil {
		t.Fatal(err)
	}

	sut := Duration{D: 1}
	for d := 0; d < 365; d++ {
		if got := current.Hour(); got != 0 {
			t.Fatalf("Day %d: want=0, got=%d", d, got)
		}
		current = sut.Shift(current)
	}
}

func TestCanParse(t *testing.T) {
	cases := []struct {
		from string
		want Duration
	}{
		{"P1Y", Duration{Y: 1}},
		{"P-1Y", Duration{Y: -1}},
		{"P1M", Duration{M: 1}},
		{"P-1M", Duration{M: -1}},
		{"P2M", Duration{M: 2}},
		{"P-2M", Duration{M: -2}},
		{"P1W", Duration{W: 1}},
		{"P-1W", Duration{W: -1}},
		{"P1D", Duration{D: 1}},
		{"P-1D", Duration{D: -1}},
		{"PT1H", Duration{TH: 1}},
		{"PT-1H", Duration{TH: -1}},
		{"PT1M", Duration{TM: 1}},
		{"PT-1M", Duration{TM: -1}},
		{"PT1S", Duration{TS: 1}},
		{"PT-1S", Duration{TS: -1}},
		{"PT-24H", Duration{TH: -24}},
		{"P10Y5M8DT5H10M6S", Duration{Y: 10, M: 5, D: 8, TH: 5, TM: 10, TS: 6}},
		{"P-10Y5M8DT5H10M6S", Duration{Y: -10, M: 5, D: 8, TH: 5, TM: 10, TS: 6}},
		{"P-10Y-5M8DT5H10M6S", Duration{Y: -10, M: -5, D: 8, TH: 5, TM: 10, TS: 6}},
	}

	for k, c := range cases {
		got, err := ParseISO8601(c.from)
		if err != nil {
			t.Fatal(err)
		}
		if c.want != got {
			t.Fatalf("Case %d: want=%+v, got=%+v", k, c.want, got)
		}
	}
}

func TestCanRejectBadString(t *testing.T) {
	cases := []string{
		"",
		"PP1D",
		"P1D2F",
		"P2F",
	}

	for _, c := range cases {
		_, err := ParseISO8601(c)
		if err == nil {
			t.Fatalf("%s: Expected error, got none", c)
		}
	}
}

func TestCanStringifyZeroValue(t *testing.T) {
	sut := Duration{}
	want := "P0D"
	got := sut.String()
	if want != got {
		t.Fatalf("want=%s, got=%s", want, got)
	}
}

func TestCanStringify(t *testing.T) {
	cases := []string{
		"P1Y",
		"P2M",
		"P3W",
		"P4D",
		"PT5H",
		"PT6M",
		"PT7S",
		"P1Y2M3W4DT5H6M7S",
	}
	for _, want := range cases {
		sut, err := ParseISO8601(want)
		if err != nil {
			t.Fatal(err)
		}
		got := sut.String()
		if want != got {
			t.Fatalf("Want %s, got %s", want, got)
		}
	}
}

func TestCanMarshalJSON(t *testing.T) {
	s := "P1Y2M3W4DT5H6M7S"
	sut, _ := ParseISO8601(s)

	b, err := json.Marshal(sut)
	if err != nil {
		t.Fatal(err)
	}

	want := `"P1Y2M3W4DT5H6M7S"`
	got := string(b)
	if got != want {
		t.Fatalf("want=%s, got=%s", want, got)
	}
}

func TestCanUnmarshalJSON(t *testing.T) {
	j := []byte(`"P1Y2M3W4DT5H6M7S"`)
	var got Duration
	err := json.Unmarshal(j, &got)
	if err != nil {
		t.Fatal(err)
	}

	s := "P1Y2M3W4DT5H6M7S"
	want, _ := ParseISO8601(s)

	if got != want {
		t.Fatalf("want=%+v, got=%+v", want, got)
	}
}

func TestCanRejectDurationInJSON(t *testing.T) {
	j := []byte(`"PZY"`)
	var got Duration
	err := json.Unmarshal(j, &got)
	if err == nil {
		t.Fatal("expected error, got none")
	}
}

func TestCanRejectBadJSON(t *testing.T) {
	j := []byte(`{"foo":"bar"}`)
	var got Duration
	err := json.Unmarshal(j, &got)
	if err == nil {
		t.Fatal("expected error, got none")
	}
}
