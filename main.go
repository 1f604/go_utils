package main

import (
	"log"
	"main/util"
	"runtime"
	"time"
)

var TCP_port = 12345
var logfilepath = "/tmp/logs/log.txt"
var g_count = 0

func foo() {
	g_count++
	log.Println("foo was executed: ", g_count)
	if g_count < 3 || g_count > 10 && g_count < 16 {
		time.Sleep(4 * time.Second)
		runtime.GC()
	}
}

func main() {
	l := util.CheckTCPPort(TCP_port)
	f := util.SetLogFile(logfilepath)
	defer util.Use(l, f) // need this otherwise GC will close TCP port.

	go util.Retryfunc("func foo", foo, 3*time.Second, 3*time.Second)
	go util.Retryproc("../sleep/sleep1", 3*time.Second, 3*time.Second)
	time.Sleep(100 * time.Hour)
}
