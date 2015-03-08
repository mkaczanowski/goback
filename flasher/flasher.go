package flasher

import (
	"../config"
	"../power"
	"../step"
	"../util"
	"bufio"
	"encoding/json"
	"github.com/tarm/goserial"
	"os"
	"strconv"
	"time"
	"errors"
)

type System uint
type Board uint
var Debug bool

const (
	ODROID     Board = 1
	PARALLELLA Board = 2
	WANDBOARD  Board = 3
)

const (
	UBUNTU System = 1
	FEDORA System = 2
)

type ttyConfiguration struct {
	ODROID     []string
	WANDBOARD  []string
	PARALLELLA []string
}

type Flasher struct {
	ttyConfig *ttyConfiguration
}

func NewFlasher(ttyConfig string) *Flasher {
	c := getTTYConfig(ttyConfig)
	fl := &Flasher{ttyConfig: c}

	return fl
}

func GetSystemName(sys System) string {
	switch sys {
	case UBUNTU:
		return "ubuntu"
	case FEDORA:
		return "fedora"
	}

	return "Invalid sys type provided"
}

func GetSystem(sys string) (System, error) {
	switch sys {
	case "ubuntu":
		return UBUNTU, nil
	case "fedora":
		return FEDORA, nil
	}

	return System(0), errors.New("Invalid system name")
}

func GetBoardName(no Board) string {
	switch no {
	case ODROID:
		return "ODROID"
	case PARALLELLA:
		return "PARALLELLA"
	case WANDBOARD:
		return "WANDBOARD"
	}

	return "Invalid board type provided"
}

func (fl *Flasher) FlashBoard(b Board, n uint, sys System) (quit chan bool, out chan string) {
	quit = make(chan bool)
	out = make(chan string)

	if Debug == true {
		util.Debug = true
	}

	switch b {
	case ODROID:
		go fl.flashOdroid(quit, out, n, sys)
	case WANDBOARD:
		go fl.flashWandboard(quit, out, n, sys)
	case PARALLELLA:
		go fl.flashParallella(quit, out, n, sys)
	}
	return quit, out
}

func (fl *Flasher) flashOdroid(quit chan bool, out chan string, n uint, sys System) {
	tty := fl.ttyConfig.ODROID

	if n > uint(len(tty)) {
		out <- "[ Err ] Invalid board number"
		quit <- true
		return
	}

	conf := &step.StepConfig{
		ServerAddr:    "192.168.4.2",
		ImageFilename: "odroid/"+GetSystemName(sys)+".img.xz",
		Device:        "/dev/mmcblk0",
		UbootTar:      "odroid/odroid_uboot.tar.gz",
		IpAddr:        "192.168.4.4" + strconv.Itoa((int(n))),
		MacAddr:       "00:10:75:2A:AE:C" + strconv.Itoa((int(n))),
		Hostname:      "odroid"+strconv.Itoa((int(n))),
	}

	// Initialize serial line
	c := &serial.Config{Name: tty[n], Baud: 115200}
	s, err := serial.OpenPort(c)
	if err != nil {
		out <- "[ Err ] " + err.Error()
		quit <- true

		return
	}
	defer func() {
		out <- "[ Close ] Serial connection"
		s.Close()
		quit <- true
	}()

	boardName := "odroid" + strconv.Itoa(int(n))
	writer := bufio.NewWriter(s)
	reader := bufio.NewReader(s)

	// Initialize actions as a list of step
	stepList := config.GetOdroidStepList(writer, conf)

	// Redirect serial to chanel and setup timeout
	// start is blocking
	start(boardName, out, stepList, reader, writer)
}

func (fl *Flasher) flashWandboard(quit chan bool, out chan string, n uint, sys System) {
	tty := fl.ttyConfig.WANDBOARD

	if n > uint(len(tty)) {
		out <- "[ Err ] Invalid board number"
		quit <- true
		return
	}

	conf := &step.StepConfig{
		ServerAddr:    "192.168.4.2",
		ImageFilename: "wand/"+GetSystemName(sys)+".img.xz",
		Device:        "/dev/mmcblk0",
		IpAddr:        "192.168.4.5" + strconv.Itoa((int(n))),
		MacAddr:       "00:10:75:2A:AE:D" + strconv.Itoa((int(n))),
		Hostname:      "wandboard"+strconv.Itoa((int(n))),
	}

	// Initialize serial line
	c := &serial.Config{Name: tty[n], Baud: 115200}
	s, err := serial.OpenPort(c)
	if err != nil {
		out <- "[ Err ] " + err.Error()
		quit <- true

		return
	}
	defer func() {
		out <- "[ Close ] Serial connection"
		s.Close()
		quit <- true
	}()

	boardName := "wandboard" + strconv.Itoa(int(n))
	writer := bufio.NewWriter(s)
	reader := bufio.NewReader(s)

	// Initialize actions as a list of step
	stepList := config.GetWandboardStepList(writer, conf)

	// Redirect serial to chanel and setup timeout
	// start is blocking
	start(boardName, out, stepList, reader, writer)
}

func (fl *Flasher) flashParallella(quit chan bool, out chan string, n uint, sys System) {
	tty := fl.ttyConfig.PARALLELLA

	if n > uint(len(tty)) {
		out <- "[ Err ] Invalid board number"
		quit <- true
		return
	}

	conf := &step.StepConfig{
		ServerAddr:    "192.168.4.2",
		ImageFilename: "parallella/"+GetSystemName(sys)+".img.xz",
		Device:        "/dev/mmcblk0",
		IpAddr:        "192.168.4.6" + strconv.Itoa((int(n))),
		MacAddr:       "00:10:75:2A:AE:E" + strconv.Itoa((int(n))),
		Hostname:      "parallella"+strconv.Itoa((int(n))),
	}

	// Initialize serial line
	c := &serial.Config{Name: tty[n], Baud: 115200}
	s, err := serial.OpenPort(c)
	if err != nil {
		out <- "[ Err ] " + err.Error()
		quit <- true

		return
	}
	defer func() {
		out <- "[ Close ] Serial connection"
		s.Close()
		quit <- true
	}()

	boardName := "parallella" + strconv.Itoa(int(n))
	writer := bufio.NewWriter(s)
	reader := bufio.NewReader(s)

	// Initialize actions as a list of step
	stepList := config.GetParallellaStepList(writer, conf)

	// Redirect serial to chanel and setup timeout
	// start is blocking
	start(boardName, out, stepList, reader, writer)
}

func start(boardName string, out chan string, sl *step.StepList,
	r *bufio.Reader, w *bufio.Writer) {

	curStep := sl.Head
	serialOutput := make(chan string)
	quit := make(chan bool)
	timer := time.AfterFunc(curStep.GetTimeout(), func() { quit <- true })
	go monitor(serialOutput, r)

	// Try to restart
	out <- "[ Pre ] Restart machine"
	power.Switch("off", boardName)
	power.Switch("on", boardName)

loop:
	for {
		select {
		case line, ok := <-serialOutput:
			if !ok {
				break loop
			}

			if Debug {
				out <- line
			}

			curStep.CheckTrigger(line)

			if curStep.Status.Triggered && curStep.CheckExpect(line) {
				out <- "[ OK ] " + curStep.Msg
				if curStep.Next != nil {
					curStep = curStep.Next
					out <- "[ Run ] " + curStep.Msg
				} else {
					out <- "[ STOP ] No more steps"
					break loop
				}

				if curStep.SendProbe {
					util.MustSendCmd(w, "\n", true)
				}
			}

			timer.Reset(curStep.GetTimeout())
		case <-quit:
			break loop
		}
	}
}

func monitor(ch chan string, r *bufio.Reader) {
	defer close(ch)
	for {
		line, _, err := r.ReadLine()
		if len(line) > 0 {
			ch <- string(line)
		}
		if err != nil {
			break
		}
	}
}

func getTTYConfig(path string) *ttyConfiguration {
	file, _ := os.Open(path)
	decoder := json.NewDecoder(file)
	configuration := &ttyConfiguration{}
	err := decoder.Decode(configuration)
	if err != nil {
		panic("Couldn't load TTY config: " + err.Error())
	}

	return configuration
}
