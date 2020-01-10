package main

import (
	"flag"
	"io"
	"io/ioutil"
	"os"
	"time"
        "runtime"

	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/api/resource"
)
/*
#define _GNU_SOURCE
#include <sched.h>
#include <pthread.h>

void lock_thread(int cpuid) {
    pthread_t tid;
    cpu_set_t cpuset;
    int i,n,setsize;

    tid = pthread_self();
    setsize = sizeof(cpu_set_t);

    pthread_getaffinity_np(tid, setsize, &cpuset);
    n = CPU_COUNT(&cpuset);
    cpuid = cpuid % n;
    i = -1;
    while (cpuid > -1 ) {
        i++;
        if(CPU_ISSET(i, &cpuset))
            cpuid--;
    }
    CPU_ZERO(&cpuset);
    CPU_SET(i, &cpuset);
    pthread_setaffinity_np(tid, setsize, &cpuset);
}
*/
import "C"

var (
	argMemTotal         = flag.String("mem-total", "0", "total memory to be consumed. Memory will be consumed via multiple allocations.")
	argMemStepSize      = flag.String("mem-alloc-size", "4Ki", "amount of memory to be consumed in each allocation")
	argMemSleepDuration = flag.Duration("mem-alloc-sleep", time.Millisecond, "duration to sleep between allocations")
	argCpus             = flag.Int("cpus", 0, "total number of CPUs to utilize")
        argFirstCpu         = flag.Int("first-cpu", 0, "first cpu in contiguous series to attempt to pin")
	buffer              [][]byte
)

func main() {
	flag.Parse()
	total := resource.MustParse(*argMemTotal)
	stepSize := resource.MustParse(*argMemStepSize)
	glog.Infof("Allocating %q memory, in %q chunks, with a %v sleep between allocations", total.String(), stepSize.String(), *argMemSleepDuration)
	burnCPU()
	allocateMemory(total, stepSize)
	glog.Infof("Allocated %q memory", total.String())
	select {}
}

func worker(id int, src *os.File) {
    glog.Infof("Pinning thread to CPU %d", id)
    runtime.LockOSThread()
    C.lock_thread(C.int(id))

    _, err := io.Copy(ioutil.Discard, src)
    if err != nil {
        glog.Fatalf("failed to copy from /dev/zero to /dev/null: %v", err)
    }
}

func burnCPU() {
	src, err := os.Open("/dev/zero")
	if err != nil {
		glog.Fatalf("failed to open /dev/zero")
	}
	for i := *argFirstCpu; i < *argCpus+*argFirstCpu; i++ {
		glog.Infof("Spawning a thread to consume CPU")
		go worker(i, src)
	}
}

func allocateMemory(total, stepSize resource.Quantity) {
	for i := int64(1); i*stepSize.Value() <= total.Value(); i++ {
		newBuffer := make([]byte, stepSize.Value())
		for i := range newBuffer {
			newBuffer[i] = 0
		}
		buffer = append(buffer, newBuffer)
		time.Sleep(*argMemSleepDuration)
	}
}
