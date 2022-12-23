// Copyright 2022 The Mandar Khadilkar. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// handler.go

package logrotate

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"sync"
	"time"

	"github.com/dustin/go-humanize"
)

// logfile maintains an internal state of the current log file while in use.
// Any modification should first lock the Rotator which uses object reference.
type logfile struct {
	filename string
	index    int

	size uint64
}

// Rotator is responsible for maintaining the io.Writer interface and at the same
// time runs the go routine which writes to the actual current log file.
// It currently checks the written size and rotates as needed.

type Rotator struct {
	internal    chan []byte
	limitbytes  uint64
	fileprefix  string
	maxlogfiles int
	locker      *sync.RWMutex
	running     bool
	lfile       *logfile
}

// Needed for
func (r Rotator) Write(Bytes []byte) (n int, err error) {
	if Bytes != nil {
		var bytes = make([]byte, len(Bytes))
		for i, _ := range Bytes {
			bytes[i] = Bytes[i]
		}
		select {
		case r.internal <- bytes:
			return len(bytes), nil
		case <-time.After(10 * time.Second):
			return 0, errors.New("Can not write to channel...")
		}
	}
	return 0, nil
}

func NewRotator() *Rotator {
	var R Rotator
	R.locker = new(sync.RWMutex)
	R.fileprefix = "logfile"
	R.maxlogfiles = 5
	R.limitbytes, _ = humanize.ParseBytes("100 mib")
	R.internal = make(chan []byte, 10)
	R.lfile = new(logfile)
	R.lfile.filename = R.fileprefix
	R.lfile.index = 0

	return &R
}

func (r *Rotator) Lock() {
	r.locker.Lock()
}
func (r *Rotator) UnLock() {
	r.locker.Unlock()
}
func (r *Rotator) resetFile() (w *os.File, err error) {
	r.lfile.size = uint64(0)
	w, err = os.OpenFile(r.lfile.filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return
	}
	fi, err := w.Stat()
	if err == nil {
		r.lfile.size = uint64(fi.Size())
	}
	return
}

func (r *Rotator) Start() {
	r.Lock()
	defer r.UnLock()

	if !r.running {

		ctl := make(chan bool, 0)

		go func(ct chan<- bool) {
			w, err := r.resetFile()
			if err != nil {
				return
			}
			ct <- true
			for {
				ba, ok := <-r.internal
				if !ok || ba == nil {
					w.Sync()
					w.Close()
					fmt.Println("Exiting handler")
					return
				} else {
					if r.lfile != nil {
						n, err := w.Write(ba)
						if err == nil {
							r.lfile.size += uint64(n)
						} else {
							fmt.Printf("Log writer error %+v\n", err)
						}
					}
				}

				r.Lock()

				if r.lfile.size > r.limitbytes {
					w.Sync()
					w.Close()

					err = os.Rename(r.lfile.filename, r.fileprefix+"."+strconv.Itoa(r.lfile.index))

					if r.lfile.index < (r.maxlogfiles - 1) {
						r.lfile.index++
					} else {
						r.lfile.index = 0
					}

					if err == nil {
						if w, err = r.resetFile(); err != nil {
							return
						}
					} else {
						fmt.Printf("error %+v\n", err)
					}
				}
				r.UnLock()
			}
		}(ctl)

		<-ctl

		r.running = true
	}
}

func (r *Rotator) Stop() {
	r.Lock()
	r.running = false
	r.UnLock()
	r.internal <- nil

	fmt.Printf("Stopped!\n")
}

func (r *Rotator) Close() {
	r.Lock()
	defer r.UnLock()
	close(r.internal)
	r.running = false
	fmt.Printf("closed!\n")
}

func (r *Rotator) Set(logfileSizeLimitbytes string, numberoflogfiles int, pathprefix string) (err error) {
	r.Lock()
	defer r.UnLock()
	r.limitbytes, err = humanize.ParseBytes(logfileSizeLimitbytes)
	if err != nil {
		return err
	}
	r.fileprefix = pathprefix
	if r.lfile != nil {
		r.lfile.filename = r.fileprefix
	}
	r.maxlogfiles = numberoflogfiles
	return nil
}
