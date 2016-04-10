package flog

import (
	"testing"
	"os"
	"path"
	"bufio"
	"strconv"
)

//测试所有都默认
func TestDefault(t *testing.T) {
	loger := New()
	loger.Debug("d", "debug_message")
	loger.Info("i", "info_message")
	loger.Warning("w", "warning_message")
	loger.Error("e", "error_message")
	filename := path.Join(loger.LogPath, loger.FileName)
	f, err := os.Open(filename)
	if err != nil {
		t.Fatal(err)
	}
	b := bufio.NewReader(f)
	lineNum := 0
	for {
		line, _, err := b.ReadLine()
		if err != nil {
			break
		}
		if len(line) > 0 {
			lineNum++
		}
	}
	var expected = (LEVEL_ERROR + 1) * 2
	if lineNum != expected {
		t.Fatal(lineNum, "not " + strconv.Itoa(expected) + " lines")
	}
	os.RemoveAll(loger.LogPath)
}

//测试等级
func TestLevel(t *testing.T) {
	loger := New()
	loger.Level = LEVEL_ERROR
	loger.Debug("d", "debug_message")
	loger.Info("i", "info_message")
	loger.Warning("w", "warning_message")
	loger.Error("e", "error_message")
	filename := path.Join(loger.LogPath, loger.FileName)
	f, err := os.Open(filename)
	if err != nil {
		t.Fatal(err)
	}
	b := bufio.NewReader(f)
	lineNum := 0
	for {
		line, _, err := b.ReadLine()
		if err != nil {
			break
		}
		if len(line) > 0 {
			lineNum++
		}
	}
	var expected = 2
	if lineNum != expected {
		t.Fatal(lineNum, "not " + strconv.Itoa(expected) + " lines")
	}
	os.RemoveAll(loger.LogPath)
}

//测试等级
func TestLogPath(t *testing.T) {
	loger := New("/tmp/flog")
	loger.Debug("d", "debug_message")
	filename := path.Join(loger.LogPath, loger.FileName)
	_,err := os.Open(filename)
	if err != nil && os.IsNotExist(err){
		t.Fatal(err)
	}
	os.RemoveAll(loger.LogPath)
}

//测试文件名
func TestFileName(t *testing.T) {
	loger := New()
	loger.FileName = "app.log"
	loger.Debug("d", "debug_message")
	filename := path.Join(loger.LogPath, loger.FileName)
	_,err := os.Open(filename)
	if err != nil && os.IsNotExist(err){
		t.Fatal(err)
	}
	os.RemoveAll(loger.LogPath)
}

//测试日期格式化
func TestDateFormat(t *testing.T) {
	loger := New()
	loger.DateFormat = "20060102"
	loger.Debug("d", "debug_message")
	filename := path.Join(loger.LogPath, loger.FileName)
	_,err := os.Open(filename)
	if err != nil && os.IsNotExist(err){
		t.Fatal(err)
	}
	os.RemoveAll(loger.LogPath)
}