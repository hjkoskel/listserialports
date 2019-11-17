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
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

//ByDeviceName is type for sorting
type ByDeviceName []Entry

func (a ByDeviceName) Len() int      { return len(a) }
func (a ByDeviceName) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByDeviceName) Less(i, j int) bool {
	//Split in between base and number
	baseIarr := strings.FieldsFunc(a[i].DeviceFile, unicode.IsNumber)
	baseJarr := strings.FieldsFunc(a[j].DeviceFile, unicode.IsNumber)

	if len(baseIarr) == 0 || len(baseJarr) == 0 {
		return false
	}
	baseI := baseIarr[0]
	baseJ := baseJarr[0]

	cmp := strings.Compare(baseI, baseJ)
	if cmp != 0 {
		return 0 < cmp
	}
	numberI, _ := strconv.ParseInt(strings.Replace(a[i].DeviceFile, baseI, "", -1), 10, 64)
	numberJ, _ := strconv.ParseInt(strings.Replace(a[j].DeviceFile, baseJ, "", -1), 10, 64)
	return numberI < numberJ

}

//Entry collects final result of serial port status
type Entry struct {
	DeviceFile   string  //With complete path
	UsedByPids   []int64 //In case of conflict, file is open by multiple pids
	Certain      bool
	DeviceByID   string
	DeviceByPath string
}

//Special PID return values
const (
	NOTINUSEPID  = -1 //port is free
	UNCERTAINPID = -2 //can not determine is port free or not (missing rights, non root user)
)

//Equal check
func (p *Entry) Equal(a Entry) bool {
	if len(p.UsedByPids) != len(a.UsedByPids) || p.DeviceFile != a.DeviceFile || p.DeviceByID != a.DeviceByID || p.Certain != a.Certain || p.DeviceByPath != a.DeviceByPath {
		return false
	}
	for i, v := range p.UsedByPids {
		if a.UsedByPids[i] != v {
			return false
		}
	}
	return true
}

//Updates list entries updated. Got new filenames etc..
func Updates(oldEntrys []Entry, newEntrys []Entry) []Entry {
	result := []Entry{}
	for _, oldE := range oldEntrys {
		for _, newE := range newEntrys {
			if oldE.DeviceFile == newE.DeviceFile {
				if !oldE.Equal(newE) {
					result = append(result, newE) //For some reason not equal
				}
			}
		}
	}
	return result
}

//NewEntries lists really new entries
func NewEntries(oldEntrys []Entry, newEntrys []Entry) []Entry {
	result := []Entry{}
	for _, newE := range newEntrys {
		found := false
		for _, oldE := range oldEntrys {
			if newE.DeviceFile == oldE.DeviceFile {
				found = true
				break
			}
		}
		if !found {
			result = append(result, newE)
		}
	}
	return result
}

//HasAny tells is portname matching to any file or link name
func (p *Entry) HasAny(portname string) bool {
	return (portname == p.DeviceFile) || (portname == p.DeviceByID) || (portname == p.DeviceByPath)
}

//ToPrintoutFormat for formatting command line printout
func (p *Entry) ToPrintoutFormat() string { //Tab
	usedByString := ""
	if 0 < len(p.UsedByPids) {
		if len(p.UsedByPids) == 1 {
			usedByString = fmt.Sprintf("(used by PID %v)", p.UsedByPids[0])
		} else {
			usedByString = fmt.Sprintf("(used by PIDs %v)", p.UsedByPids)
		}
	}

	result := fmt.Sprintf("%s %s\n", p.DeviceFile, usedByString)
	if 0 < len(p.DeviceByID) {
		result += fmt.Sprintf("\t%s\n", p.DeviceByID)
	}
	if 0 < len(p.DeviceByPath) {
		result += fmt.Sprintf("\t%s\n", p.DeviceByPath)
	}

	return result
}

/*
listProcesses lists all processes ids running at the moment
*/
func listProcesses() ([]int64, error) {
	procDir, errProc := ioutil.ReadDir("/proc")
	if errProc != nil {
		return []int64{}, errProc
	}
	result := []int64{}
	for _, fil := range procDir {
		if fil.IsDir() {
			pid, pidErr := strconv.ParseInt(fil.Name(), 10, 64)
			if pidErr == nil { //valid.. number only
				result = append(result, pid)
			}
		}
	}
	return result, nil
}

/*
Check all file descriptors owned by that pid
if found link, even with failures, then it is using
if any error and not found file... then it is uncertain
*/
func processUsesFile(pid int64, filename string) (bool, error) {
	dirname := fmt.Sprintf("/proc/%v/fd/", pid)
	list, listErr := ioutil.ReadDir(dirname)
	if listErr != nil {
		return false, listErr //can not determine... might use or not to use
	}
	var latestErr error

	actualFilename, errLink := filepath.EvalSymlinks(filename)
	if errLink != nil {
		actualFilename = filename //Fallback, safer in this way
	}

	for _, fil := range list {
		linkfilename := dirname + fil.Name()
		link, linkErr := filepath.EvalSymlinks(linkfilename)
		if linkErr != nil {
			latestErr = linkErr
		}
		if link == actualFilename {
			return true, nil //USING
		}
	}
	return false, latestErr
}

func fileIsInUseByPids(filename string, pidlist []int64) ([]int64, bool) {
	results := []int64{}
	noErrors := true
	for _, pid := range pidlist {
		uses, errUse := processUsesFile(pid, filename)
		if errUse != nil {
			noErrors = false
		}
		if uses {
			results = append(results, pid)
		}
	}
	return results, noErrors
}

//Exists tell if file exists
func Exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

//FileIsInUse  Direct check. Is certain
func FileIsInUse(filename string) ([]int64, bool, error) {
	if !Exists(filename) {
		return []int64{}, true, fmt.Errorf("File %v does not exists", filename)
	}

	pidlist, pidlistErr := listProcesses()
	if pidlistErr != nil {
		return []int64{}, false, pidlistErr
	}
	//pidlist []int64
	//TODO Kaiva mikä on linkin takana
	actualFilename, errLink := filepath.EvalSymlinks(filename)
	if errLink != nil {
		actualFilename = filename //Fallback, safer in this way
	}

	pidlist, certain := fileIsInUseByPids(actualFilename, pidlist)
	return pidlist, certain, nil
}

//ListDev  lists serial ports named directly under /dev/
func ListDev() ([]string, error) {
	//TODO  parse /proc/tty/drivers
	serialDevPrefixes := []string{"ttyUSB", "ttyACM", "ttyS", "rfcomm", "ttyAMA", "ttySAC", "serial"}

	devdir, errList := ioutil.ReadDir("/dev/")
	if errList != nil {
		return []string{}, errList
	}
	result := []string{}

	for _, fil := range devdir {
		if !fil.IsDir() {
			name := fil.Name()
			for _, prefix := range serialDevPrefixes {
				if strings.HasPrefix(name, prefix) {
					result = append(result, "/dev/"+name)
				}
			}
		}
	}
	return result, nil
}

/*
not all distros have support for /dev/serial

actual device file is key

by-id or by-path
*/
const (
	SERIALBYWHATBYID   = "by-id"
	SERIALBYWHATBYPATH = "by-path"
)

//listByWhat  is used for listing ports by id or path
func listByWhat(what string) (map[string]string, error) {
	dirName := fmt.Sprintf("/dev/serial/%s", what)
	dirlist, errList := ioutil.ReadDir(dirName)
	if errList != nil {
		return nil, errList
	}

	result := make(map[string]string)
	for _, fil := range dirlist {
		if !fil.IsDir() {
			fname := dirName + "/" + fil.Name()
			target, linkErr := filepath.EvalSymlinks(fname)
			target = strings.Replace(target, "../../", "/dev/", 1)
			if linkErr == nil { //Just skip non-links
				result[target] = fname
			}
		}
	}
	return result, nil
}

// Probe is function for getting list of serial ports on system with details
func Probe() ([]Entry, error) {
	//Check, are ports in use
	processList, procListErr := listProcesses()
	if procListErr != nil { //If can not read proc.. it is total fail. wrong OS or something
		return []Entry{}, fmt.Errorf("Reading proc failed %v", procListErr)
	}

	result := map[string]Entry{}

	//Old fashioned, just serial port list
	devNames, errDevNames := ListDev()
	if errDevNames != nil {
		return []Entry{}, errDevNames
	}
	for _, devname := range devNames {
		found := false
		for _, ref := range result {
			if ref.DeviceFile == devname {
				found = true
				break
			}
		}

		if !found {
			pids, certain := fileIsInUseByPids(devname, processList)
			result[devname] = Entry{
				DeviceFile: devname,
				UsedByPids: pids,
				Certain:    certain,
			}
		}
	}

	//If distro supports by-id and by-path. Then group those together
	byIDMap, idmapErr := listByWhat(SERIALBYWHATBYID)
	if idmapErr == nil {
		byPathmap, pathmapErr := listByWhat(SERIALBYWHATBYPATH)
		if pathmapErr == nil {
			//Ok, lets list
			for devname, byidfile := range byIDMap {
				path, hazPath := byPathmap[devname]
				if hazPath {
					pidlist, certain := fileIsInUseByPids(devname, processList)
					result[devname] = Entry{
						DeviceFile:   devname,
						UsedByPids:   pidlist,
						Certain:      certain,
						DeviceByID:   byidfile,
						DeviceByPath: path,
					}
				}
			}
		}
	}

	//Hack
	resultArr := []Entry{}
	for _, v := range result {
		resultArr = append(resultArr, v)
	}

	sort.Sort(ByDeviceName(resultArr))
	return resultArr, nil
}
