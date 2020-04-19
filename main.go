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

	midiController := MidiController{driver: drv, name: "X-Touch One"}
	must(midiController.OpenIn())
	must(midiController.OpenOut())
	defer midiController.Close()

	must(midiController.Reset())
	defer midiController.Reset()

	sessionBus, err := dbus.SessionBus()
	must(err)
	defer sessionBus.Close()

	playerMonitor := DbusMediaPlayerMonitor{bus: sessionBus}
	must(playerMonitor.Init())

	audioMixer := AudioMixer{}
	must(audioMixer.Init())

	eventHandler := EventHandler{
		controller: &midiController,
		monitor:    &playerMonitor,
		mixer:      &audioMixer,
	}

	eventHandler.Setup()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	<-signals
}

func must(err error) {
	if err != nil {
		panic(err.Error())
	}
}
