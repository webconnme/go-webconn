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
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"runtime"
)

type Client struct {
	url string

	httpClient *http.Client

	done    chan bool
	running bool

	channelMap map[string]chan []byte
	handlerMap map[string]RecvHandler
}

func (client *Client) getMessages() ([]Message, error) {
	var messages []Message

	resp, err := client.httpClient.Get(client.url)
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

func (client *Client) receiver() error {
	for client.running {
		messages, err := client.getMessages()
		if err != nil {
			continue
		}
		for _, m := range messages {
			if h, ok := client.handlerMap[m.Command]; ok {
				err := h([]byte(m.Data))
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (client *Client) postMessages(cmd string, ch <-chan []byte) (int, error) {
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
				resp, err := client.httpClient.Post(client.url, "application/json", r)
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

func (client *Client) sender() error {
	for client.running {
		count := 0
		for k, v := range client.channelMap {
			c, _ := client.postMessages(k, v)
			count += c
		}
		if count == 0 {
			runtime.Gosched()
		}
	}

	return nil
}

func NewClient(url string) *Client {
	client := new(Client)

	client.url = url

	client.done = make(chan bool, 1)
	client.handlerMap = make(map[string]RecvHandler)
	client.channelMap = make(map[string]chan []byte)

	client.httpClient = new(http.Client)

	return client
}

func (client *Client) Stop() {
	client.running = false
	client.done <- true
}

func (client *Client) Register() error {
	return nil
}

func (client *Client) AddHandler(cmd string, handler RecvHandler) {
	client.handlerMap[cmd] = handler
}

func (client *Client) Write(cmd string, data []byte) {
	c, ok := client.channelMap[cmd]
	if !ok {
		client.channelMap[cmd] = make(chan []byte, 100)
		c, ok = client.channelMap[cmd]
	}

	c <- data
}

func (client *Client) Run() error {
	client.running = true

	go client.receiver()
	go client.sender()

	<-client.done
	return nil
}
