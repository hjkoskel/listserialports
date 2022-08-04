/*
not all distros have support for /dev/serial
actual device file is key
by-id or by-path
*/
package listserialports

import (
	"os"
	"strings"
)

var serialFS = os.DirFS("/dev/serial/")

func ListById() (map[string]string, error) {
	lst, errList := ReadDirEntryArr(serialFS, "by-id")
	if errList != nil {
		return nil, errList
	}
	dirlist := lst.FilesOnly("/dev/serial/by-id")

	result := make(map[string]string)
	for _, fname := range dirlist {
		target, linkErr := symlinkEval.Eval(fname)
		if linkErr == nil { //Just skip non-links
			result[strings.Replace(target, "../../", "/dev/", 1)] = fname
		}
	}
	return result, nil
}

func ListByPath() (map[string]string, error) {
	lst, errList := ReadDirEntryArr(serialFS, "by-path")
	if errList != nil {
		return nil, errList
	}
	dirlist := lst.FilesOnly("/dev/serial/by-path")

	result := make(map[string]string)
	for _, fname := range dirlist {
		target, linkErr := symlinkEval.Eval(fname)
		if linkErr == nil { //Just skip non-links
			result[strings.Replace(target, "../../", "/dev/", 1)] = fname
		}
	}
	return result, nil
}

//ListDev  lists serial ports named directly under /dev/  old fashioned (or minimal system)
func ListByDev() ([]string, error) {
	serialDevPrefixes, errPrefixes := ListOfSerialTTYDriverPrefixes()
	if errPrefixes != nil {
		return []string{}, nil
	}
	devdir, errList := ReadDirEntryArr(devFS, ".")

	if errList != nil {
		return []string{}, errList
	}
	filelist := devdir.FilesOnly("/dev/")

	result := []string{}
	for _, name := range filelist {
		for _, prefix := range serialDevPrefixes {
			if strings.HasPrefix(name, prefix) {
				result = append(result, name)
			}
		}
	}
	return result, nil
}
