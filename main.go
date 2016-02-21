package main

import (
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xwindow"
	"log"
	"time"
	"zedsdead/clients"
	"zedsdead/events"
	"zedsdead/utils"
)

const DISPLAY = ":1"

func main() {
	var err error

	X, err := xgbutil.NewConnDisplay(DISPLAY)
	if err != nil {
		log.Fatal("Error connecting to X.", err)
	}
	defer X.Conn().Close()

	if err := utils.Own(X); err != nil {
		log.Fatal(err)
	}

	rootWindow := xwindow.New(X, X.RootWin())
	events.Listen(X, rootWindow)

	pingBefore, pingAfter, pingQuit := xevent.MainPing(X)

	go func() {
		time.Sleep(time.Second * 60)
		close(pingQuit)
	}()

	chan_ticker := events.Tick()

EVENT_LOOP:
	for {
		select {
		case <-pingBefore:
			// Wait for the event to finish processing.
			<-pingAfter
		case <-chan_ticker:
			log.Print(".")
		case <-pingQuit:
			break EVENT_LOOP
		}
	}

	log.Println("Clients: ", len(clients.Clients()))
}
