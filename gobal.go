package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"runtime"
	"time"
)

var filename = flag.String("f", "", "Path to file. Each line is a shell command.")
var capFlag = flag.Int("n", 1, "Number of threats in parallel.")
var vFlag = flag.Bool("v", false, "Outputs version information.")
var tot int

type Task struct {
	id int
	s  string
}

func main() {
	infoString := "GOBAL 1.2.1: a simple execution balancer. It reads a file containing one lines and runs these continuously on n processors."
	flag.Parse()
	if *vFlag {
		fmt.Println(infoString)
		os.Exit(0)
	}
	if *filename == "" {
		fmt.Println(infoString)
		fmt.Println("Please specify task file.")
		os.Exit(1)
	}

	start := time.Now()
	log.Println("RunWith", *capFlag, "hMaxGoroutines", runtime.GOMAXPROCS(0), "CPUs", runtime.NumCPU())
	tot, _ = lineCounter()
	log.Println("Starting", *filename, "with", tot, "tasks.")

	task := make(chan Task)
	quit := make(chan bool)

	for i := 0; i < *capFlag; i++ {
		go startWorker(i, task, quit)
	}

	// open a file
	file, err := os.Open(*filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var i int
	for scanner.Scan() {
		task <- Task{i, scanner.Text()}
		i++
		if i%10 == 5 && *capFlag < i {
			elapsed := time.Since(start)
			factor := float64(tot-i+1) / float64(i-*capFlag)
			estimatedTime := time.Duration(factor*float64(elapsed.Milliseconds())) * time.Millisecond
			log.Printf("======   elap:%s; remain:%s; total:%s; compl:%s =====", elapsed.Round(time.Second), estimatedTime.Round(time.Second), (elapsed + estimatedTime).Round(time.Second), time.Now().Add(estimatedTime).Format("02.01.2006 15:04:05"))
		}
	}

	for i := 0; i < *capFlag; i++ {
		quit <- true
	}
	close(quit)
	close(task)

	elapsed := time.Since(start)
	log.Printf("Gobal run took %s", elapsed)
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
			log.Printf("%v\\%v\t: %s\n", t.id+1, tot, t.s)

			f, err := ioutil.TempFile(".", "ex")
			check(err)
			_, err = f.Write([]byte(t.s))
			check(err)
			f.Close()

			if err := exec.Command("sh", f.Name()).Run(); err != nil {
				log.Println(i, ": error", err.Error())
			}
			err = os.Remove(f.Name())
			check(err)

		case <-quit:
			return
		}
	}
}

func lineCounter() (int, error) {
	r, err := os.Open(*filename)
	if err != nil {
		panic(err)
	}
	defer r.Close()
	buf := make([]byte, 32*1024)
	count := 0
	lineSep := []byte{'\n'}

	for {
		c, err := r.Read(buf)
		count += bytes.Count(buf[:c], lineSep)

		switch {
		case err == io.EOF:
			return count, nil

		case err != nil:
			return count, err
		}
	}
}
