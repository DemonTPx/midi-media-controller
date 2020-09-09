package main

import (
	"github.com/godbus/dbus"
	"log"
	"strings"
)

const (
	mediaPlayerPrefix = "org.mpris.MediaPlayer2."
	listNames         = "org.freedesktop.DBus.ListNames"
	getNameOwner      = "org.freedesktop.DBus.GetNameOwner"
	nameOwnerChanged  = "org.freedesktop.DBus.NameOwnerChanged"
	propertiesChanged = "org.freedesktop.DBus.Properties.PropertiesChanged"
)

type DbusMediaPlayerMonitor struct {
	bus                         *dbus.Conn
	activePlayer                *string
	playerList                  map[string]*DbusMediaPlayer
	signal                      chan *dbus.Signal
	activePlayerChangedCallback func(player *DbusMediaPlayer)
}

func NewDbusMediaPlayerMonitor(bus *dbus.Conn) *DbusMediaPlayerMonitor {
	return &DbusMediaPlayerMonitor{
		bus: bus,
	}
}

func (m *DbusMediaPlayerMonitor) Init() error {
	m.playerList = make(map[string]*DbusMediaPlayer)

	var names []string
	err := m.bus.BusObject().Call(listNames, 0).Store(&names)

	if err != nil {
		return err
	}

	for _, name := range names {
		if strings.HasPrefix(name, mediaPlayerPrefix) {
			owner := ""
			err := m.bus.BusObject().Call(getNameOwner, 0, name).Store(&owner)

			if err != nil {
				return err
			}

			m.addPlayer(name, owner)
		}
	}

	m.signal = make(chan *dbus.Signal, 10)

	m.bus.AddMatchSignal(dbus.WithMatchMember("NameOwnerChanged"))
	m.bus.Signal(m.signal)

	go func() {
		for signal := range m.signal {
			m.handleSignal(signal)
		}
	}()

	return nil
}

func (m *DbusMediaPlayerMonitor) GetActivePlayer() *DbusMediaPlayer {
	if m.activePlayer == nil {
		return nil
	}

	return m.playerList[*m.activePlayer]
}

func (m *DbusMediaPlayerMonitor) handleSignal(signal *dbus.Signal) {
	switch signal.Name {
	case nameOwnerChanged:
		m.onNameOwnerChanged(signal.Body[0].(string), signal.Body[1].(string), signal.Body[2].(string))
	case propertiesChanged:
		m.onPropertiesChanged(signal.Sender, signal.Body[1].(map[string]dbus.Variant))
	default:
		log.Printf("Received unknown signal: %+v", signal)
	}
}

func (m *DbusMediaPlayerMonitor) onNameOwnerChanged(name string, oldOwner string, newOwner string) {
	if !strings.HasPrefix(name, mediaPlayerPrefix) {
		return
	}

	if len(newOwner) != 0 && len(oldOwner) == 0 {
		m.addPlayer(name, newOwner)
	} else if len(oldOwner) != 0 && len(newOwner) == 0 {
		m.removePlayer(name, oldOwner)
	} else {
		m.playerChangeOwner(name, oldOwner, newOwner)
	}
}

func (m *DbusMediaPlayerMonitor) onPropertiesChanged(sender string, properties map[string]dbus.Variant) {
	player, ok := m.playerList[sender]

	if !ok {
		return
	}

	player.onPropertiesChanged(properties)
}

func (m *DbusMediaPlayerMonitor) addPlayer(name string, ownerName string) {
	log.Printf("Adding new player %s owner %s", name, ownerName)

	player := DbusMediaPlayer{bus: m.bus, busName: name, owner: ownerName}
	player.Init()

	m.playerList[ownerName] = &player

	if m.activePlayer == nil {
		m.activePlayer = &ownerName

		if m.activePlayerChangedCallback != nil {
			m.activePlayerChangedCallback(&player)
		}
	}
}

func (m *DbusMediaPlayerMonitor) removePlayer(name string, ownerName string) {
	log.Printf("Removing player %s owner %s", name, ownerName)

	player, ok := m.playerList[ownerName]

	if !ok || player.busName != name {
		return
	}

	if *m.activePlayer == ownerName {
		m.SelectPlayer(-1)
	}

	player.Close()
	delete(m.playerList, ownerName)

	if *m.activePlayer == ownerName {
		m.activePlayer = nil
		if m.activePlayerChangedCallback != nil {
			m.activePlayerChangedCallback(nil)
		}
	}
}

func (m *DbusMediaPlayerMonitor) playerChangeOwner(name string, oldOwner string, newOwner string) {
	log.Printf("Changing owner of player %s from %s to %s", name, oldOwner, newOwner)

	player, ok := m.playerList[oldOwner]

	if !ok || player.busName != name {
		return
	}

	m.playerList[newOwner] = player
	m.playerList[newOwner].owner = newOwner
	delete(m.playerList, oldOwner)

	if m.activePlayer != nil && *m.activePlayer == oldOwner {
		m.activePlayer = &newOwner
	}
}

func (m *DbusMediaPlayerMonitor) SetActivePlayerChangedCallback(callback func(player *DbusMediaPlayer)) {
	m.activePlayerChangedCallback = callback
}

func (m *DbusMediaPlayerMonitor) SelectPlayer(offset int) {
	if len(m.playerList) < 2 {
		return
	}

	var keyList []string
	current := -1
	index := 0
	for key := range m.playerList {
		keyList = append(keyList, key)
		if key == *m.activePlayer {
			current = index
		}
		index++
	}

	newIndex := (((current + offset) % index) + index) % index
	m.activePlayer = &keyList[newIndex]

	if m.activePlayerChangedCallback != nil {
		m.activePlayerChangedCallback(m.GetActivePlayer())
	}
}
