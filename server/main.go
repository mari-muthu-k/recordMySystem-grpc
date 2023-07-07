package main

import (
	"fmt"
	"io"
	"log"
	"net"

	"github.com/recordMySystem-grpc/client/systeminfo"
	sys "github.com/recordMySystem-grpc/proto"
	"github.com/recordMySystem-grpc/server/clients"
	"github.com/recordMySystem-grpc/server/database"
	"google.golang.org/grpc"
)


type Server struct{}

func(*Server) Store( reqStream sys.RecordMySystem_StoreServer)error{
	for {
		var sysData systeminfo.SystemInfo

		req,err := reqStream.Recv()
		if err == io.EOF {
			return reqStream.SendAndClose(&sys.SystemInfoRes{Message: "ok"})
		}
		if err != nil {
			panic(err)
		}

		sysData.Id = req.GetId()
		sysData.HostName = req.GetHostName()
		sysData.BatteryPercentage = float64(req.GetBatteryPercentage())
		sysData.CpuPercentage     = float64(req.GetCpuPercentage())
		sysData.MemoryUsage       = float64(req.GetMemoryUsage())
		sysData.Temperature       = float64(req.GetTemperature())

		_,err = database.InsertData(&sysData)
		if err != nil {
			log.Fatalf("unable to insert data : ",err)
		}
	}
}

func appStartup(){
	fmt.Println("connecting influx client...")
	//Connect influx db
	database.ConnectDB(&clients.InfluxClient)
	defer clients.InfluxClient.Close()

	clients.InfluxWriteAPI = clients.InfluxClient.WriteAPIBlocking(clients.ORGANIZATION,clients.BUCKET)
	clients.InfluxQueryAPI = clients.InfluxClient.QueryAPI(clients.ORGANIZATION)
	fmt.Println("influx client connected")
}

func main(){
	appStartup()
	net,err := net.Listen("tcp","localhost:8080"); if err != nil {
		panic(err)
	}

	s := grpc.NewServer()
	sys.RegisterRecordMySystemServer(s,&Server{})

	if err := s.Serve(net); err != nil {
		panic(err)
	}
}