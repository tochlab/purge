package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

var purgeTime time.Time
var startTime time.Time
var reallyDelete bool
var beVerbose bool
var deletedFilesCount uint64
var rescanInterval time.Duration

func deleteIt(filePath string) {
	rmErr := os.Remove(filePath)
	if rmErr != nil {
		log.Println(filePath, ": ",rmErr)
	}
}

func checkFile(filePath string, info os.FileInfo, err error) error {
	if info.IsDir() {
		return nil
	}
	if info.ModTime().Before(purgeTime) {
		deletedFilesCount++
		if beVerbose {
			log.Println("Delete:", filePath, info.ModTime())
		}
		if reallyDelete {
			deleteIt(filePath)
		}
		if IsEmpty(path.Dir(filePath)) {
			log.Println("Delete empty dir", path.Dir(filePath))
			if reallyDelete {
				deleteIt(path.Dir(filePath))
			}
		}
	}
	return nil
}

func IsEmpty(name string) bool {
	f, err := os.Open(name)
	if err != nil {
		log.Println(name, err)
		return false
	}
	defer f.Close()

	_, err = f.Readdirnames(1)
	if err == io.EOF {
		return true
	}
	return false
}

func IsExists(name string) bool {
	_, err := os.Stat(name)
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}

func main() {

	additionalFlags := flag.NewFlagSet("additional", flag.ContinueOnError)

	additionalFlags.BoolVar(&beVerbose, "v", false, "Be verbose")
	additionalFlags.BoolVar(&reallyDelete,"d", false,"Delete files for real")
	additionalFlags.DurationVar(&rescanInterval, "i", 0, "Rescan interval")

	argsWithoutProg := os.Args[1:]
	if len(argsWithoutProg) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <path> <time> [-v] [-d] [-i time]\n", os.Args[0])
		fmt.Fprintf(os.Stderr,"path\tpath to scan\n")
		fmt.Fprintf(os.Stderr,"time\tsearch files older then <time>\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr,"\nTime examples:\n")
		fmt.Fprintf(os.Stderr,"\t3h6m12s - 3 hours 6 minutes and 12 seconds\n")
		fmt.Fprintf(os.Stderr,"\t7m11s   - 7 minutes and 11 seconds\n")
		fmt.Fprintf(os.Stderr,"\t100500s - 100500 seconds\n")
		fmt.Fprintf(os.Stderr,"\t3h      - 3 hours\n")
		os.Exit(1)
	}
	scanPath := argsWithoutProg[0]
	cutSecStr := argsWithoutProg[1]

	additionalFlags.Parse(os.Args[3:])

	if rescanInterval == 0 && !IsExists(scanPath) {
		os.Exit(1)
	}

	if strings.ContainsAny(cutSecStr, "shm") == false {
		cutSecStr = cutSecStr + "s"
	}

	cutDur, parseErr := time.ParseDuration(cutSecStr)
	if parseErr != nil {
		log.Panic(parseErr)
	}

	for {
		startTime = time.Now()

		purgeTime = startTime.Add(-cutDur)

		log.Println("Scan start time", startTime)
		log.Println("Purge time", purgeTime)

		if IsExists(scanPath) {
			filepath.Walk(scanPath, checkFile)

			if !reallyDelete {
				log.Printf("*** TEST MODE! ***")
			}
			log.Printf("Done. Deleted %d files. %.2f files/sec.", deletedFilesCount, float64(deletedFilesCount)/time.Now().Sub(startTime).Seconds())
		}

		if rescanInterval != 0 {
			time.Sleep(rescanInterval)
			deletedFilesCount = 0
		} else {
			break
		}
	}
}
