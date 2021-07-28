package main

import (
	"fmt"
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
			rlog.println("launching " + taskname + " ...")
		}
		dotask()
		duration := time.Since(start)
		if print_log {
			rlog.Printf("%s finished after %d seconds.\n", taskname, duration/time.Second)
		}
		if duration > expected_duration {
			count.zero()
		} else {
			count.incr()
		}
		if print_log {
			rlog.Printf("%s: sleeping for %d seconds before re-running\n", taskname, count.getValue()/time.Second)
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

func Pkill(pname string) {
	cmd := exec.Command("pkill", pname)
	rlog.println("killing", pname)
	err := cmd.Run()
	if err != nil {
		rlog.println("error:", err)
	} else {
		rlog.println("done.")
	}
}

func Retryproc(procname string, expected_duration time.Duration, max_wait time.Duration) {
	f := func() {
		cmd := exec.Command(procname)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		rlog.Printf("launching process %s ...\n", procname)
		err := cmd.Run()

		if err != nil {
			rlog.Printf("process %s: an error occurred: %v\n", procname, err)
		} else {
			rlog.Printf("process %s completed without error.\n", procname)
		}
	}
	Retryfunc("command "+procname, f, expected_duration, max_wait)
}

// https://rosettacode.org/wiki/Determine_if_only_one_instance_is_running#Port
func CheckTCPPort(port int) net.Listener {
	var l net.Listener
	var err error
	if l, err = net.Listen("tcp", ":"+fmt.Sprint(port)); err != nil {
		rlog.Fatal("an instance was already running")
	}
	fmt.Println("single instance started.")
	return l
}

// https://stackoverflow.com/questions/21743841/how-to-avoid-annoying-error-declared-and-not-used
func Use(vals ...interface{}) {
	for _, val := range vals {
		_ = val
	}
}

// from https://stackoverflow.com/a/28797984
type RotateWriter struct {
	sync.Mutex
	filename string // should be set to the actual filename
	fp       *os.File
}

// Make a new RotateWriter. Return nil if error occurs during setup.
func NewRotateWriter(filename string) *RotateWriter {
	w := &RotateWriter{filename: filename}
	err := w.Rotate()
	if err != nil {
		panic("WRITER NOT CREATED! CHECK THE LOG FILE DIRECTORY EXISTS!")
	}
	return w
}

// Write satisfies the io.Writer interface.
func (w *RotateWriter) Write(output []byte) (int, error) {
	w.Lock()
	defer w.Unlock()
	return w.fp.Write(output)
}

// Perform the actual act of rotating and reopening file.
func (w *RotateWriter) Rotate() (err error) {
	w.Lock()
	defer w.Unlock()

	// Close existing file if open
	if w.fp != nil {
		err = w.fp.Close()
		w.fp = nil
		if err != nil {
			return
		}
	}
	// Rename dest file if it already exists
	_, err = os.Stat(w.filename)
	if err == nil {
		err = os.Rename(w.filename, w.filename+"."+time.Now().Format(time.RFC3339))
		if err != nil {
			return
		}
	}

	// Create a file.
	w.fp, err = os.Create(w.filename)
	return
}

// Rotates the file every x nanoseconds.
func (w *RotateWriter) LogRotater(x time.Duration) {
	ticker := time.NewTicker(x)
	for range ticker.C {
		w.Rotate()
	}
}

func (w *RotateWriter) Print(args ...interface{}) (int, error) {
	datestring := time.Now().Format("2006-01-02 15:04:05")
	output := []interface{}{datestring + " "}
	output = append(output, args...)
	str := fmt.Sprint(output...)
	return w.Write([]byte(str))
}

func (w *RotateWriter) Printf(formatstr string, args ...interface{}) (int, error) {
	formattedstr := fmt.Sprintf(formatstr, args...)
	datestring := time.Now().Format("2006-01-02 15:04:05")
	output := datestring + " " + formattedstr
	return w.Write([]byte(output))
}

func (w *RotateWriter) println(vals ...interface{}) (int, error) {
	vals = append(vals, "\n")
	return w.Print(vals...)
}

func (w *RotateWriter) Fatal(vals ...interface{}) {
	output := []interface{}{"FATAL: "}
	output = append(output, vals...)
	w.println(output...)
	os.Exit(1)
}
