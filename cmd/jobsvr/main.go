package main

import (
	"log"
	"net"

	"github.com/wxio/job/api"
	"github.com/wxio/job/service/job"
	"google.golang.org/grpc"
)

const (
	port = ":50051"
)

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	svr := grpc.NewServer()

	jbSvr := &job.JobSvr{}
	logSvr := &job.LogSvr{}
	api.RegisterJobServer(svr, jbSvr)
	api.RegisterLogServer(svr, logSvr)
	// reflection.Register(s)
	if err := svr.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
