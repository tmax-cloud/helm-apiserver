package internal

import (
	"net/http"
	"os"

	gsocket "github.com/gorilla/websocket"
	"k8s.io/klog"
)

// FileExists checks if the file exists in path
func FileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func UpgradeWebsocket(res http.ResponseWriter, req *http.Request) (*gsocket.Conn, error) {
	upgrader := gsocket.Upgrader{
		// TODO : FIX ME for specific domain
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	c, err := upgrader.Upgrade(res, req, nil)
	if err != nil {
		klog.Errorln(err)
		return nil, err
	}
	return c, err
}
