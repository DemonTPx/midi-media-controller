package main

import (
	"fmt"
	"github.com/mozillazg/go-unidecode"
	"gitlab.com/gomidi/midi"
	"gitlab.com/gomidi/midi/mid"
	"gitlab.com/gomidi/midi/midimessage/channel"
	"strings"
)

type EventHandler struct {
	controller    *MidiController
	monitor       *DbusMediaPlayerMonitor
	mixer         *AudioMixer
	player        *DbusMediaPlayer
	track         *Track
	displayScroll int
	displayMode   int
}

const (
	displayArtistTitle = 0
	displayArtist      = 1
	displayTitle       = 2
	displayAlbum       = 3
)

var playerColorMap = map[string]uint8{
	"spotify":   COLOR_GREEN,
	"chrome":    COLOR_YELLOW,
	"rhythmbox": COLOR_CYAN,
}

func (h *EventHandler) Setup() {
	h.HandleVolume(h.mixer.volume)
	h.mixer.SetOnVolumeChangeCallback(h.HandleVolume)
	h.monitor.SetActivePlayerChangedCallback(h.OnActivePlayerChanged)
	h.player = h.monitor.GetActivePlayer()
	h.InitPlayer()
	h.controller.reader.Msg.Each = h.HandleMidiMessage
}

func (h *EventHandler) InitPlayer() {
	if h.player != nil {
		playbackStatus, track := h.player.FetchProperties()
		h.OnPropertiesChanged(playbackStatus, track)
		h.player.SetOnPropertiesChangedHandler(h.OnPropertiesChanged)
	} else {
		h.OnPropertiesChanged("None", Track{})
	}
}

func (h *EventHandler) OnActivePlayerChanged(player *DbusMediaPlayer) {
	if h.player != nil {
		h.player.SetOnPropertiesChangedHandler(nil)
	}

	h.player = player
	h.InitPlayer()
}

func (h *EventHandler) OnPropertiesChanged(playbackStatus string, track Track) {
	switch playbackStatus {
	case "None":
		h.controller.writer.NoteOn(NOTE_STOP, 0)
		h.controller.writer.NoteOn(NOTE_PLAY, 0)
	case "Playing":
		h.controller.writer.NoteOn(NOTE_STOP, 0)
		h.controller.writer.NoteOn(NOTE_PLAY, 127)
	case "Paused":
		h.controller.writer.NoteOn(NOTE_STOP, 127)
		h.controller.writer.NoteOn(NOTE_PLAY, 0)
	case "Stopped":
		h.controller.writer.NoteOn(NOTE_STOP, 127)
		h.controller.writer.NoteOn(NOTE_PLAY, 0)
	}

	if track.isDifferent(h.track) {
		h.displayScroll = 0
	}
	h.track = &track

	h.UpdateDisplay()
}

func (h *EventHandler) UpdateDisplay() {
	invert := INVERT_NONE
	color := COLOR_BLACK

	if h.player != nil {
		color = COLOR_WHITE
		if playerColor, ok := playerColorMap[h.player.nameLower]; ok {
			color = playerColor
		}
	}

	text := ""
	if h.track != nil {
		switch h.displayMode {
		case displayArtistTitle:
			text = PadRight(h.track.artist, 7, h.displayScroll) + PadRight(h.track.title, 7, h.displayScroll)
			invert = INVERT_TOP
		case displayArtist:
			text = PadRight(h.track.artist, 14, h.displayScroll)
			invert = INVERT_BOTH
		case displayTitle:
			text = PadRight(h.track.title, 14, h.displayScroll)
		case displayAlbum:
			text = PadRight(h.track.album, 14, h.displayScroll)
		}
	}

	h.controller.writer.SysEx(h.controller.CreateLcdDisplayData(text, color, invert))

	trackText := ""
	if h.track != nil && h.track.trackNumber != 0 {
		trackText = fmt.Sprintf("%d", h.track.trackNumber)
	}

	playerText := "NoPlayer"
	if h.player != nil {
		playerText = "  " + h.player.name
	}

	text = PadRight(playerText, 9, 0) + PadLeft(trackText, 3)
	h.controller.writer.SysEx(h.controller.CreateSegmentDisplayData(text))
}

func (h *EventHandler) HandleMidiMessage(pos *mid.Position, msg midi.Message) {
	if note, ok := msg.(channel.NoteOn); ok {
		h.handleNoteOn(&note)
	}
	if note, ok := msg.(channel.NoteOff); ok {
		h.handleNoteOff(&note)
	}
	if controlChange, ok := msg.(channel.ControlChange); ok {
		h.handleControlChange(&controlChange)
	}
}

func (h *EventHandler) handleNoteOn(note *channel.NoteOn) {
	if note.Velocity() == 0 {
		return
	}

	switch note.Key() {
	case NOTE_PREVIOUS:
		if h.player != nil {
			h.player.Previous()
			h.player.Play()
		}
	case NOTE_NEXT:
		if h.player != nil {
			h.player.Next()
			h.player.Play()
		}
	case NOTE_STOP:
		if h.player != nil {
			h.player.Stop()
		}
	case NOTE_PLAY:
		if h.player != nil {
			h.player.PlayPause()
		}
	case NOTE_ENCODER:
		h.displayMode = (h.displayMode + 1) % 4
		h.displayScroll = 0
		h.UpdateDisplay()
	case NOTE_BANK_LEFT:
		h.monitor.SelectPlayer(-1)
	case NOTE_BANK_RIGHT:
		h.monitor.SelectPlayer(+1)
	case NOTE_FADER:
		h.mixer.SetOnVolumeChangeCallback(nil)
	}
}

func (h *EventHandler) handleNoteOff(note *channel.NoteOff) {
	switch note.Key() {
	case NOTE_FADER:
		h.mixer.SetOnVolumeChangeCallback(h.HandleVolume)
	}
}

func (h *EventHandler) handleControlChange(cc *channel.ControlChange) {
	switch cc.Controller() {
	case CC_FADER:
		h.mixer.SetVolume(float32(cc.Value()) / 127)
	case CC_LED_RING:
		if cc.Value() == 1 {
			h.displayScroll += 1
		}
		if cc.Value() == 65 {
			h.displayScroll -= 1
			if h.displayScroll < 0 {
				h.displayScroll = 0
			}
		}
		h.UpdateDisplay()
	}
}

func (h *EventHandler) HandleVolume(volume float32) {
	h.controller.writer.ControlChange(CC_FADER, uint8(volume*127))
}

func PadRight(text string, l int, offset int) string {
	text = unidecode.Unidecode(text)
	if offset >= len(text) {
		return strings.Repeat(" ", l)
	}

	r := offset + l
	if r <= len(text) {
		return text[offset:r]
	}

	return text[offset:] + strings.Repeat(" ", l-len(text)+offset)
}

func PadLeft(text string, l int) string {
	text = unidecode.Unidecode(text)
	if len(text) > l {
		return text[:l]
	}

	return strings.Repeat(" ", l-len(text)) + text
}
