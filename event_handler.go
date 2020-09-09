package main

import (
	"fmt"
	"github.com/mozillazg/go-unidecode"
	"gitlab.com/gomidi/midi"
	"gitlab.com/gomidi/midi/mid"
	"gitlab.com/gomidi/midi/midimessage/channel"
	"strings"
	"time"
)

type EventHandler struct {
	controller *MidiController
	monitor    *DbusMediaPlayerMonitor
	mixer      *AudioMixer
	player     *DbusMediaPlayer
	track      *Track

	displayScroll int
	displayMode   int

	segmentDisplayMode int
}

func NewEventHandler(controller *MidiController, monitor *DbusMediaPlayerMonitor, mixer *AudioMixer) *EventHandler {
	return &EventHandler{
		controller: controller,
		monitor:    monitor,
		mixer:      mixer,
	}
}

const (
	displayArtistTitle = 0
	displayArtist      = 1
	displayTitle       = 2
	displayAlbum       = 3
)

const (
	segmentDisplayPlayer = 0
	segmentDisplayTime   = 1
)

var playerColorMap = map[string]uint8{
	"spotify":   ColorGreen,
	"chrome":    ColorYellow,
	"rhythmbox": ColorCyan,
}

func (h *EventHandler) Setup() {
	h.HandleVolume(h.mixer.volume)
	h.mixer.SetOnVolumeChangeCallback(h.HandleVolume)
	h.monitor.SetActivePlayerChangedCallback(h.OnActivePlayerChanged)
	h.player = h.monitor.GetActivePlayer()
	h.InitPlayer()
	h.controller.reader.Msg.Each = h.HandleMidiMessage
	go Ticker(250*time.Millisecond, h)
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
		h.controller.writer.NoteOn(NoteStop, 0)
		h.controller.writer.NoteOn(NotePlay, 0)
	case "Playing":
		h.controller.writer.NoteOn(NoteStop, 0)
		h.controller.writer.NoteOn(NotePlay, 127)
	case "Paused":
		h.controller.writer.NoteOn(NoteStop, 127)
		h.controller.writer.NoteOn(NotePlay, 0)
	case "Stopped":
		h.controller.writer.NoteOn(NoteStop, 127)
		h.controller.writer.NoteOn(NotePlay, 0)
	}

	if track.isDifferent(h.track) {
		h.ResetDisplayScroll()
	}
	h.track = &track

	h.UpdateDisplay()
}

func (h *EventHandler) UpdateDisplay() {
	invert := InvertNone
	color := ColorBlack

	if h.player != nil {
		color = ColorWhite
		if playerColor, ok := playerColorMap[h.player.nameLower]; ok {
			color = playerColor
		}
	}

	text := ""
	if h.track != nil {
		switch h.displayMode {
		case displayArtistTitle:
			text = PadRight(h.track.artist, 7, h.displayScroll) + PadRight(h.track.title, 7, h.displayScroll)
			invert = InvertTop
		case displayArtist:
			text = PadRight(h.track.artist, 14, h.displayScroll)
			invert = InvertBoth
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

	segmentText := ""
	segmentDisplayData := SegmentDisplayData{}
	switch h.segmentDisplayMode {
	case segmentDisplayPlayer:
		segmentText = "NoPlayer"
		if h.player != nil {
			segmentText = "  " + h.player.name
		}
		text = PadRight(segmentText, 9, 0) + PadLeft(trackText, 3)
		segmentDisplayData = NewSegmentDisplayData(text)
	case segmentDisplayTime:
		segmentText = "   " + time.Now().Format("150405")
		text = PadRight(segmentText, 9, 0) + PadLeft(trackText, 3)
		segmentDisplayData = NewSegmentDisplayDataTime(text)
	}

	h.controller.writer.SysEx(h.controller.CreateSegmentDisplayData(segmentDisplayData))
}

func (h *EventHandler) OnTick() {
	if h.segmentDisplayMode == segmentDisplayTime {
		h.UpdateDisplay()
	}
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
	case NotePrevious:
		if h.player != nil {
			h.player.Previous()
			h.player.Play()
		}
	case NoteNext:
		if h.player != nil {
			h.player.Next()
			h.player.Play()
		}
	case NoteStop:
		if h.player != nil {
			h.player.Stop()
		}
	case NotePlay:
		if h.player != nil {
			h.player.PlayPause()
		}
	case NoteEncoder:
		h.displayMode = (h.displayMode + 1) % 4
		h.ResetDisplayScroll()
		h.UpdateDisplay()
	case NoteBankLeft:
		h.monitor.SelectPlayer(-1)
	case NoteBankRight:
		h.monitor.SelectPlayer(+1)
	case NoteFader:
		h.mixer.SetOnVolumeChangeCallback(nil)
	case NoteTime:
		if h.segmentDisplayMode == segmentDisplayPlayer {
			h.segmentDisplayMode = segmentDisplayTime
			h.controller.writer.NoteOn(NoteTime, 127)
		} else {
			h.segmentDisplayMode = segmentDisplayPlayer
			h.controller.writer.NoteOn(NoteTime, 0)
		}

		h.UpdateDisplay()
	}
}

func (h *EventHandler) handleNoteOff(note *channel.NoteOff) {
	switch note.Key() {
	case NoteFader:
		h.mixer.SetOnVolumeChangeCallback(h.HandleVolume)
	}
}

func (h *EventHandler) handleControlChange(cc *channel.ControlChange) {
	switch cc.Controller() {
	case CcFader:
		h.mixer.SetVolume(float32(cc.Value()) / 127)
	case CcLedRing:
		h.displayScroll = int(cc.Value())
		h.UpdateDisplay()
	}
}

func (h *EventHandler) ResetDisplayScroll() {
	h.displayScroll = 0
	h.controller.writer.ControlChange(CcLedRing, 0)
}

func (h *EventHandler) HandleVolume(volume float32) {
	h.controller.writer.ControlChange(CcFader, uint8(volume*127))
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
