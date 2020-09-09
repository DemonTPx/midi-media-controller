package main

import (
	"github.com/godbus/dbus"
	"gitlab.com/gomidi/rtmididrv"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	drv, err := rtmididrv.New()
	must(err)
	defer drv.Close()

	midiController := NewMidiController(drv, "X-Touch One")
	must(midiController.OpenIn())
	must(midiController.OpenOut())
	defer midiController.Close()

	must(midiController.Reset())
	defer midiController.Reset()

	sessionBus, err := dbus.SessionBus()
	must(err)
	defer sessionBus.Close()

	playerMonitor := NewDbusMediaPlayerMonitor(sessionBus)
	must(playerMonitor.Init())

	audioMixer := NewAudioMixer()
	must(audioMixer.Init())

	eventHandler := NewEventHandler(midiController, playerMonitor, audioMixer)

	eventHandler.Setup()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

	<-signals
}

func must(err error) {
	if err != nil {
		panic(err.Error())
	}
}
