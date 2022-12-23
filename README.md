# logrotate
This is a general purpose file rotator based on size of the each file and limited by the number of such files. It acts as a ring buffer i.e. when the last indexed file reaches the limit, it writes to the first file and then next. It is well suited for log writing.

It implements io.Writer interface. It spins off one go routine.

The actual writes happen to a 'prefixed file' and then when that reaches the limit, it renames the file chunk to next indexed file. It reinitiat0es the write to original file and repeats itself.

Steps to use-

1. import the package

1. Use the constructor given to create an Instance (say r)
  ```go
  var r *logrotate.Rotator = logrotate.NewRotator()
  ```

3. Call a Set function to prefix file name to use, number of files and size per file in human readable format (100 mib/1 big etc)
  ```go
  r.Set("10 kib", 2, "e.log")
  ```
  
4. When ready call Start function
  ```go
  r.Start()
  ```

5. Now a logger can be set with that r (it implements io.Writer). log.SetOutput(r)
  ```go
  log.SetOutput(r)
  ```
  
6. When all is done and time to stop writing, call r.Stop() for temporary pause of writing and r.Start() again when ready. Meanwhile, the writes after the stop (max 100) will be in 'hold' state waiting for 'Start' to be called again. Upon restart, pending writes will be made to file
  ```go
  r.Stop()
  ..........
  r.Start()
  ```

7. When needed, call r.Close() to completely close the writer. In that case, all pending writes will be abandoned
  ```go
  r.Close()
