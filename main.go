package main

import (
	"runtime"
	"time"
)

var (
	TCP_port                           = 12345
	rlog                 *RotateWriter = NewRotateWriter(log_directory)
	log_directory                      = "/tmp/logs/log.txt"
	log_rotationinterval               = 30 * time.Second
)
var g_count = 0

func foo() {
	g_count++
	rlog.println("foo was executed: ", g_count)
	if g_count < 3 || g_count > 10 && g_count < 16 {
		time.Sleep(4 * time.Second)
		runtime.GC()
	}
}

func main() {
	l := CheckTCPPort(TCP_port)
	go rlog.LogRotater(log_rotationinterval)
	defer Use(l) // need this otherwise GC will close TCP port.

	go Retryfunc("func foo", foo, 3*time.Second, 3*time.Second)
	//go Retryproc("../sleep/sleep1", 3*time.Second, 3*time.Second)
	time.Sleep(100 * time.Hour)
}
