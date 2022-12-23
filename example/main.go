// Copyright 2022 The Mandar Khadilkar. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package example show how can one use the logrotate feaure.
//

package main

import (
	"log"
	"os"

	"github.com/mkhadilk/logrotate"
)

func main() {
	var r *logrotate.Rotator = logrotate.NewRotator()
	if err := r.Set("10 kib", 2, "e.log"); err != nil {
		log.Printf("error %+v", err)
		os.Exit(-1)
	}
	r.Start()
	log.SetOutput(r)

	for i := 0; i < 10000; i++ {
		log.Printf("log line %d", i)
	}

	log.SetOutput(os.Stdout)

	r.Stop()
	r.Close()

}
