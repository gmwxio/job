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

type jobSvc struct {
	run chan struct {
	}
	listen chan struct {
		ctx      context.Context
		listener chan string
	}
	logs chan struct {
		resp chan struct {
			logs   []string
			status api.JobStatus
		}
	}
}

func newJob(id uint32) *jobSvc {
	js := &jobSvc{
		run: make(chan struct{}),
		listen: make(chan struct {
			ctx      context.Context
			listener chan string
		}),
		logs: make(chan struct {
			resp chan struct {
				logs   []string
				status api.JobStatus
			}
		}),
	}
	go func() {
		var (
			status    api.JobStatus
			logs      []string
			listeners []struct {
				ctx      context.Context
				listener chan string
			}
		)
		ticker := (<-chan time.Time)(nil)
		i := 0
		for {
			select {
			case <-ticker:
				line := fmt.Sprintf("At the beep the time will be %v, ... beep", time.Now())
				log.Printf("%d: %s", id, line)
				logs = append(logs, line)
				for _, li := range listeners {
					select {
					case <-li.ctx.Done():
					case li.listener <- line:
					}
				}
				i++
				if i >= 4 {
					ticker = nil
					status = api.JobStatus_finished
					for _, li := range listeners {
						select {
						case <-li.ctx.Done():
						default:
							close(li.listener)
						}
					}
				}
			case <-js.run:
				status = api.JobStatus_running
				ticker = time.Tick(1 * time.Second)
			case li := <-js.listen:
				listeners = append(listeners, li)
			case client := <-js.logs:
				client.resp <- struct {
					logs   []string
					status api.JobStatus
				}{
					logs:   logs,
					status: status,
				}
			}
		}
	}()
	return js
}

var (
	jobs = make([]*jobSvc, 0)
	jmu  sync.Mutex
)

func (ps *JobSvr) Init(ctx context.Context, req *api.InitReq) (*api.InitResp, error) {
	jmu.Lock()
	defer jmu.Unlock()
	id := uint32(len(jobs)) + 1
	jobs = append(jobs, newJob(id))
	return &api.InitResp{Id: id}, nil
}

func (ps *JobSvr) Run(ctx context.Context, req *api.RunReq) (*api.RunResp, error) {
	jmu.Lock()
	if req.Id > uint32(len(jobs)) {
		jmu.Unlock()
		return nil, errors.New("job id is out of bounds")
	}
	jb := jobs[req.Id-1]
	jmu.Unlock()
	fmt.Printf("1--- %v\n", *jb)
	jb.run <- struct{}{}
	return &api.RunResp{}, nil
}

func (ls *LogSvr) Get(ctx context.Context, req *api.LogReq) (*api.LogResp, error) {
	jmu.Lock()
	if req.Id > uint32(len(jobs)) {
		jmu.Unlock()
		return nil, errors.New("job id is out of bounds")
	}
	jb := jobs[req.Id-1]
	jmu.Unlock()
	respChan := make(chan struct {
		logs   []string
		status api.JobStatus
	})
	jb.logs <- struct {
		resp chan struct {
			logs   []string
			status api.JobStatus
		}
	}{
		resp: respChan,
	}
	resp := <-respChan
	logResp := &api.LogResp{
		Lines:  resp.logs,
		Status: resp.status,
	}
	return logResp, nil
}

func (ls *LogSvr) GetStream(req *api.LogStreamReq, resp api.Log_GetStreamServer) error {
	jmu.Lock()
	if req.Id > uint32(len(jobs)) {
		jmu.Unlock()
		return errors.New("job id is out of bounds")
	}
	jb := jobs[req.Id-1]
	jmu.Unlock()

	logChan := make(chan string)
	jb.listen <- struct {
		ctx      context.Context
		listener chan string
	}{
		ctx:      context.Background(),
		listener: logChan,
	}
	for line := range logChan {
		err := resp.Send(&api.LogStreamResp{Line: line})
		if err != nil {
			return err
		}
	}
	return nil
}
