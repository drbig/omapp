// See LICENSE.txt for licensing information.

package overmapper

import (
	"bufio"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

const (
	BG = iota
	FG
	NOTE
	GRID
	ORIGIN
)

type Conf struct {
	MapX, MapY   int
	Scale, Level int
	Colors       []*image.Uniform
}

type Map struct {
	Width, Height int
	N, S, W, E    int
	Maps          map[image.Point]string
}

func (m *Map) String() string {
	return fmt.Sprintf("Map %dx%d (%d)", m.Width, m.Height, len(m.Maps))
}

var (
	ErrMultiChars = errors.New("multiple characters detected")
	ErrNotFound   = errors.New("no seen files found")

	Config = Conf{
		180, 180,
		2, 10,
		[]*image.Uniform{
			image.NewUniform(color.RGBA{0, 0, 0, 255}),
			image.NewUniform(color.RGBA{255, 255, 255, 255}),
			image.NewUniform(color.RGBA{0, 0, 255, 255}),
			image.NewUniform(color.RGBA{255, 0, 0, 180}),
			image.NewUniform(color.RGBA{0, 255, 0, 180}),
		},
	}

	rxpSeenFile = regexp.MustCompile(`\#(.*?)\.seen\.(-?\d+)\.(-?\d+)$`)
)

func NewMap(path string) (*Map, error) {
	var id string
	m := &Map{Maps: make(map[image.Point]string)}
	err := filepath.Walk(path, func(p string, i os.FileInfo, er error) error {
		if er != nil {
			return nil
		}
		if p != path && i.IsDir() {
			return filepath.SkipDir
		}
		if rm := rxpSeenFile.FindStringSubmatch(p); rm != nil {
			if id == "" {
				id = rm[1]
			}
			if id != rm[1] {
				return ErrMultiChars
			}
			x, err := strconv.Atoi(rm[2])
			if err != nil {
				return err
			}
			y, err := strconv.Atoi(rm[3])
			if err != nil {
				return err
			}
			m.Maps[image.Point{x, y}] = p
			if x < m.W {
				m.W = x
			}
			if x > m.E {
				m.E = x
			}
			if y > m.N {
				m.N = y
			}
			if y < m.S {
				m.S = y
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(m.Maps) == 0 {
		return nil, ErrNotFound
	}
	m.Width = m.E - m.W + 1
	m.Height = m.N - m.S + 1
	return m, nil
}

func skipToLevel(s *bufio.Scanner) error {
	lstr := fmt.Sprintf("L %d", Config.Level)
	for s.Scan() {
		line := s.Text()
		if line == lstr {
			break
		}
	}
	return s.Err()
}

func parseSeen(i io.Reader) ([]image.Rectangle, error) {
	var boxes []image.Rectangle
	var visited, length, position int
	for {
		n, err := fmt.Fscanf(i, "%d %d", &visited, &length)
		if err != nil && err != io.EOF {
			return nil, err
		}
		if n != 2 {
			break
		}
		if visited == 1 {
			x0 := position % Config.MapX
			y0 := position / Config.MapX
			x1 := (position + length - 1) % Config.MapX
			y1 := (position + length - 1) / Config.MapX
			if y0 == y1 {
				boxes = append(boxes, image.Rect(x0, y0, x1, y1))
			} else {
				boxes = append(boxes, image.Rect(x0, y0, Config.MapX-1, y0))
				if y0+1 != y1 {
					boxes = append(boxes, image.Rect(0, y0+1, Config.MapX-1, y1-1))
				}
				boxes = append(boxes, image.Rect(0, y1, x1, y1))
			}
		}
		position += length
	}
	return boxes, nil
}

func parseNotes(s *bufio.Scanner) []image.Rectangle {
	var notes []image.Rectangle
	var nx, ny int
	for s.Scan() {
		data := strings.NewReader(s.Text())
		n, _ := fmt.Fscanf(data, "N %d %d", &nx, &ny)
		if n != 2 {
			break
		}
		notes = append(notes, image.Rect(nx, ny, nx, ny))
		s.Scan()
	}
	return notes
}

func transformBox(r *image.Rectangle, x, y int) {
	r.Min.X = (x * (Config.MapX * Config.Scale)) + (r.Min.X * Config.Scale)
	r.Min.Y = (y * (Config.MapY * Config.Scale)) + (r.Min.Y * Config.Scale)
	r.Max.X = (x * (Config.MapX * Config.Scale)) + ((r.Max.X + 1) * Config.Scale)
	r.Max.Y = (y * (Config.MapY * Config.Scale)) + ((r.Max.Y + 1) * Config.Scale)
}

func drawGrid(i draw.Image, color image.Image, x, y int) {
	var b image.Rectangle
	b = image.Rect(0, 0, Config.MapX-1, 0) // top-edge
	transformBox(&b, x, y)
	draw.Draw(i, b, color, image.ZP, draw.Over)
	b = image.Rect(0, Config.MapY-1, Config.MapX-1, Config.MapY-1) // bottom-edge
	transformBox(&b, x, y)
	draw.Draw(i, b, color, image.ZP, draw.Over)
	b = image.Rect(0, 0, 0, Config.MapY-1) // left-edge
	transformBox(&b, x, y)
	draw.Draw(i, b, color, image.ZP, draw.Over)
	b = image.Rect(Config.MapX-1, 0, Config.MapX-1, Config.MapY-1) // right-edge
	transformBox(&b, x, y)
	draw.Draw(i, b, color, image.ZP, draw.Over)
}

func (m *Map) Draw() (*image.RGBA, error) {
	img := image.NewRGBA(image.Rect(0, 0, m.Width*(Config.MapX*Config.Scale), m.Height*(Config.MapY*Config.Scale)))
	draw.Draw(img, img.Bounds(), Config.Colors[BG], image.ZP, draw.Src)
	iy := 0
	for y := m.S; y <= m.N; y += 1 {
		ix := 0
		for x := m.W; x <= m.E; x += 1 {
			path, present := m.Maps[image.Point{x, y}]
			if present {
				file, err := os.Open(path)
				if err != nil {
					return nil, err
				}
				scanner := bufio.NewScanner(file)
				if err := skipToLevel(scanner); err != nil {
					return nil, err
				}
				scanner.Scan()
				data := strings.NewReader(scanner.Text())
				boxes, err := parseSeen(data)
				if err != nil {
					return nil, err
				}
				for _, box := range boxes {
					transformBox(&box, ix, iy)
					draw.Draw(img, box, Config.Colors[FG], image.ZP, draw.Src)
				}
				scanner.Scan() // E 10
				scanner.Scan() // 0 32400
				for _, box := range parseNotes(scanner) {
					transformBox(&box, ix, iy)
					draw.Draw(img, box, Config.Colors[NOTE], image.ZP, draw.Src)
				}
				file.Close()
			}
			if x == 0 && y == 0 {
				drawGrid(img, Config.Colors[ORIGIN], ix, iy)
			} else {
				drawGrid(img, Config.Colors[GRID], ix, iy)
			}
			ix += 1
		}
		iy += 1
	}
	return img, nil
}
