/*
listserialports
Used for listing and checking serial device availability. On typical use case,
program needs serial port and before opening serial port check that port is not in use by other instance

This is golang library, check lsserials as example command line tool


This library is only part of my other project.

Later, there will be some features found from setserial utility (if my project requires)
https://github.com/brgl/busybox/blob/master/miscutils/setserial.c
like checking real uart status
*/

package listserialports

import (
	"io/ioutil"
	"os"
	"sort"
	"strings"
)

var procFS = os.DirFS("/proc")
var devFS = os.DirFS("/dev")

var symlinkEval SymlinkEvaluator = FilepathLinkEvaluator{}

func ListOfSerialTTYDriverPrefixes() ([]string, error) {
	f, fErr := procFS.Open("tty/drivers")
	if fErr != nil {
		return nil, fErr
	}
	byt, errRead := ioutil.ReadAll(f)
	if errRead != nil {
		return nil, errRead
	}

	result := []string{}
	rows := strings.Split(string(byt), "\n")
	for _, row := range rows {
		fields := strings.Fields(row)
		if len(fields) == 5 {
			if fields[4] == "serial" {
				result = append(result, fields[1])
			}
		}
	}
	return result, nil
}

// Probe is function for getting list of serial ports on system with details
func Probe(checkSerialBy bool) ([]Entry, error) {
	//Check, are ports in use

	//Old fashioned, just serial port list
	devNames, errDevNames := ListByDev()
	if errDevNames != nil {
		return []Entry{}, errDevNames
	}
	result := make([]Entry, len(devNames))
	for index, devname := range devNames {
		pids, certain, err := FileIsInUseByPids(devname)
		if err != nil {
			return nil, err
		}
		result[index] = Entry{
			DeviceFile: devname,
			UsedByPids: pids,
			Certain:    certain,
		}
	}

	sort.Sort(ByDeviceName(result))
	//Add extra information from /dev/serial
	if checkSerialBy {
		byIDMap, idmapErr := ListById()
		byPathmap, pathmapErr := ListByPath()
		if idmapErr != nil || pathmapErr != nil {
			return result, nil //Not /dev/serial supported, return what got
		}
		for i, entry := range result {
			sId, hazId := byIDMap[entry.DeviceFile]
			if hazId {
				result[i].DeviceByID = sId
			}
			sPath, hazPath := byPathmap[entry.DeviceFile]
			if hazPath {
				result[i].DeviceByPath = sPath
			}
		}
	}

	return result, nil
}
