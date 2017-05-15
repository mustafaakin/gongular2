package gongular2

import (
	"bytes"
	"fmt"
	"net/http"
	"time"
)

type HandlerStat struct {
	FuncName  string
	Duration  time.Duration
	Error     error
	StopChain bool
}

type RouteStat struct {
	Request       *http.Request
	Handlers      []HandlerStat
	MatchedPath   string
	TotalDuration time.Duration
	ResponseSize  int
	ResponseCode  int
	Logs          *bytes.Buffer
}

type RouteCallback func(stat RouteStat)

var DefaultRouteCallback RouteCallback = func(stat RouteStat) {
	fmt.Println(stat.Request.Method, stat.Request.RemoteAddr, stat.MatchedPath,
		stat.Request.RequestURI, stat.TotalDuration, stat.ResponseSize, stat.ResponseCode)

	for idx, h := range stat.Handlers {
		fmt.Println("\t", idx, " ", h.FuncName, " ", h.Duration)
	}
}
