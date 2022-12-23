// Copyright 2022 The Mandar Khadilkar. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package logrotate

import (
	"io"
	"log"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/dustin/go-humanize"
)

func TestCreate(t *testing.T) {
	r := NewRotator()
	if r.fileprefix != "logfile" {
		log.Printf("default file name is not correct: neded %s found %s", "logfile", r.fileprefix)
		t.FailNow()
	}
	if r.limitbytes != uint64(104857600) {
		log.Printf("default size is not correct: needed %d found %d", 104857600, r.limitbytes)
		t.FailNow()
	}
	if r.lfile.index != 0 {
		log.Printf("default index is not correct: needed %d found %d", 0, r.lfile.index)
		t.FailNow()
	}
	if r.maxlogfiles != 5 {
		log.Printf("default index is not correct: needed %d found %d", 5, r.maxlogfiles)
		t.FailNow()
	}
}

func TestSet(t *testing.T) {
	r := NewRotator()
	r.Set("100 KB", 6, "somelog")
	if r.fileprefix != "somelog" {
		log.Printf("default file name is not correct: neded %s found %s", "somelog", r.fileprefix)
		t.FailNow()
	}
	if r.limitbytes != uint64(100000) {
		log.Printf("default size is not correct: needed %d found %d", 100000, r.limitbytes)
		t.FailNow()
	}
	if r.maxlogfiles != 6 {
		log.Printf("Max  files is not correct: needed %d found %d", 6, r.maxlogfiles)
		t.FailNow()
	}
}

func TestSetError(t *testing.T) {
	r := NewRotator()
	err := r.Set("100 mmmKB", 6, "somelog")
	if err == nil {
		t.FailNow()
	}
}

func TestResetFile(t *testing.T) {
	r := NewRotator()
	w, err := r.resetFile()
	if err != nil {
		t.Fail()
	}
	w.Close()
}

func TestWrite(t *testing.T) {
	r := NewRotator()
	ba := []byte("some data")
	n, err := r.Write(ba)
	if err != nil {
		t.FailNow()
	}
	if n != len(ba) {
		t.FailNow()
	}
	bb := <-r.internal
	if len(bb) != len(ba) {
		t.FailNow()
	}
}

func TestWriteError(t *testing.T) {
	r := NewRotator()
	ba := []byte("some data")
	for i := 0; i < 10; i++ {
		r.Write(ba)
	}
	_, err := r.Write(ba)
	if err == nil {
		t.FailNow()
	}
}

func TestStart(t *testing.T) {
	r := NewRotator()
	r.Set("10", 2, "logfile")
	r.Start()
	r.Stop()
	fi, err := os.Stat("logfile")
	if err != nil {
		t.FailNow()
	}
	if !fi.Mode().IsRegular() {
		t.FailNow()
	}
	log.Printf(fi.Mode().Perm().String())
}

func TestStartRotateOnce(t *testing.T) {
	r := NewRotator()
	r.Set("10", 2, "logfile")
	r.Start()
	for {
		_, err := r.Write([]byte("som"))
		if err == nil {
			break
		}
	}
	for {
		_, err := r.Write([]byte("1om789"))
		if err == nil {
			break
		}
	}
	for {
		_, err := r.Write([]byte("1om123"))
		if err == nil {
			break
		}
	}
	for {
		_, err := r.Write([]byte("2om123"))
		if err == nil {
			break
		}
	}
	for {
		<-time.After(1 * time.Second)
		if len(r.internal) == 0 {
			break
		}
	}
	r.Stop()
	r.Stop()
	fi, err := os.Stat("logfile")

	if err != nil {
		log.Println(err)
		t.FailNow()
	}
	if !fi.Mode().IsRegular() {
		log.Println("not a regular file")
		t.FailNow()
	}
	log.Printf(fi.Mode().Perm().String())

	fi, err = os.Stat("logfile.0")

	if err != nil {
		log.Println(err)
		t.FailNow()
	}
}

func TestStartRotateManytimes(t *testing.T) {
	r := NewRotator()
	r.Set("1 mib", 2, "logfile")
	os.Remove(r.fileprefix)
	os.Remove(r.fileprefix + "." + strconv.Itoa(r.lfile.index))
	os.Remove(r.fileprefix + "." + strconv.Itoa(r.lfile.index+1))
	r.Start()
	var i uint64
	limit, _ := humanize.ParseBytes("1 mib")
	for i = 0; i < limit*3; i++ {
		_, err := r.Write([]byte(strconv.Itoa(int(i)) + " somgar \n"))
		if err != nil {
			t.FailNow()
			break
		}
	}
	for {
		<-time.After(1 * time.Second)
		if len(r.internal) == 0 {
			break
		}
	}
	log.Printf("Wrote %d * %d bytes and limit was %d", i, len(strconv.Itoa(int(i))+" somgar \n"), limit)
	r.Stop()
	r.Close()
	fi, err := os.Stat("logfile")
	if err != nil {
		log.Println(err)
		t.FailNow()
	}
	if !fi.Mode().IsRegular() {
		log.Println("not a regular file")
		t.FailNow()
	}
	log.Printf(fi.Mode().Perm().String())

	fi, err = os.Stat("logfile.0")

	if err != nil {
		log.Println(err)
		t.FailNow()
	}
	fi, err = os.Stat("logfile.1")

	if err != nil {
		log.Println(err)
		t.FailNow()
	}
}

func TestStopUnstarted(t *testing.T) {
	r := NewRotator()
	r.Set("10", 2, "logfile")
	r.Start()
	r.Stop()
	fi, err := os.Stat("logfile")
	if err != nil {
		t.FailNow()
	}
	if !fi.Mode().IsRegular() {
		t.FailNow()
	}
	log.Printf(fi.Mode().Perm().String())
	r.Stop()
}

func TestRestart(t *testing.T) {
	r := NewRotator()
	os.Remove("logfilerestart")
	r.Set("1000", 2, "logfilerestart")
	r.Start()
	r.Write([]byte("before stop\n"))
	for {
		<-time.After(1 * time.Second)
		if len(r.internal) == 0 {
			break
		}
	}
	r.Stop()
	in, err := os.Open(r.fileprefix)
	if err != nil {
		r.Close()
		t.Log(err)
		t.FailNow()
	}
	ba, err := io.ReadAll(in)
	in.Close()
	if string(ba) != "before stop\n" {
		r.Close()
		t.Logf("Data not matching - after stop wanted \n%s actual %s", "before stop\n", string(ba))
		t.FailNow()
	}
	r.Write([]byte("after stop\n"))
	in, err = os.Open(r.fileprefix)
	if err != nil {
		r.Close()
		t.Log(err)
		t.FailNow()
	}
	ba, err = io.ReadAll(in)
	in.Close()
	if string(ba) != "before stop\n" {
		r.Close()
		t.Logf("Data not matching - after stop wanted \n%s actual %s", "before stop\n", string(ba))
		t.FailNow()
	}
	fi, err := os.Stat("logfilerestart")
	if err != nil {
		t.FailNow()
	}
	if !fi.Mode().IsRegular() {
		t.FailNow()
	}
	log.Printf(fi.Mode().Perm().String())
	r.Start()
	for {
		<-time.After(1 * time.Second)
		if len(r.internal) == 0 {
			break
		}
	}
	in, err = os.Open(r.fileprefix)
	if err != nil {
		r.Close()
		t.Log(err)
		t.FailNow()
	}
	ba, err = io.ReadAll(in)
	in.Close()
	if string(ba) != "before stop\nafter stop\n" {
		r.Close()
		t.Logf("Data not matching - after restart wanted %s actual %s", "before stop\nafter stop\n", string(ba))
		t.FailNow()
	}
	r.Write([]byte("after restart\n"))
	<-time.After(time.Second * 1)
	r.Stop()
	<-time.After(time.Second * 1)
	r.Close()
	<-time.After(time.Second * 1)
	in, err = os.Open(r.fileprefix)
	if err != nil {
		r.Close()
		t.Log(err)
		t.FailNow()
	}
	ba, err = io.ReadAll(in)
	in.Close()
	if string(ba) != "before stop\nafter stop\nafter restart\n" {
		t.Logf("Data not matching - after restart wanted %s actual %s", "before stop\nafter stop\nafter restart\n", string(ba))
		t.FailNow()
	}
}
