package main

import (
	"fmt"
	"github.com/mozillazg/go-unidecode"
	"gitlab.com/gomidi/midi/mid"
	"log"
	"strings"
)

type MidiController struct {
	driver mid.Driver
	name   string
	in     mid.In
	out    mid.Out
	writer *mid.Writer
	reader *mid.Reader
}

func NewMidiController(driver mid.Driver, name string) *MidiController {
	return &MidiController{
		driver: driver,
		name:   name,
	}
}

const (
	CcFader       uint8 = 70
	CcLedRing     uint8 = 80
	CcLedMeter    uint8 = 90
	NoteEncoder   uint8 = 0
	NoteTime      uint8 = 1
	NotePrevious  uint8 = 20
	NoteNext      uint8 = 21
	NoteStop      uint8 = 22
	NotePlay      uint8 = 23
	NoteBankLeft  uint8 = 25
	NoteBankRight uint8 = 26
	NoteFader     uint8 = 110
	ColorBlack    uint8 = 0
	ColorRed      uint8 = 1
	ColorGreen    uint8 = 2
	ColorYellow   uint8 = 3
	ColorBlue     uint8 = 4
	ColorMagenta  uint8 = 5
	ColorCyan     uint8 = 6
	ColorWhite    uint8 = 7
	InvertNone    uint8 = 0
	InvertTop     uint8 = 1
	InvertBottom  uint8 = 2
	InvertBoth    uint8 = 3
)

func (c *MidiController) OpenOut() error {
	outs, err := c.driver.Outs()

	if err != nil {
		return err
	}

	for _, out := range outs {
		if strings.HasPrefix(out.String(), c.name) {
			log.Printf("opening out port with name %s\n", out.String())
			c.out = out

			err := c.out.Open()
			if err != nil {
				return err
			}

			c.writer = mid.ConnectOut(c.out)
			c.writer.ConsolidateNotes(false)

			return nil
		}
	}

	return fmt.Errorf("no midi output found starting with name %s", c.name)
}

func (c *MidiController) OpenIn() error {
	ins, err := c.driver.Ins()

	if err != nil {
		return err
	}

	for _, in := range ins {
		if strings.HasPrefix(in.String(), c.name) {
			log.Printf("opening in port with name %s\n", in.String())
			c.in = in

			err := c.in.Open()
			if err != nil {
				return err
			}

			c.reader = mid.NewReader(mid.NoLogger())

			return mid.ConnectIn(c.in, c.reader)
		}
	}

	return fmt.Errorf("no midi input found starting with name %s", c.name)
}

func (c *MidiController) Close() error {
	if c.in != nil {
		c.in.Close()
	}
	if c.out != nil {
		c.out.Close()
	}

	return nil
}

func (c *MidiController) Reset() error {
	for n := uint8(1); n <= 35; n++ {
		err := c.writer.NoteOn(n, 0)
		if err != nil {
			return err
		}
	}

	c.writer.ControlChange(CcFader, 0)
	c.writer.ControlChange(CcLedRing, 64)
	c.writer.ControlChange(CcLedMeter, 0)

	c.writer.SysEx(c.CreateSegmentDisplayData(EmptySegmentDisplayData()))
	c.writer.SysEx(c.CreateLcdDisplayData("", ColorBlack, InvertNone))

	return nil
}

func (c *MidiController) CreateLcdDisplayData(characters string, color uint8, invert uint8) []byte {
	data := make([]byte, 14)
	copy(data, unidecode.Unidecode(characters))

	colorCode := color | (invert << 4)

	return append([]byte{0x00, 0x20, 0x32, 0x41, 0x4c, 0x00, colorCode}, data...)
}

func (c *MidiController) CreateSegmentDisplayData(data SegmentDisplayData) []byte {
	return append([]byte{0x00, 0x20, 0x32, 0x41, 0x37}, append(data.text, data.dots...)...)
}

type SegmentDisplayData struct {
	text []byte
	dots []byte
}

func EmptySegmentDisplayData() SegmentDisplayData {
	return SegmentDisplayData{
		text: make([]byte, 12),
		dots: make([]byte, 2),
	}
}

func NewSegmentDisplayData(text string) SegmentDisplayData {
	textBytes := make([]byte, 12)
	copy(textBytes, unidecode.Unidecode(text))

	return SegmentDisplayData{
		text: lcd7bitRender(textBytes),
		dots: lcd7bitRenderDots(textBytes),
	}
}

func NewSegmentDisplayDataTime(text string) SegmentDisplayData {
	textBytes := make([]byte, 12)
	copy(textBytes, unidecode.Unidecode(text))

	return SegmentDisplayData{
		text: lcd7bitRender(textBytes),
		dots: []byte{0x50, 0x00},
	}
}
