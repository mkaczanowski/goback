package main

import (
	"./flasher"
	"./job"
	"./key"
	"flag"
	"log"
	"strconv"
	"strings"
	"sync"
)

var wg sync.WaitGroup
var fl *flasher.Flasher

func flashDevice(j *job.Job) {
	b, no := key.Decrypt(j.Id)
	board := flasher.Board(b)
	prefix := "[T:" + flasher.GetBoardName(board) + " N:" + strconv.Itoa(int(no)) + "]"

	defer func() {
		log.Println(prefix, "[ Job ] Cleanning up")
		j.Cleanup()
		wg.Done()
	}()

	quit, out := fl.FlashBoard(board, no)

	log.Println(prefix, "[ Job ] Start")
forever:
	for {
		select {
		case msg := <-out:
			j.LastMsg = msg
			log.Println(prefix, j.LastMsg)
		case <-quit:
			log.Println(prefix, "[ Job ] Flasher finished by itself")
			j.Sch.RemoveJob(j)
			break forever
		case <-j.Exit:
			quit <- true
			j.Sch.RemoveJob(j)
			log.Println(prefix, "[ Job ] Terminated from console")
			break forever
		}
	}
}

func getKey(board string) uint {
	var k uint
	num := board[len(board)-1:]
	no, err := strconv.Atoi(num)
	if err != nil {
		log.Fatal("Board no is not provided: ", board)
	}

	if strings.HasPrefix(board, "odroid") {
		k = key.Encrypt(uint(flasher.ODROID), uint(no))
	} else if strings.HasPrefix(board, "wandboard") {
		k = key.Encrypt(uint(flasher.WANDBOARD), uint(no))
	} else if strings.HasPrefix(board, "parallella") {
		k = key.Encrypt(uint(flasher.PARALLELLA), uint(no))
	}

	return k
}

func main() {
	log.SetFlags(0)
	fBoards := flag.String("boards", "", "List boards to flash")
	flag.Parse()

	fl = flasher.NewFlasher("tty.json")
	sch := job.NewJobScheduler()

	boards := strings.Split(*fBoards, ",")
	for i := range boards {
		key := getKey(boards[i])
		j := job.NewJob(key, sch)
		j.SetHandler(flashDevice)
		sch.StartJob(j)
		wg.Add(1)
	}

	wg.Wait()
}
