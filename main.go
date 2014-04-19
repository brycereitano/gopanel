package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

/*
 * Xtitle is used to pass the stream of data from xtitle -s into the ch string.
 * It passes all of its errors through the errCh 
 * To invoke this function: go Xtitle(2000, ch, errCh)
*/
func Xtitle(interval int64, ch chan string, errCh chan error) {
	buff := make([]byte, 1024)
	cmd := exec.Command("xtitle", "-sf", "'%s'")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		errCh <- err
		return
	}

	err = cmd.Start()
	if err != nil {
		errCh <- err
		return
	}

	func() {
		for {
			_, err = stdout.Read(buff)
			if err != nil {
				errCh <- err
				return
			}

			ch <- string(buff)
			buff = make([]byte, 1024)

			time.Sleep(time.Duration(interval) * time.Millisecond)
		}
	}()
}

/*
 * Load is used to pass the results of cat /proc/loadavg into the ch string every 'interval' milliseconds.
 * It passes all of its errors through the errCh 
 * To invoke this function: go Load(2000, ch, errCh)
*/
func Load(interval int, ch chan string, errCh chan error) {
	for {
		out, err := ioutil.ReadFile("/proc/loadavg")
		if err != nil {
			errCh <- err
			return
		}

		outStr := strings.Fields(string(out))
		result := fmt.Sprintf("Load: %s %s %s\n", outStr[0], outStr[1], outStr[2])
		ch <- result

		time.Sleep(time.Duration(interval) * time.Millisecond)
	}
}

/*
 * Memory is used to pass the results of free -m into the ch string every 'interval' milliseconds.
 * It passes all of its errors through the errCh 
 * To invoke this function: go Memory(2000, ch, errCh)
*/
func Memory(interval int64, ch chan string, errCh chan error) {
	for {
		out, err := exec.Command("free", "-m").Output()
		if err != nil {
			errCh <- err
			return
		}

		outStr := strings.Split(string(out), "\n")
		outStr = strings.Fields(outStr[1])

		used, err := strconv.ParseFloat(outStr[3], 32)
		if err != nil {
			errCh <- err
			return
		}

		total, err := strconv.ParseFloat(outStr[1], 32)
		if err != nil {
			errCh <- err
			return
		}

		result := fmt.Sprintf("Mem: %.2f/%.2fG\n", used/1024, total/1024)
		ch <- result

		time.Sleep(time.Duration(interval) * time.Millisecond)
	}
}

/*
 * Volume is used to pass the processed results of pacmd dump into the ch string every 'interval' milliseconds.
 * It passes all of its errors through the errCh 
 * To invoke this function: go Volume(2000, ch, errCh)
*/
func Volume(interval int64, ch chan string, errCh chan error) {
	for {
		out, err := exec.Command("pacmd", "dump").Output()
		if err != nil {
			errCh <- err
			return
		}

		outStr := strings.Split(string(out), "\n")
		var desiredInt uint64
		for _, value := range outStr {
			if strings.HasPrefix(value, "set-sink-volume") {
				strs := strings.Fields(value)
				desiredInt, err = strconv.ParseUint(strs[len(strs)-1][2:], 16, 16)
				if err != nil {
					errCh <- err
					return
				}
			}
		}

		ch <- fmt.Sprintf("Vol: %d%%\n", int(float64(desiredInt)/655.36))
		time.Sleep(time.Duration(interval) * time.Millisecond)
	}
}

/*
 * Time is used to pass the formatted results of time.Now() into the ch channel string every 'interval' milliseconds.
 * It passes all of its errors through the errCh 
 * To invoke this function: go Time(2000, ch, errCh)
*/
func Time(interval int, ch chan string, errCh chan error) {
	for {
		ch <- time.Now().Format("Mon, Jan 2, 2006 15:04:05\n")
		time.Sleep(time.Duration(interval) * time.Millisecond)
	}
}

/*
 * Temp is used to pass the contents of the path specified into the ch channel string every 'interval' milliseconds.
 * It passes all of its errors through the errCh 
 * To invoke this function: go Temp(2000, ch, errCh)
*/
func Temp(interval int, ch chan string, errCh chan error) {
	for {
		out, err := ioutil.ReadFile("/sys/bus/platform/devices/coretemp.0/temp1_input")
		if err != nil {
			errCh <- err
			return
		}

		temp, err := strconv.ParseInt(string(out[:len(out)-4]), 10, 16)
		if err != nil {
			errCh <- err
			return
		}

		ch <- fmt.Sprintf("Temp: %dC\n", temp)
		time.Sleep(time.Duration(interval) * time.Millisecond)
	}
}

/*
 * StatusWm is used to pass the contents of the path specified into the ch channel string every 'interval' milliseconds.
 * It passes all of its errors through the errCh 
 * To invoke this function: go Temp(2000, ch, errCh)
*/
func StatusWm(interval int64, ch chan string, errCh chan error) {
	buff := make([]byte, 1024)
	cmd := exec.Command("bspc", "control", "--subscribe")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		errCh <- err
		return
	}

	err = cmd.Start()
	if err != nil {
		errCh <- err
		return
	}

	func() {
		for {
			_, err = stdout.Read(buff)
			if err != nil {
				errCh <- err
				return
			}

			ch <- string(buff)
			buff = make([]byte, 1024)

			time.Sleep(time.Duration(interval) * time.Millisecond)
		}
	}()
}

func main() {
	errCh := make(chan error)

	xtitleCh := make(chan string)
	loadCh := make(chan string)
	memoryCh := make(chan string)
	volumeCh := make(chan string)
	timeCh := make(chan string)
	tempCh := make(chan string)
	statusWmCh := make(chan string)

	go Load(10000, loadCh, errCh)
	go Memory(10000, memoryCh, errCh)
	go Volume(5000, volumeCh, errCh)
	go Time(1000, timeCh, errCh)
	go Temp(1000, tempCh, errCh)
	go Xtitle(200, xtitleCh, errCh)
	go StatusWm(200, statusWmCh, errCh)

	for {
		select {
		case err := <-errCh:
			log.Fatal(err)
		default:
			select {
			case xtitle := <-xtitleCh:
				fmt.Print(xtitle)
			case load := <-loadCh:
				fmt.Print(load)
			case memory := <-memoryCh:
				fmt.Print(memory)
			case volume := <-volumeCh:
				fmt.Print(volume)
			case time := <-timeCh:
				fmt.Print(time)
			case temp := <-tempCh:
				fmt.Print(temp)
			case wm := <-statusWmCh:
				fmt.Print(wm)
			}
		}
	}
}
