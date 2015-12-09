package webconn

import (
	"errors"
	"net/http"
	"io/ioutil"
	"bytes"
	"runtime"
)

const (
	WEBCONN_SERVER = 1
	WEBCONN_CLIENT = 2
)

type RecvHandler func([]byte) error

type Webconn struct {
	ip string
	port int

	mode int
	url string

	client *http.Client

	done chan bool
	running bool


	writeChan chan []byte
	handler RecvHandler
}



func (w *Webconn) getHttp() error {
	for w.running {
		resp, err := w.client.Get(w.url)

		if err != nil {
			return err
		}
		defer resp.Body.Close()

		content, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		if w.handler != nil {
			return w.handler(content)
		}
	}
	return nil
}

func getReader(list [][]byte) *bytes.Reader {
	b := bytes.Join(list, []byte(","))
	l := make([][]byte, 3)
	l[0] = []byte("[")
	l[1] = b
	l[2] = []byte("[")

	res := bytes.Join(l, []byte(""))
	return bytes.NewReader(res)
}

func (w *Webconn) postHttp() error {
	list := make([][]byte, 0)

	for w.running {
		select {
			case <-w.writeChan:
				buf := <-w.writeChan
				list = append(list, buf)
			default:
				if len(list) > 0 {
					r := getReader(list)
					resp, err := w.client.Post(w.url, "application/json", r)
					if err != nil {
						return err
					}
					defer resp.Body.Close()

					_, err = ioutil.ReadAll(resp.Body)
					if err != nil {
						return err
					}
					list = make([][]byte, 0)
				} else {
					runtime.Gosched()
				}
		}
	}

	return nil
}

func Client(url string) *Webconn {
	w := new(Webconn)

	w.mode = WEBCONN_CLIENT
	w.url = url

	w.init()

	return w
}

func Server(ip string, port int) *Webconn {
	w := new(Webconn)

	w.mode = WEBCONN_SERVER
	w.ip = ip
	w.port = port

	w.init()

	return w
}

func (w *Webconn) init() {
	w.done = make(chan bool, 1)
	w.writeChan = make(chan []byte, 100)
}

func (w *Webconn) Stop() {
	w.done <- true
}

func (w *Webconn) Register() error {
	if w.mode == WEBCONN_SERVER {
		return errors.New("Unsupported action")
	}

	return nil
}

func (w *Webconn) AddHandler(handler RecvHandler) {
	w.handler = handler
}

func (w *Webconn) runClient() error {
	go w.getHttp()
	go w.postHttp()

	<-w.done
	return nil
}

func (w *Webconn) runServer() error {
	w.running = false
	<-w.done
	return nil
}

func (w *Webconn) Run() error {
	w.running = true

	switch (w.mode) {
	case WEBCONN_SERVER:
		return w.runServer()
	case WEBCONN_CLIENT:
		return w.runClient()
	}
	return errors.New("Unsupported mode")
}