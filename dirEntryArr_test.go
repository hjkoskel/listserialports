package listserialports

import (
	"fmt"
	"sort"
	"strings"
	"testing"
	"testing/fstest"
)

type TestLinkEvaluator struct{}

func (p TestLinkEvaluator) Eval(filename string) (string, error) {
	fmt.Printf("\nTEST LINK EVALUATOR CALLED\n\n")
	if strings.HasPrefix(filename, "/dev/ttyS") || strings.HasPrefix(filename, "/dev/ttyUSB") {
		return filename, nil
	}
	if filename == "/dev/serial/by-id/usb-1a86_USB2.0-Serial-if00-port0" || filename == "/dev/serial/by-path/pci-0000:05:00.3-usb-0:2:1.0-port0" {
		return "/dev/ttyUSB1", nil
	}
	return "", fmt.Errorf("Unit test failed. File %v not defined", filename)
}

func TestDevListing(t *testing.T) {

	symlinkEval = TestLinkEvaluator{}

	devFS = fstest.MapFS{
		"aaa":             {},
		"bbb":             {},
		"dirhere/subfile": {},
		"69":              {},
		"42":              {},
	}

	arr, errArr := ReadDirEntryArr(devFS, ".")
	if errArr != nil {
		t.Error(errArr)
	}
	fmt.Printf("Arr is now (%v items)\n", len(arr))
	for i, item := range arr {
		fmt.Printf("%v: %#v dir=%v\n", i, item.Name(), item.IsDir())
	}

	numbers := arr.NumberFiles()
	if len(numbers) != 2 {
		if (numbers[0] == 42 && numbers[1] == 69) || (numbers[0] == 69) && (numbers[1] == 42) {

		} else {
			t.Errorf("Number files test failed %#v", numbers)
		}
	}

	onlyfiles := arr.FilesOnly("prefix")
	sort.Strings(onlyfiles)
	s := strings.Join(onlyfiles, "|")
	if s != "prefix/42|prefix/69|prefix/aaa|prefix/bbb" {
		t.Errorf("Onlyfiles fail arr= %#v", onlyfiles)
	}

	onlydirs := arr.DirsOnly("prefix")
	s = strings.Join(onlydirs, "|")
	if s != "prefix/dirhere" {
		t.Errorf("dirsonly fail, %s", s)
	}
}
