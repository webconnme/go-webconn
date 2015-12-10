package gpio

import (
	"os"
	"fmt"
	"io/ioutil"
	"strconv"
)

const (
	IN = "in"
	OUT = "out"
)

const (
	LOW = 0
	HIGH = 1
)

type Gpio struct {
	Pin int
	Dir string
}

func (g *Gpio) Open() error {
	driverFile := fmt.Sprintf("/sys/class/gpio/gpiochip%d", g.Pin)
	_, err := os.Stat(driverFile)

	if err != nil {
		return err
	}

	err = ioutil.WriteFile("/sys/class/gpio/export", []byte(strconv.Itoa(g.Pin)), 0200)
	if err != nil {
		return err
	}

	dirFile := fmt.Sprintf("/sys/class/gpio/gpio%d/direction", g.Pin)
	_, err = os.Stat(dirFile)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(dirFile, []byte(g.Dir), 0644)
	if err != nil {
		return err
	}
	return nil
}

func (g *Gpio) Close() error {
	err := ioutil.WriteFile("/sys/class/gpio/unexport", []byte(strconv.Itoa(g.Pin)), 0200)
	if err != nil {
		return err
	}
	return nil
}

func (g *Gpio) Out(value int) error {
	valueFile := fmt.Sprintf("/sys/class/gpio/gpio%d/value", g.Pin)
	_, err := os.Stat(valueFile)
	if err != nil {
		return err
	}

	var v []byte
	if value == 0 {
		v = []byte("0")
	} else {
		v = []byte("1")
	}
	err = ioutil.WriteFile(valueFile, []byte(v), 0644)
	if err != nil {
		return err
	}
	
	return nil
}

func (g *Gpio) In() (int, error) {
	valueFile := fmt.Sprintf("/sys/class/gpio/gpio%d/value", g.Pin)
	_, err := os.Stat(valueFile)
	if err != nil {
		return -1, err
	}

	buf, err := ioutil.ReadFile(valueFile)
	if err != nil {
		return -1, err
	}

	i, err := strconv.Atoi(string(buf))
	if err != nil {
		return -1, err
	}

	return i, nil
}