package logw

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"time"
	"types"
	"utils"
)

//Event - event type
type Event struct {
	Message string
}

//Error - error type
type Error struct {
	Message string
}

//Log type
type Log struct {
	Duration float64
}

//File - type
type File struct {
	Name string
	Date time.Time
}

//Init - starts log service.
func (log Log) Init(config *types.Config) *Log {
	log.Duration = 24 * config.LogDuration
	os.MkdirAll("./logs", os.ModePerm)
	os.MkdirAll("./logs/errors", os.ModePerm)
	os.MkdirAll("./logs/events", os.ModePerm)
	utils.Schedule(log.cleanUp, 2*time.Hour)
	return &log
}

//cleanUp - removes expired files
func (log Log) cleanUp() {
	log.cleanUpErros()
	log.cleanUpEvents()
}

func (log Log) cleanUpErros() {
	var files []File
	err := filepath.Walk("./logs/errors", func(path string, info os.FileInfo, err error) error {
		files = append(files, File{Date: info.ModTime(), Name: info.Name()})
		return nil
	})
	if err != nil {
		return
	}
	for _, file := range files {
		if file.Name == "errors" {
			continue
		}
		duration := time.Since(file.Date)
		hours := duration.Hours()
		if hours > log.Duration {
			os.Remove("./logs/errors/" + file.Name)
		}
	}
}

func (log Log) cleanUpEvents() {
	var files []File
	err := filepath.Walk("./logs/events", func(path string, info os.FileInfo, err error) error {
		files = append(files, File{Date: info.ModTime(), Name: info.Name()})
		return nil
	})
	if err != nil {
		return
	}
	for _, file := range files {
		if file.Name == "events" {
			continue
		}
		duration := time.Since(file.Date)
		hours := duration.Hours()
		if hours > log.Duration {
			os.Remove("./logs/events/" + file.Name)
		}
	}
}

//LogError - logs a error given
func (log Log) LogError(err Error) {
	yy, mm, dd := time.Now().Date()
	date := strconv.Itoa(dd) + "-" + mm.String() + "-" + strconv.Itoa(yy)
	data := []byte("\n\n" + time.Now().Format("2006-01-02 15:04:05: ") + err.Message)
	fileName := "./logs/errors/" + date + ".log"
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		er := ioutil.WriteFile(fileName, data, 0644)
		if er != nil {
			fmt.Println("Log: Could not write to file")
			return
		}
	} else {
		f, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			fmt.Println("Log: Could not open file")
			return
		}
		defer f.Close()
		if _, err = f.WriteString(string(data)); err != nil {
			fmt.Println("Log: Could not append to file")
			return
		}
	}

}

//LogEvent - logs a event given
func (log Log) LogEvent(event Event) {
	yy, mm, dd := time.Now().Date()
	date := strconv.Itoa(dd) + "-" + mm.String() + "-" + strconv.Itoa(yy)
	data := []byte("\n\n" + time.Now().Format("2006-01-02 15:04:05: ") + event.Message)
	fileName := "./logs/events/" + date + ".log"
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		er := ioutil.WriteFile(fileName, data, 0644)
		if er != nil {
			fmt.Println("Log: Could not write to file")
			return
		}
	} else {
		f, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			fmt.Println("Log: Could not open file")
			return
		}
		defer f.Close()
		if _, err = f.WriteString(string(data)); err != nil {
			fmt.Println("Log: Could not append to file")
			return
		}
	}
}
