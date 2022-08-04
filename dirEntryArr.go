/*
dirEntryArr is utility for sorting directory listings
*/
package listserialports

import (
	"fmt"
	"io/fs"
	"path"
	"strconv"
)

type DirEntryArr []fs.DirEntry

func ReadDirEntryArr(filesys fs.FS, name string) (DirEntryArr, error) {

	/*	stat, errstat := fs.Stat(filesys, name)
		if errstat != nil {
			fmt.Printf("STATERR %s\n", errstat.Error())
		}
		fmt.Printf("stat=%#v errstat=%#v\n", stat, errstat)*/

	entries, errRead := fs.ReadDir(filesys, name)
	return DirEntryArr(entries), errRead
}

func (p *DirEntryArr) NumberFiles() []int64 {
	result := []int64{}
	for _, dEntry := range *p {
		n, nErr := strconv.ParseInt(dEntry.Name(), 10, 64)
		if nErr == nil { //valid.. number only
			result = append(result, n)
		}
	}
	return result
}

func (p *DirEntryArr) FilesOnly(addPath string) []string {
	result := []string{}
	for _, finfo := range *p {
		if !finfo.IsDir() {
			name := finfo.Name()
			if 0 < len(addPath) {
				name = path.Join(addPath, name)
			}
			result = append(result, name)
		}
	}
	return result
}

func (p *DirEntryArr) DirsOnly(addPath string) []string {
	result := []string{}
	for _, finfo := range *p {
		fmt.Printf("Checking %v  isdir=%v\n", finfo.Name(), finfo.IsDir())
		if finfo.IsDir() {
			name := finfo.Name()
			if 0 < len(addPath) {
				name = path.Join(addPath, name)
			}
			result = append(result, name)
		}
	}
	return result
}

func (p *DirEntryArr) ResolvedNames(pathname string, skipUnresolved bool) ([]string, []error) {
	result := []string{}
	errlist := []error{}
	for _, finfo := range *p {
		pathAndFilename := path.Join(pathname, finfo.Name())

		solved, err := symlinkEval.Eval(pathAndFilename)
		if err == nil {
			result = append(result, solved)
		} else {
			if !skipUnresolved {
				result = append(result, solved)
			}
		}
	}
	if len(errlist) == 0 {
		return result, nil
	}
	return result, errlist
}
