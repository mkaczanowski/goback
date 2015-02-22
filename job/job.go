package job

import (
	"strconv"
	"sync"
	"time"
)

const (
	JOB_CREATED = 0
	JOB_STARTED = 1
)

type JobFunc func(job *Job)

type Job struct {
	Id      uint
	Status  int
	LastMsg string
	Created time.Time
	Updated time.Time
	Exit    chan int

	Sch     *JobScheduler
	handler JobFunc
}

func NewJob(key uint, sch *JobScheduler) (j *Job) {
	t := time.Now().Local()
	j = &Job{Id: key, Status: JOB_CREATED, Created: t, Updated: t, Sch: sch}
	j.Exit = make(chan int)
	return j
}

func (j *Job) Execute() {
	j.Status = JOB_STARTED
	j.handler(j)
}

func (j *Job) SetHandler(handle JobFunc) {
	j.handler = handle
}

func (j *Job) Cleanup() {
	close(j.Exit)
}

func (j *Job) String() string {
	return "[ Job ] Id: " + strconv.Itoa(int(j.Id)) + " Status: " + strconv.Itoa(j.Status) + " Created: " + j.Created.String()
}

type JobScheduler struct {
	list map[uint]*Job
	m    sync.RWMutex
}

func NewJobScheduler() *JobScheduler {
	sch := new(JobScheduler)
	sch.list = make(map[uint]*Job)

	return sch
}

func (st *JobScheduler) StartJob(j *Job) (r bool) {
	st.m.Lock()
	defer st.m.Unlock()

	if _, ok := st.list[j.Id]; !ok {
		st.list[j.Id] = j
		go j.Execute()
		r = true
	}

	return r
}

func (st *JobScheduler) RemoveJob(j *Job) {
	st.m.Lock()
	delete(st.list, j.Id)
	st.m.Unlock()
}

func (st *JobScheduler) StopJob(j *Job) (ok bool) {
	st.m.RLock()
	if _, ok := st.list[j.Id]; ok {
		j.Exit <- 1
	}
	st.m.RUnlock()

	return ok
}
