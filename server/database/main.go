package database

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	systeminfo "github.com/recordMySystem-grpc/client/systeminfo"
	db "github.com/recordMySystem-grpc/server/clients"
)

func ConnectDB(dbClient *influxdb2.Client){
	token := os.Getenv("INFLUXDB_TOKEN")
	url := "http://localhost:8086"
	*dbClient = 	influxdb2.NewClient(url, token)
}

func InsertData(sysData *systeminfo.SystemInfo)(bool,error){
	var tags = map[string]string{
		"id" : sysData.Id,
	}

	var fields = map[string]interface{}{
		"hostName":sysData.HostName,
		"batteryPercentage":sysData.BatteryPercentage,
		"memoryUsage":sysData.MemoryUsage,
		"temperature":sysData.Temperature,
		"cpuPercentage":sysData.CpuPercentage,
	}

	point := write.NewPoint(db.MEASUREMENT,tags,fields,time.Now())
	if err := db.InfluxWriteAPI.WritePoint(context.Background(),point); err != nil {
		return false,err
	}

	return true,nil
}

func QueryData(fields []string,startTime string,endTime string,id string,sysData *systeminfo.GetSystemInfoData)(bool,error){
	var isDataExist bool
	query := BuildQuery(fields,startTime,endTime,id)
	results, err := db.InfluxQueryAPI.Query(context.Background(), query)
	if err != nil {
		return isDataExist,err
	}

	for results.Next() {
		var tArr []interface{}
		record := results.Record()
		val := record.Value()
        if !isDataExist || val != nil {
			isDataExist = true
		}

		tArr = append(tArr,record.Time())
		if record.Field() != "hostName" {
			tArr = append(tArr, val.(float64))
		}

		switch record.Field() {
		case "cpuPercentage":
			sysData.CpuPercentage = append(sysData.CpuPercentage,tArr)
		case "temperature":
			sysData.Temperature = append(sysData.Temperature,tArr)
		case "memoryUsage":
			sysData.MemoryUsage = append(sysData.MemoryUsage,tArr)
		case "batteryPercentage":
			sysData.BatteryPercentage = append(sysData.BatteryPercentage,tArr)
		case "hostName":
			tArr = append(tArr,val.(string))
			sysData.HostName = append(sysData.HostName,tArr)
		}
	}
	
	if err := results.Err(); err != nil {
		return isDataExist,err
	}

	return isDataExist,nil
}

func BuildQuery(fields []string,startTime string,endTime string,id string)string{
	sTime,err := strconv.Atoi(startTime)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	eTime,err := strconv.Atoi(endTime)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	// Convert Unix timestamp to time.Time
	st := time.Unix(0,  int64(sTime)*int64(time.Millisecond))
	et := time.Unix(0,  int64(eTime)*int64(time.Millisecond))

	sISO := st.UTC().Format("2006-01-02T15:04:05.999Z")
	eISO := et.UTC().Format("2006-01-02T15:04:05.999Z")

	query := fmt.Sprintf(
		      `from(bucket: "%s")
			  |> range(start:%s,stop:%s)
			  |> filter(fn: (r) => r["_measurement"] == "%s")`,db.BUCKET,sISO,eISO,db.MEASUREMENT)

	for _,field := range fields {
		    query += fmt.Sprintf(`|> filter(fn: (r)=> r["_field"]=="%s")`,field)
	}

	query += fmt.Sprintf(`|> filter(fn:(r)=> r["id"]=="%s")
	                      |> keep(columns:["_field","_value","_time"])`,id)
	return query
}