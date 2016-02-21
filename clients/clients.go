package clients

import (
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xwindow"
	"log"
)

var clients []Client = []Client{}

type Client struct {
	Win        *xwindow.Window
	Is_managed bool
}

func (c *Client) Id() xproto.Window {
	return c.Win.Id
}

func New(X *xgbutil.XUtil, winId xproto.Window) *Client {
	log.Println("New window: ", winId)
	X.Grab()
	defer X.Ungrab()

	win := xwindow.New(X, winId)
	if _, err := win.Geometry(); err != nil {
		log.Println("Could not manage client %d because: %s", winId, err)
		return nil
	}

	client := &Client{Win: win, Is_managed: true}
	//clients = append(clients, *client)
	win.MoveResize(0, 0, 600, 600)
	win.Map()

	//client.attachMapNotify().Connect(X, client.Id())
	//client.attachUnmapNotify().Connect(X, client.Id())
	client.attachDestroyNotify(X).Connect(X, client.Id())

	return client
}

func (c *Client) attachDestroyNotify(X *xgbutil.XUtil) xevent.DestroyNotifyFun {
	f := func(X *xgbutil.XUtil, ev xevent.DestroyNotifyEvent) {
		log.Print("Destroy")
		c.Is_managed = false
	}
	return xevent.DestroyNotifyFun(f)
}

func UnMap(winId xproto.Window) {
	for i, client := range clients {
		if winId == client.Win.Id {
			client.Is_managed = false
			// Memory leak?
			clients = append(clients[:i], clients[i+1:]...)
			log.Print("Removed client: ", winId)
			break
		}
	}
}

func Clients() []Client {
	return clients
}
