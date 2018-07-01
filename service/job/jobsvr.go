package job

import (
	"context"

	"github.com/wxio/job/api"
)

type jobSvr struct {
}
type logSvr struct {
}

var _ api.JobServer = &jobSvr{}
var _ api.LogServer = &logSvr{}

func (ps *jobSvr) Init(context.Context, *api.InitReq) (*api.InitResp, error) {
	return nil, nil
}
func (ps *jobSvr) Run(context.Context, *api.RunReq) (*api.RunResp, error) {
	return nil, nil
}

func (ls *logSvr) Get(context.Context, *api.LogReq) (*api.LogResp, error) {
	return nil, nil
}
func (ls *logSvr) GetStream(req *api.LogStreamReq, resp api.Log_GetStreamServer) error {
	return nil
}
