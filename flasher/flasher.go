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
)

type Board uint

const (
	ODROID     Board = 1
	PARALLELLA Board = 2
	WANDBOARD  Board = 3
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

func (fl *Flasher) FlashBoard(b Board, n uint) (quit chan bool, out chan string) {
	quit = make(chan bool)
	out = make(chan string)

	switch b {
	case ODROID:
		go fl.flashOdroid(quit, out, n)
	case WANDBOARD:
		go fl.flashWandboard(quit, out, n)
	case PARALLELLA:
		go fl.flashParallella(quit, out, n)
	}
	return quit, out
}

func (fl *Flasher) flashOdroid(quit chan bool, out chan string, n uint) {
	tty := fl.ttyConfig.ODROID

	if n > uint(len(tty)) {
		out <- "[ Err ] Invalid board number"
		quit <- true
		return
	}

	conf := &step.StepConfig{
		ServerAddr:    "192.168.4.2",
		ImageFilename: "odroid/odroid.img.xz",
		Device:        "/dev/mmcblk0",
		UbootTar:      "odroid/odroid_uboot.tar.gz",
		IpAddr:        "192.168.4.4" + strconv.Itoa((int(n))),
		MacAddr:       "00:10:75:2A:AE:C" + strconv.Itoa((int(n))),
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

func (fl *Flasher) flashWandboard(quit chan bool, out chan string, n uint) {
	tty := fl.ttyConfig.WANDBOARD

	if n > uint(len(tty)) {
		out <- "[ Err ] Invalid board number"
		quit <- true
		return
	}

	conf := &step.StepConfig{
		ServerAddr:    "192.168.4.2",
		ImageFilename: "wand/wandboard.img.xz",
		Device:        "/dev/mmcblk0",
		IpAddr:        "192.168.4.5" + strconv.Itoa((int(n))),
		MacAddr:       "00:10:75:2A:AE:D" + strconv.Itoa((int(n))),
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

func (fl *Flasher) flashParallella(quit chan bool, out chan string, n uint) {
	tty := fl.ttyConfig.PARALLELLA

	if n > uint(len(tty)) {
		out <- "[ Err ] Invalid board number"
		quit <- true
		return
	}

	conf := &step.StepConfig{
		ServerAddr:    "192.168.4.2",
		ImageFilename: "parallella/parallella.img.xz",
		Device:        "/dev/mmcblk0",
		IpAddr:        "192.168.4.6" + strconv.Itoa((int(n))),
		MacAddr:       "00:10:75:2A:AE:E" + strconv.Itoa((int(n))),
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
