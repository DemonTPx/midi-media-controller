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

const (
	CC_FADER        uint8 = 70
	CC_LED_RING     uint8 = 80
	CC_LED_METER    uint8 = 90
	NOTE_ENCODER    uint8 = 0
	NOTE_PREVIOUS   uint8 = 20
	NOTE_NEXT       uint8 = 21
	NOTE_STOP       uint8 = 22
	NOTE_PLAY       uint8 = 23
	NOTE_BANK_LEFT  uint8 = 25
	NOTE_BANK_RIGHT uint8 = 26
	NOTE_FADER      uint8 = 110
	COLOR_BLACK     uint8 = 0
	COLOR_RED       uint8 = 1
	COLOR_GREEN     uint8 = 2
	COLOR_YELLOW    uint8 = 3
	COLOR_BLUE      uint8 = 4
	COLOR_MAGENTA   uint8 = 5
	COLOR_CYAN      uint8 = 6
	COLOR_WHITE     uint8 = 7
	INVERT_NONE     uint8 = 0
	INVERT_TOP      uint8 = 1
	INVERT_BOTTOM   uint8 = 2
	INVERT_BOTH     uint8 = 3
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

	c.writer.ControlChange(CC_FADER, 0)
	c.writer.ControlChange(CC_LED_RING, 64)
	c.writer.ControlChange(CC_LED_METER, 0)

	c.writer.SysEx(c.CreateSegmentDisplayData(""))
	c.writer.SysEx(c.CreateLcdDisplayData("", COLOR_BLACK, INVERT_NONE))

	return nil
}

func (c *MidiController) CreateLcdDisplayData(characters string, color uint8, invert uint8) []byte {
	data := make([]byte, 14)
	copy(data, unidecode.Unidecode(characters))

	colorCode := color | (invert << 4)

	return append([]byte{0x00, 0x20, 0x32, 0x41, 0x4c, 0x00, colorCode}, data...)
}

func (c *MidiController) CreateSegmentDisplayData(characters string) []byte {
	textBytes := make([]byte, 12)
	copy(textBytes, unidecode.Unidecode(characters))
	data := lcd7bitRender(textBytes)
	dots := lcd7bitRenderDots(textBytes)

	return append([]byte{0x00, 0x20, 0x32, 0x41, 0x37}, append(data, dots...)...)
}
