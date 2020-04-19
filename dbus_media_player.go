package main

import (
	"github.com/godbus/dbus"
	"strings"
)

const (
	mprisPath       = "/org/mpris/MediaPlayer2"
	mprisName       = "org.mpris.MediaPlayer2"
	mprisPlayerName = "org.mpris.MediaPlayer2.Player"

	propertiesGet = "org.freedesktop.DBus.Properties.Get"

	stop      = mprisPlayerName + ".Stop"
	play      = mprisPlayerName + ".Play"
	playPause = mprisPlayerName + ".PlayPause"
	previous  = mprisPlayerName + ".Previous"
	next      = mprisPlayerName + ".Next"
)

type DbusMediaPlayer struct {
	bus                       *dbus.Conn
	busName                   string
	owner                     string
	name                      string
	nameLower                 string
	mprisObj                  dbus.BusObject
	playbackStatus            string
	track                     Track
	propertiesChangedCallback func(playbackStatus string, track Track)
}

func (p *DbusMediaPlayer) Init() {
	p.mprisObj = p.bus.Object(p.busName, mprisPath)

	p.name = p.busName

	var identity string
	p.mprisObj.Call(propertiesGet, 0, mprisName, "Identity").Store(&identity)

	if len(identity) != 0 {
		p.name = identity
	}
	p.nameLower = strings.ToLower(p.name)

	p.mprisObj.AddMatchSignal("org.freedesktop.DBus.Properties", "PropertiesChanged")
}

func (p *DbusMediaPlayer) Close() {
	p.mprisObj.RemoveMatchSignal("org.freedesktop.DBus.Properties", "PropertiesChanged")
}

func (p *DbusMediaPlayer) Stop() {
	p.mprisObj.Call(stop, 0).Store()
}

func (p *DbusMediaPlayer) Play() {
	p.mprisObj.Call(play, 0).Store()
}

func (p *DbusMediaPlayer) PlayPause() {
	p.mprisObj.Call(playPause, 0).Store()
}

func (p *DbusMediaPlayer) Previous() {
	p.mprisObj.Call(previous, 0).Store()
}

func (p *DbusMediaPlayer) Next() {
	p.mprisObj.Call(next, 0).Store()
}

func (p *DbusMediaPlayer) FetchProperties() (string, Track) {
	p.mprisObj.Call(propertiesGet, 0, mprisPlayerName, "PlaybackStatus").Store(&p.playbackStatus)

	var metadataVariant map[string]dbus.Variant
	p.mprisObj.Call(propertiesGet, 0, mprisPlayerName, "Metadata").Store(&metadataVariant)
	p.track = parseMetadata(metadataVariant)

	return p.playbackStatus, p.track
}

func (p *DbusMediaPlayer) onPropertiesChanged(propertiesVariant map[string]dbus.Variant) {
	if variant, found := propertiesVariant["PlaybackStatus"]; found {
		if val, ok := variant.Value().(string); ok {
			p.playbackStatus = val
		}
	}

	if variant, found := propertiesVariant["Metadata"]; found {
		if metadata, ok := variant.Value().(map[string]dbus.Variant); ok {
			p.track = parseMetadata(metadata)
		}
	}

	if p.propertiesChangedCallback != nil {
		p.propertiesChangedCallback(p.playbackStatus, p.track)
	}
}

func (p *DbusMediaPlayer) SetOnPropertiesChangedHandler(callback func(playbackStatus string, track Track)) {
	p.propertiesChangedCallback = callback
}

func parseMetadata(metadata map[string]dbus.Variant) Track {
	return Track{
		artist:      getMetaFirstOrEmptyString(metadata, "xesam:artist"),
		albumArtist: getMetaFirstOrEmptyString(metadata, "xesam:albumArtist"),
		album:       getMetaOrEmptyString(metadata, "xesam:album"),
		title:       getMetaOrEmptyString(metadata, "xesam:title"),
		trackNumber: getMetaOrZero(metadata, "xesam:trackNumber"),
	}
}

func getMetaOrEmptyString(metadata map[string]dbus.Variant, key string) string {
	if variant, ok := metadata[key]; ok {
		if val, ok := variant.Value().(string); ok {
			return val
		}
	}

	return ""
}

func getMetaOrZero(metadata map[string]dbus.Variant, key string) int {
	if variant, ok := metadata[key]; ok {
		if val, ok := variant.Value().(int32); ok {
			return int(val)
		}
	}

	return 0
}

func getMetaFirstOrEmptyString(metadata map[string]dbus.Variant, key string) string {
	if variant, ok := metadata[key]; ok {
		if val, ok := variant.Value().([]string); ok {
			if len(val) != 0 {
				return val[0]
			}
		}
	}

	return ""
}
