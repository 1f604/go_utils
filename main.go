package main

import (
	"log"
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
	l := checkTCPPort(TCP_port)
	f := setLogFile(logfilepath)
	defer Use(l, f) // need this otherwise GC will close TCP port.

	go retryfunc("func foo", foo, 3*time.Second, 3*time.Second)
	go retryproc("../sleep/sleep1", 3*time.Second, 3*time.Second)
	time.Sleep(100 * time.Hour)
}
