package listserialports

import (
	"fmt"
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
