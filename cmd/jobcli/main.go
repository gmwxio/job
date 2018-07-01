package main

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/wxio/job/api"
	"google.golang.org/grpc"
)

const (
	jobSvrAddress = "localhost:50051"
)

func main() {
	conn, err := grpc.Dial(jobSvrAddress, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}
	defer conn.Close()
	conn2, err := grpc.Dial(jobSvrAddress, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}
	defer conn2.Close()
	jobclient := api.NewJobClient(conn)
	logclient := api.NewLogClient(conn2)
	initResp, err := jobclient.Init(context.Background(), &api.InitReq{})
	if err != nil {
		log.Fatalf("init failed: %v", err)
	}
	fmt.Printf("job init resp %v\n", initResp)
	_, err = jobclient.Run(context.Background(), &api.RunReq{Id: initResp.Id})
	if err != nil {
		log.Fatalf("run failed: %v", err)
	}
	lc_stream, err := logclient.GetStream(context.Background(), &api.LogStreamReq{Id: initResp.Id})
	if err != nil {
		log.Fatalf("get stream: %v", err)
	}
	for {
		lsr, err := lc_stream.Recv()
		if err == io.EOF {
			fmt.Printf("------------------\n")
			break
		}
		if err != nil {
			log.Fatalf("in stream: %v", err)
		}
		fmt.Printf("%d : %s\n", initResp.Id, lsr.Line)
	}
	lresp, err := logclient.Get(context.Background(), &api.LogReq{Id: initResp.Id})
	if err != nil {
		log.Fatalf("get log: %v", err)
	}
	fmt.Printf("%d : %v\n", initResp.Id, lresp.Status)
	for _, line := range lresp.Lines {
		fmt.Println(line)
	}

}
