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

var COLORS = map[string]string {
	"{{ColorFg}}"									:"#FFdddddd",	"{{ColorBg}}"									:"#FF101010",
	"{{ColorFocusedOccupiedFg}}"	:"#FFdddddd",	"{{ColorFocusedOccupiedBg}}"	:"#FF404040",
	"{{ColorFocusedFreeFg}}"			:"#FFdddddd",	"{{ColorFocusedFreeBg}}"			:"#FF101010",
	"{{ColorFocusedUrgentFg}}"		:"#FFdddddd",	"{{ColorFocusedUrgentBg}}"		:"#FFe84f4f",
	"{{ColorOccupiedFg}}"					:"#FFdddddd",	"{{ColorOccupiedBg}}"					:"#FF404040",
	"{{ColorFreeFg}}"							:"#FF555555",	"{{ColorFreeBg}}"							:"#FF101010",
	"{{ColorUrgentFg}}"						:"#FFdddddd",	"{{ColorUrgentBg}}"						:"#FFd23d3d",
	"{{ColorLayoutFg}}"						:"#FFdddddd",	"{{ColorLayoutBg}}"						:"#FF101010",
	"{{ColorTitleFg}}"						:"#FF555555",	"{{ColorTitleBg}}"						:"#FF101010",
	"{{ColorStatusFg}}"						:"#FFdddddd",	"{{ColorStatusBg}}"						:"#FF101010",
}

type Element struct {
	name string
	result string
	format string
}

/*
 * Xtitle is used to pass the stream of data from xtitle -s into the ch string.
 * It passes all of its errors through the errCh
 * To invoke this function: go Xtitle(2000, ch, errCh)
*/
func Xtitle(interval int64, ch chan Element, errCh chan error) {
	for {
		out, err := exec.Command("xtitle", "-f", "'%s'").Output()
		if err != nil {errCh <- err;return}

		outStr := strings.TrimSpace(string(out))
		result := "%{F{{ColorTitleFg}}B{{ColorTitleBg}}}" + outStr
		ch <-Element{name:"xtitle",result:string(result)}

		time.Sleep(time.Duration(interval) * time.Millisecond)
	}
}

/*
 * Load is used to pass the results of cat /proc/loadavg into the ch string every 'interval' milliseconds.
 * It passes all of its errors through the errCh
 * To invoke this function: go Load(2000, ch, errCh)
*/
func Load(interval int, ch chan Element, errCh chan error) {
	for {
		out, err := ioutil.ReadFile("/proc/loadavg")
		if err != nil {errCh <- err;return}

		outStr := strings.Fields(string(out))
		result := "%{F{{ColorStatusFg}}B{{ColorStatusBg}}}" + fmt.Sprintf("Load: %s %s %s", outStr[0], outStr[1], outStr[2])
		ch <-Element{name:"load",result:string(result)}

		time.Sleep(time.Duration(interval) * time.Millisecond)
	}
}

/*
 * Memory is used to pass the results of free -m into the ch string every 'interval' milliseconds.
 * It passes all of its errors through the errCh
 * To invoke this function: go Memory(2000, ch, errCh)
*/
func Memory(interval int64, ch chan Element, errCh chan error) {
	for {
		out, err := exec.Command("free", "-m").Output()
		if err != nil {errCh <- err;return}

		outStr := strings.Fields(strings.Split(string(out), "\n")[1])

		total, err := strconv.ParseFloat(outStr[1], 32)
		if err != nil {errCh <- err;return}
		used, err := strconv.ParseFloat(outStr[3], 32)
		if err != nil {errCh <- err;return}

		result := "%{F{{ColorStatusFg}}B{{ColorStatusBg}}}" + fmt.Sprintf("Mem: %.2f/%.2fG", used/1024, total/1024)
		ch <-Element{name:"memory",result:string(result)}

		time.Sleep(time.Duration(interval) * time.Millisecond)
	}
}

/*
 * Volume is used to pass the processed results of pacmd dump into the ch string every 'interval' milliseconds.
 * It passes all of its errors through the errCh
 * To invoke this function: go Volume(2000, ch, errCh)
*/
func Volume(interval int64, ch chan Element, errCh chan error) {
	for {
		out, err := exec.Command("pacmd", "dump").Output()
		if err != nil {errCh <- err;return}

		outStr := strings.Split(string(out), "\n")
		var desiredInt uint64
		for _, value := range outStr {
			if strings.HasPrefix(value, "set-sink-volume") {
				strs := strings.Fields(value)
				desiredInt, err = strconv.ParseUint(strs[len(strs)-1][2:], 16, 32)
				if err != nil {errCh <- err;return}
			}
		}

		result :="%{F{{ColorStatusFg}}B{{ColorStatusBg}}}" +  fmt.Sprintf("Vol: %d%%", int(float64(desiredInt)/655.36))
		ch <-Element{name:"volume",result:result}

		time.Sleep(time.Duration(interval) * time.Millisecond)
	}
}

/*
 * Time is used to pass the formatted results of time.Now() into the ch channel string every 'interval' milliseconds.
 * It passes all of its errors through the errCh
 * To invoke this function: go Time(2000, ch, errCh)
*/
func Time(interval int, ch chan Element, errCh chan error) {
	for {
		result :="%{F{{ColorStatusFg}}B{{ColorStatusBg}}}" +  time.Now().Format("Mon, Jan 2, 2006 15:04:05")
		ch <-Element{name:"time",result:string(result)}

		time.Sleep(time.Duration(interval) * time.Millisecond)
	}
}

/*
 * Temp is used to pass the contents of the path specified into the ch channel string every 'interval' milliseconds.
 * It passes all of its errors through the errCh
 * To invoke this function: go Temp(2000, ch, errCh)
*/
func Temp(interval int, ch chan Element, errCh chan error) {
	for {
		out, err := ioutil.ReadFile("/sys/bus/platform/devices/coretemp.0/temp1_input")
		if err != nil {errCh <- err;return}

		temp, err := strconv.ParseInt(string(out[:len(out)-4]), 10, 16)
		if err != nil {errCh <- err;return}

		result := "%{F{{ColorStatusFg}}B{{ColorStatusBg}}}" + fmt.Sprintf("Temp: %dC", temp)
		ch <-Element{name:"temp",result:result}

		time.Sleep(time.Duration(interval) * time.Millisecond)
	}
}

/*
 * StatusWm takes the stream of information from bspc control --subscribe and pass it to  ch channel string every 'interval' milliseconds.
 * It passes all of its errors through the errCh
 * To invoke this function: go StatusWm(2000, ch, errCh)
*/
func StatusWm(interval int64, ch chan Element, errCh chan error) {
	buff := make([]byte, 1024)
	cmd := exec.Command("bspc", "control", "--subscribe")

	stdout, err := cmd.StdoutPipe()
	if err != nil {errCh <- err;return}

	err = cmd.Start()
	if err != nil {errCh <- err;return}

	for {
		_, err = stdout.Read(buff)
		if err != nil {errCh <- err;return}

		temp := strings.TrimSpace(strings.TrimRight(string(buff), string(byte(0))))
		tempSlice := strings.Split(temp, ":")
		var out string
		for _, str := range tempSlice {
			switch str[0] {
			case 'O':
				out += "%{F{{ColorFocusedOccupiedFg}} U{{ColorFocusedOccupiedFg}} B{{ColorFocusedOccupiedBg}} +u} " + str[1:] + " %{-uB{{ColorBg}}}"
			case 'F':
				out += "%{F{{ColorFocusedFreeFg}} U{{ColorFocusedFreeFg}} B{{ColorFocusedFreeBg}} +u} " + str[1:] + " %{-uB{{ColorBg}}}"
			case 'U':
				out += "%{F{{ColorFocusedUrgentFg}} U{{ColorFocusedUrgentFg}} B{{ColorFocusedUrgentBg}} +u} " + str[1:] + " %{-uB{{ColorBg}}}"
			case 'o':
				out += "%{F{{ColorOccupiedFg}} B{{ColorOccupiedBg}}} " + str[1:] + " %{B{{ColorBg}}}"
			case 'f':
				out += "%{F{{ColorFreeFg}} B{{ColorFreeBg}}} " + str[1:] + " %{B{{ColorBg}}}"
			case 'u':
				out += "%{F{{ColorUrgentFg}} B{{ColorUrgentBg}}} " + str[1:] + " %{B{{ColorBg}}}"
			case 'L':
				out += "%{F{{ColorLayoutFg}} B{{ColorLayoutBg}}} " + strings.ToUpper(string(str[1])) + " %{B{{ColorBg}}}"
			}
		}
		ch <- Element{name:"wm",result:out}
		buff = make([]byte, 1024)

		time.Sleep(time.Duration(interval) * time.Millisecond)
	}
}

func DisplayInfo(format string) func(Element) {
	var outString string
	elements := make(map[string] Element)

	return func(item Element) {
		elements[item.name] = item
		outString = format

		for key, value := range elements {
			outString = strings.Replace(outString, "{{"+key+"}}", value.result, -1)
		}

		for key, value := range COLORS {
			outString = strings.Replace(outString, key,	value, -1)
		}

		fmt.Println(outString)
	}
}

func main() {
	ch := make(chan Element)
	errCh := make(chan error)

	go Load(10000, ch, errCh)
	go Memory(10000, ch, errCh)
	go Volume(5000, ch, errCh)
	go Time(1000, ch, errCh)
	go Temp(1000, ch, errCh)
	go Xtitle(200, ch, errCh)
	go StatusWm(200, ch, errCh)

	format := fmt.Sprint("%{l}{{wm}} %{c}{{xtitle}} %{r}{{volume}} | {{memory}} | {{load}} | {{temp}} | {{time}}")
	//format = fmt.Sprint("%{S0}", format, "%{S1}", format, "%{S2}", format)
	format = fmt.Sprint("%{S0}", format)
	DisplayClosure := DisplayInfo(format)
	for {
		select {
		case err := <-errCh:
			log.Fatal(err)
		default:
			select {
			case item := <-ch:
				go DisplayClosure(item)
			}
		}
	}
}
