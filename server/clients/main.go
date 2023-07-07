package clients

import (
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

var InfluxClient influxdb2.Client
var InfluxWriteAPI api.WriteAPIBlocking
var InfluxQueryAPI api.QueryAPI

const (
	ORGANIZATION = "DOT"
    BUCKET       = "SystemData"
	MEASUREMENT  = "systemInfo"
)