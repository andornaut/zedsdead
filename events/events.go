package events

import (
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xprop"
	"github.com/BurntSushi/xgbutil/xwindow"
	"log"
	"time"
	"zedsdead/clients"
)

/*
enter_event
destroy_event
button_press_event
key_press_event
map_event
configure_event
unmap_event
client_message_event
*/

func Listen(X *xgbutil.XUtil, rootWindow *xwindow.Window) {

	err := rootWindow.Listen(
		xproto.EventMaskPropertyChange |
			xproto.EventMaskFocusChange |
			xproto.EventMaskButtonPress |
			xproto.EventMaskButtonRelease |
			xproto.EventMaskStructureNotify |
			xproto.EventMaskSubstructureNotify |
			xproto.EventMaskSubstructureRedirect |
			xproto.EventMaskVisibilityChange)
	if err != nil {
		log.Fatal("Could not listen to root window events:", err)
	}

	xevent.MapRequestFun(func(X *xgbutil.XUtil, ev xevent.MapRequestEvent) {
		log.Print("Map: ", ev.Window)
		clients.New(X, ev.Window)
	}).Connect(X, rootWindow.Id)

	xevent.ClientMessageFun(
		func(X *xgbutil.XUtil, ev xevent.ClientMessageEvent) {
			val, err := xprop.AtomName(X, ev.Type)
			if err != nil {
				log.Fatal(err)
			}
			log.Print("Client msg: ", val)
		}).Connect(X, X.RootWin())

}

func Tick() chan bool {
	ch := make(chan bool, 0)
	go func() {
		defer close(ch)

		for {
			ch <- true
			time.Sleep(time.Second)
		}
	}()
	return ch
}
