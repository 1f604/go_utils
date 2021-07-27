package util

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"sync"
	"time"
)

type retrylib_task func()

type retrylib_counter struct {
	mu               sync.Mutex
	max_val          time.Duration
	private_variable time.Duration
}

func (c *retrylib_counter) incr() {
	c.mu.Lock()
	if c.private_variable < c.max_val {
		c.private_variable += time.Second
	}
	c.mu.Unlock()
}

func (c *retrylib_counter) getValue() time.Duration {
	c.mu.Lock()
	var n = c.private_variable
	c.mu.Unlock()
	return n
}

func (c *retrylib_counter) maxValReached() bool {
	c.mu.Lock()
	var n = c.private_variable
	c.mu.Unlock()
	return n >= c.max_val
}

func (c *retrylib_counter) zero() {
	c.mu.Lock()
	c.private_variable = 0
	c.mu.Unlock()
}

func newRetrylibCounter(maxval time.Duration) *retrylib_counter {
	return &retrylib_counter{max_val: maxval}
}

func Retryfunc(taskname string, dotask retrylib_task, expected_duration time.Duration, max_wait time.Duration) {
	count := newRetrylibCounter(max_wait)
	for {
		start := time.Now()
		print_log := !count.maxValReached()
		if print_log {
			log.Printf("launching %s ...\n", taskname)
		}
		dotask()
		duration := time.Since(start)
		if print_log {
			log.Printf("%s finished after %d seconds.\n", taskname, duration/time.Second)
		}
		if duration > expected_duration {
			count.zero()
		} else {
			count.incr()
		}
		if print_log {
			log.Printf("%s: sleeping for %d seconds before re-running\n", taskname, count.getValue()/time.Second)
		}
		time.Sleep(count.getValue())
	}
}

func RetryprocWithArgs(taskname string, procname string, args []string, expected_duration time.Duration, max_wait time.Duration) {
	f := func() {
		cmd := exec.Command(procname, args...)
		cmd.Run()
	}
	Retryfunc("command "+taskname, f, expected_duration, max_wait)
}

func Retryproc(procname string, expected_duration time.Duration, max_wait time.Duration) {
	f := func() {
		cmd := exec.Command(procname)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		log.Printf("launching process %s ...\n", procname)
		err := cmd.Run()

		if err != nil {
			log.Printf("process %s: an error occurred: %v\n", procname, err)
		} else {
			log.Printf("process %s completed without error.\n", procname)
		}
	}
	Retryfunc("command "+procname, f, expected_duration, max_wait)
}

// https://rosettacode.org/wiki/Determine_if_only_one_instance_is_running#Port
func CheckTCPPort(port int) net.Listener {
	var l net.Listener
	var err error
	if l, err = net.Listen("tcp", ":"+fmt.Sprint(port)); err != nil {
		log.Fatal("an instance was already running")
	}
	fmt.Println("single instance started.")
	return l
}

// https://stackoverflow.com/questions/19965795/how-to-write-log-to-file
func SetLogFile(filename string) *os.File {
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening log file: %v", err)
	}

	log.SetOutput(f)
	log.Println("Started logging.")
	return f
}

// https://stackoverflow.com/questions/21743841/how-to-avoid-annoying-error-declared-and-not-used
func Use(vals ...interface{}) {
	for _, val := range vals {
		_ = val
	}
}
