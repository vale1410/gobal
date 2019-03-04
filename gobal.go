package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
)

var filename = flag.String("f", "", "Path to file. Each line is a shell command.")
var capFlag = flag.Int("n", 1, "Number of threats in parallel.")

type Task struct {
	id int
	s  string
}

func main() {
	flag.Parse()
	fmt.Println("GOBAL 1.1 Beta: a simple execution balancer. It reads a file containing one-liners of shell executions and runs these continuously on n processors.")
	fmt.Println("RunWith", *capFlag, "hMaxGoroutines", runtime.GOMAXPROCS(0), "CPUs", runtime.NumCPU())

	task := make(chan Task)
	quit := make(chan bool)

	for i := 0; i < *capFlag; i++ {
		go startWorker(i, task, quit)
	}

	// open a file
	file, _ := os.Open(*filename)
	defer file.Close()
	scanner := bufio.NewScanner(file)
	var i int
	for scanner.Scan() {
		task <- Task{i, scanner.Text()}
		i++
	}

	for i := 0; i < *capFlag; i++ {
		quit <- true
	}
	close(quit)
	close(task)

}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func startWorker(i int, task chan Task, quit chan bool) {
	for {
		select {
		case t := <-task:
			fmt.Println(t.id, "\t:", t.s)

			f, err := ioutil.TempFile(".", "ex")
			check(err)
			_, err = f.Write([]byte(t.s))
			check(err)
			f.Close()

			if err := exec.Command("sh", f.Name()).Run(); err != nil {
				fmt.Println(i, ": error", err.Error())
			}
			err = os.Remove(f.Name())
			check(err)

		case <-quit:
			return
		}
	}
}
