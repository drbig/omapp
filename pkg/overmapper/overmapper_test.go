package overmapper

import (
	"bufio"
	"image"
	"strings"
	"testing"
)

type seenCase struct {
	in  string
	out []image.Rectangle
}

func TestParseSeen(t *testing.T) {
	cases := []seenCase{
		{`0 38400`, nil},
		{`1 10`, []image.Rectangle{image.Rect(0, 0, 9, 0)}},
		{`0 10 1 5`, []image.Rectangle{image.Rect(10, 0, 14, 0)}},
		{`1 200`, []image.Rectangle{image.Rect(0, 0, 179, 0), image.Rect(0, 1, 19, 1)}},
		{`0 10 1 900`, []image.Rectangle{image.Rect(10, 0, 179, 0), image.Rect(0, 1, 179, 4), image.Rect(0, 5, 9, 5)}},
	}
	for idx, test := range cases {
		boxes, err := parseSeen(strings.NewReader(test.in))
		if err != nil {
			t.Errorf("(%d) Err: %s", idx+1, err)
			continue
		}
		if len(boxes) != len(test.out) {
			t.Errorf("(%d) Len mismatch: %d != %d", idx+1, len(boxes), len(test.out))
			continue
		}
		for i, b := range boxes {
			if b != test.out[i] {
				t.Errorf("(%d, %d) Mismatch: %s != %s", idx+1, i+1, b, test.out[i])
			}
		}
	}
}

type notesCase struct {
	in  string
	out []image.Rectangle
}

func TestParseNotes(t *testing.T) {
	cases := []notesCase{
		{"Nope", nil},
		{"N 2 99\nTest\n", []image.Rectangle{image.Rect(2, 99, 2, 99)}},
		{"N 13 15\nOne\nN 100 20\nWhatever\n", []image.Rectangle{image.Rect(13, 15, 13, 15), image.Rect(100, 20, 100, 20)}},
	}
	for idx, test := range cases {
		boxes := parseNotes(bufio.NewScanner(strings.NewReader(test.in)))
		if len(boxes) != len(test.out) {
			t.Errorf("(%d) Len mismatch: %d != %d", idx+1, len(boxes), len(test.out))
			continue
		}
		for i, b := range boxes {
			if b != test.out[i] {
				t.Errorf("(%d, %d) Mismatch: %s != %s", idx+1, i+1, b, test.out[i])
			}
		}
	}
}

func TestSkipToLevel(t *testing.T) {
	input := bufio.NewScanner(strings.NewReader(`# version 24
L 0
0 32400
E 0
0 32400
L 1
0 32400
E 1
0 32400
L 10
1 10 0 200 1 40
L 11
0 32400
`))

	err := skipToLevel(input)
	if err != nil {
		t.Errorf("Err: %s", err)
	}
	input.Scan()
	line := input.Text()
	if line != `1 10 0 200 1 40` {
		t.Errorf("Mismatch: %s != %s", line, `1 10 0 200 1 40`)
	}
}

func TestTransformBox(t *testing.T) {
	v := image.Rect(0, 0, 10, 10)
	o := image.Rect((Config.MapX * Config.Scale), (Config.MapY * Config.Scale), (Config.MapX*Config.Scale)+(11*Config.Scale), (Config.MapY*Config.Scale)+(11*Config.Scale))
	transformBox(&v, 1, 1)
	if v != o {
		t.Errorf("Mismatch: %s != %s", v, o)
	}
}
