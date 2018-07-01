package job

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/wxio/job/api"
)

type JobSvr struct {
}
type LogSvr struct {
}

var _ api.JobServer = &JobSvr{}
var _ api.LogServer = &LogSvr{}

type job struct {
	ID        uint32
	Status    api.JobStatus
	Logs      []string
	logClient chan string
}

var (
	jobs = make([]*job, 0)
	jmu  sync.Mutex
)

func (ps *JobSvr) Init(ctx context.Context, req *api.InitReq) (*api.InitResp, error) {
	jb := &job{}
	jmu.Lock()
	defer jmu.Unlock()
	jobs = append(jobs, jb)
	jb.ID = uint32(len(jobs))
	return &api.InitResp{Id: jb.ID}, nil
}
func (ps *JobSvr) Run(ctx context.Context, req *api.RunReq) (*api.RunResp, error) {
	if req.Id > uint32(len(jobs)) {
		return nil, errors.New("job id is out of bounds")
	}
	jb := jobs[req.Id-1]
	go jb.run()
	jb.Status = api.JobStatus_running
	return &api.RunResp{}, nil
}

func (ls *LogSvr) Get(ctx context.Context, req *api.LogReq) (*api.LogResp, error) {
	if req.Id > uint32(len(jobs)) {
		return nil, errors.New("job id is out of bounds")
	}
	jb := jobs[req.Id-1]
	resp := &api.LogResp{
		Lines:  jb.Logs,
		Status: jb.Status,
	}
	return resp, nil
}
func (ls *LogSvr) GetStream(req *api.LogStreamReq, resp api.Log_GetStreamServer) error {
	if req.Id > uint32(len(jobs)) {
		return errors.New("job id is out of bounds")
	}
	jb := jobs[req.Id-1]
	jmu.Lock()
	if jb.logClient != nil {
		return errors.New("client already listening to log")
	}
	jb.logClient = make(chan string, 0)
	jmu.Unlock()
	defer func() {
		jb.logClient = nil
	}()
	for line := range jb.logClient {
		err := resp.Send(&api.LogStreamResp{Line: line})
		if err != nil {
			return err
		}
	}
	return nil
}

func (jb *job) run() {
	i := 0
	for {
		line := fmt.Sprintf("At the beep the time will be %v, ... beep", time.Now())
		log.Printf("%d: %s", jb.ID, line)
		jb.Logs = append(jb.Logs, line)
		if jb.logClient != nil {
			jb.logClient <- line
		}
		// select {
		// case jb.logClient <- line:
		// default:
		// }
		<-time.After(1 * time.Second)
		i++
		if i >= 4 {
			break
		}
	}
	if jb.logClient != nil {
		close(jb.logClient)
	}
	jb.Status = api.JobStatus_finished
}
