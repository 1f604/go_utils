Golang utils library.

# Usage

To use, simply place the `util.go` file into the same directory as your `main.go` file, then you can start using the functions provided.

# Functions provided

Currently contains 6 functionalities:

* Log rotation, you can set log rotation interval.
* Bypass "variable declared but not used".
* Make sure only 1 instance is running.
* Set log file.
* Re-launch process if it terminates (with or without args).
* Pkill process by name.

## Bypass "variable declared but not used"

The `Use` function gets around this compilation error.

You need this in order to maintain a reference to the TCP listener in the main thread, which you need to do in order to prevent the GC from closing the TCP port. 

## Check if another instance is already running

The `checkTCPPort` function checks to see if the TCP port is already used. 

If so, then it prints a message and exits your application.

Intended to be run first.

## Set log file

The `setLogFile` function sets a file to write all the log.print lines into.

If the file cannot be created or opened, it prints a message and exits your application.

Intended to be run second.

## Re-launch a process if it fails

The `retryproc` and `retryfunc` functions will keep trying to re-run a process or a function respectively.

These functions call the blocking process or function in a loop, therefore they consume almost no resources when the process or function is running.

There is a configurable backoff option. If the process or function exits too quickly, the backoff will increase until the set limit.

## Example

The main.go file has been included to show how to use these functions.

The source code for the sleep1 program is as follows:

```go
package main

import (
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	x := rand.Intn(5)
	println(x)
	time.Sleep(time.Duration(x) * time.Second)
}
```
