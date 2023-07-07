package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/recordMySystem-grpc/client/systeminfo"
	sys "github.com/recordMySystem-grpc/proto"
	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("localhost:8080", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("connection failed: %v", err)
	}
	defer conn.Close()
	c := sys.NewRecordMySystemClient(conn)
	CallClientStreaming(c)
}

func CallClientStreaming(c sys.RecordMySystemClient){
    var seconds int64 = 1
	stream,err := c.Store(context.Background())
	if err != nil {
		panic(err)
	}
 
	for{
		fmt.Println("storing stream started ",fmt.Sprint(seconds),"s ago")
		if seconds > 60 {
			break
		}
		sysData  := systeminfo.GetSystemInfo()
		req := sys.SystemInfoReq{
			Id : sysData.Id,
			HostName :sysData.HostName,
			CpuPercentage:float32(sysData.CpuPercentage),
			MemoryUsage:float32(sysData.MemoryUsage),
			Temperature:float32(sysData.Temperature),
			BatteryPercentage:float32(sysData.BatteryPercentage),
		}
		if err := stream.Send(&req); err != nil {
			panic(err)
		}
		seconds++
		time.Sleep(1*time.Second)
	}

	msg,err := stream.CloseAndRecv()
	if err != nil {
		panic(err)
	}
	fmt.Println(msg);
}