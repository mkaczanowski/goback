package step

import (
	"strings"
	"time"
)

const STEP_Timeout time.Duration = 5 * time.Second

type StepStatus struct {
	Triggered bool
}

type Step struct {
	Trigger   string // If not set, OnTrigger will run OnTrigger automatically
	Expect    string
	Msg       string
	Next      *Step
	OnTrigger func()
	Status    StepStatus
	SendProbe bool
	Timeout   time.Duration
}

type StepConfig struct {
	ServerAddr    string
	ImageFilename string
	UbootTar      string
	Device        string
	IpAddr        string
	MacAddr       string
}

type StepList struct {
	Length int
	Head   *Step
	Tail   *Step
}

func (st *Step) CheckTrigger(line string) {
	triggered := st.Status.Triggered
	// Run OnTrigger, when no condition is set
	if !triggered {
		if st.Trigger == "" {
			st.OnTrigger()
			st.Status.Triggered = true
		} else {
			if strings.Contains(line, st.Trigger) {
				if st.OnTrigger != nil {
					st.OnTrigger()
				}

				st.Status.Triggered = true
			}
		}
	}
}

func (st *Step) GetTimeout() time.Duration {
	if st.Timeout == 0 {
		return STEP_Timeout
	}
	return st.Timeout
}

func (st *Step) CheckExpect(line string) bool {
	if strings.Contains(line, st.Expect) {
		return true
	}

	return false
}

func NewList() *StepList {
	l := new(StepList)
	l.Length = 0
	return l
}

func (l *StepList) Len() int {
	return l.Length
}

func (l *StepList) IsEmpty() bool {
	return l.Length == 0
}

func (l *StepList) Append(step *Step) {
	if l.Len() == 0 {
		l.Head = step
		l.Tail = l.Head
	} else {
		formerTail := l.Tail
		formerTail.Next = step

		l.Tail = step
	}

	l.Length++
}
