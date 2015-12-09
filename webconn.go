/**
 * The MIT License (MIT)
 * 
 * Copyright (c) 2015 Edward Kim <edward@webconn.me>
 * 
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 * 
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 * 
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

package webconn

import (
	"errors"
	"net/http"
	"io/ioutil"
	"bytes"
	"runtime"
    "encoding/json"
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


    channelMap map[string]chan []byte
	handlerMap map[string]RecvHandler
}

type Message struct {
    Command string `json:"command"`
    Data string `json:"data"`
}


func (w *Webconn) getMessages() ([]Message, error) {
    var messages []Message

    resp, err := w.client.Get(w.url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    content, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }

    err = json.Unmarshal(content, &messages)
    if err != nil {
        return nil, err
    }

    return messages, nil
}

func (w *Webconn) receiver() error {
	for w.running {
        messages, err := w.getMessages()
        if err != nil {
            continue
        }
        for _, m := range messages {
            if h, ok := w.handlerMap[m.Command]; ok {
                err := h([]byte(m.Data))
                if err != nil {
                    return err
                }
            }
        }
	}
	return nil
}

func (w *Webconn) postMessages(cmd string, ch <-chan []byte) (bool, error) {
    buf := []byte{}
    for {
        select {
            case <-ch:
                b := <-ch
                buf = append(buf, b...)
            default:
                if len(buf) > 0 {
                    msg := []Message{Message{cmd, string(buf)}}
                    json_data, err := json.Marshal(msg)
                    if err != nil {
                        return false, err
                    }
                    r := bytes.NewReader(json_data)
                    resp, err := w.client.Post(w.url, "application/json", r)
                    if err != nil {
                        return false, err
                    }
                    defer resp.Body.Close()

                    _, err = ioutil.ReadAll(resp.Body)
                    if err != nil {
                        return false, err
                    }
                    return true, nil
                } else {
                    return false, nil
                }
        }
    }

    return false, nil
}

func (w *Webconn) sender() error {
	for w.running {
        for k, v := range w.channelMap {
            w.postMessages(k, v)
        }
        runtime.Gosched()
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
    w.handlerMap = make(map[string]RecvHandler)
	w.channelMap = make(map[string]chan []byte)
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

func (w *Webconn) AddHandler(cmd string, handler RecvHandler) {
    w.handlerMap[cmd] = handler
}

func (w *Webconn) Write(cmd string, data []byte) {
    c, ok := w.channelMap[cmd]
    if !ok {
        w.channelMap[cmd] = make(chan []byte, 100)
        c, ok = w.channelMap[cmd]
    }

    c <- data
}

func (w *Webconn) runClient() error {
	go w.receiver()
	go w.sender()

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