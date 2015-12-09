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

import ()

/*
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

func (w *Webconn) postMessages(cmd string, ch <-chan []byte) (int, error) {
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
                        return 0, err
                    }
                    r := bytes.NewReader(json_data)
                    resp, err := w.client.Post(w.url, "application/json", r)
                    if err != nil {
                        return 0, err
                    }
                    defer resp.Body.Close()

                    _, err = ioutil.ReadAll(resp.Body)
                    if err != nil {
                        return 0, err
                    }
                    return len(json_data), nil
                } else {
                    return 0, nil
                }
        }
    }

    return 0, nil
}

func (w *Webconn) sender() error {
	for w.running {
        count := 0
        for k, v := range w.channelMap {
            c, _ := w.postMessages(k, v)
            count += c
        }
        if count == 0 {
            runtime.Gosched()
        }
    }

    return nil
}


*/

type Server struct {
	ip   string
	port int

	done    chan bool
	running bool

	channelMap map[string]chan []byte
	handlerMap map[string]RecvHandler
}

func NewServer(ip string, port int) *Server {
	server := new(Server)

	server.ip = ip
	server.port = port

	server.done = make(chan bool, 1)
	server.handlerMap = make(map[string]RecvHandler)
	server.channelMap = make(map[string]chan []byte)

	return server
}

func (server *Server) Register() error {
	return nil
}

func (server *Server) Write(cmd string, data []byte) {
	c, ok := server.channelMap[cmd]
	if !ok {
		server.channelMap[cmd] = make(chan []byte, 100)
		c, ok = server.channelMap[cmd]
	}

	c <- data
}

func (server *Server) AddHandler(cmd string, handler RecvHandler) {
	server.handlerMap[cmd] = handler
}

func (server *Server) Run() error {
	// Do server work...

	<-server.done
	return nil
}

func (server *Server) Stop() {
	server.running = false
	server.done <- true
}
