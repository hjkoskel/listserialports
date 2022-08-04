/*
checking what files are open by specific pid
and what pids are using specific file
*/
package listserialports

import (
	"fmt"
	"io/fs"
)

func ListOpenFilesByPid(pid int64) ([]string, error) {
	//FOR some reason  pidnumber/fd does not work "invalid argument"-error
	pidFS, errSub := fs.Sub(procFS, fmt.Sprintf("%v", pid))
	if errSub != nil {
		return nil, errSub
	}
	pidFS, errSub = fs.Sub(pidFS, "fd")
	if errSub != nil {
		return nil, errSub
	}

	list, listErr := ReadDirEntryArr(pidFS, ".")
	if listErr != nil {
		return nil, errSub
	}

	//ResolvedNames requires full path
	names, errList := list.ResolvedNames(fmt.Sprintf("/proc/%v/fd", pid), false)
	if errList != nil {
		return names, fmt.Errorf("Error resolving files for pid %v err=%#v", pid, errList)
	}
	return names, nil
}

func FileIsInUseByPids(filename string) ([]int64, bool, error) {
	procDir, errProc := ReadDirEntryArr(procFS, ".")
	if errProc != nil {
		return []int64{}, false, errProc
	}

	pidlist := procDir.NumberFiles()

	results := []int64{}
	noErrors := true

	//Eval filename once

	actualFilename, errLink := symlinkEval.Eval(filename)
	if errLink != nil {
		actualFilename = filename //Fallback, safer in this way
	}
	for _, pid := range pidlist {
		evaluated, evalErrors := ListOpenFilesByPid(pid)

		for _, name := range evaluated {
			if name == actualFilename {
				results = append(results, pid)
			}
		}
		if evalErrors != nil {
			noErrors = false
		}
	}
	return results, noErrors, nil
}
