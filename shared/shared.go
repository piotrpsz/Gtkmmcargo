package shared

import (
	"io/ioutil"
	"log"
	"os"

	"Gtkmmcargo/tr"
)

func ExistsFile(dirPath string) bool {
	var err error
	var fi os.FileInfo

	if fi, err = os.Stat(dirPath); err != nil {
		if !os.IsNotExist(err) {
			log.Println(err)
		}
		return false
	}
	if !fi.IsDir() && fi.Mode().IsRegular() {
		return true
	}
	return false
}

func ExistsDir(dirPath string) bool {
	var err error
	var fi os.FileInfo

	if fi, err = os.Stat(dirPath); err != nil {
		if !os.IsNotExist(err) {
			log.Println(err)
		}
		return false
	}
	if fi.IsDir() {
		return true
	}
	return false
}

func CreateDirIfNeeded(dirPath string) bool {
	if ExistsDir(dirPath) {
		return true
	}
	if err := os.MkdirAll(dirPath, os.ModePerm); tr.IsOK(err) {
		return true
	}
	return false
}

func CreateFile(fpath string) *os.File {
	if fhandle, err := os.Create(fpath); tr.IsOK(err) {
		return fhandle
	}
	return nil
}

func OpenFile(fpath string) *os.File {
	if fhandle, err := os.Open(fpath); tr.IsOK(err) {
		return fhandle
	}
	return nil
}

func ReadFileContent(fhandle *os.File) []byte {
	if data, err := ioutil.ReadAll(fhandle); tr.IsOK(err) {
		return data
	}
	return nil
}

func OverwriteFileContent(fpath string, data []byte) bool {
	if ExistsFile(fpath) && !RemoveFile(fpath) {
		return false
	}
	if fhandle := CreateFile(fpath); fhandle != nil {
		defer fhandle.Close()

		if _, err := fhandle.Write(data); tr.IsOK(err) {
			return true
		}
	}
	return false
}

func RemoveFile(filePath string) bool {
	if err := os.Remove(filePath); tr.IsOK(err) {
		return true
	}
	return false
}

func NameComponent(fname string) (string, string) {
	if idx := lastPoint(fname); idx != -1 {
		if idx < len(fname)-1 {
			return fname[:idx], fname[idx+1:]
		}
	}
	return fname, ""
}

func PathComponents(fpath string) (string, string) {
	if idx := lastPathSeparator(fpath); idx != -1 {
		if idx < len(fpath)-1 {
			return fpath[:idx], fpath[idx+1:]
		}
		return fpath, ""
	}
	return "", fpath
}

func lastPathSeparator(fpath string) int {
	i := len(fpath)
	for i > 0 {
		i--
		if fpath[i] == os.PathSeparator {
			return i
		}
	}
	return -1
}

func lastPoint(fname string) int {
	i := len(fname)
	for i > 0 {
		i--
		if fname[i] == '.' {
			return i
		}
	}
	return -1
}
