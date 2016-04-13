package flog

import (
	"testing"
	"os"
	"path"
	"bufio"
	"strconv"
	"time"
	"strings"
)

/**
 * go test -v -bench="." flog
 *
 * @param
 * @return
 *
 */

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
	var expected = LEVEL_ERROR + 1
	if lineNum != expected {
		t.Fatal(lineNum, "not " + strconv.Itoa(expected) + " lines")
	}
	os.Remove(filename)
	//os.RemoveAll(loger.LogPath)
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
	var expected = 1
	if lineNum != expected {
		t.Fatal(lineNum, "not " + strconv.Itoa(expected) + " lines")
	}
	os.Remove(filename)
	//os.RemoveAll(loger.LogPath)
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
	os.Remove(filename)
	//os.RemoveAll(loger.LogPath)
}

//测试日期格式化
func TestDateFormat(t *testing.T) {
	loger := New()
	loger.DateFormat = "Ymd"
	loger.Debug("d", "debug_message")
	filename := path.Join(loger.LogPath, loger.FileName+"."+time.Now().Format("20060102"))
	_,err := os.Open(filename)
	if err != nil && os.IsNotExist(err){
		t.Fatal(err)
	}
	os.Remove(filename)
	//os.RemoveAll(loger.LogPath)
}

//测试文件名模式为file.level
func TestLogModeFileLevel(t *testing.T) {
	loger := New()
	loger.LogMode = LOGMODE_FILE_LEVEL
	loger.Debug("d", "debug_message")
	filename := path.Join(loger.LogPath, loger.FileName+"."+"debug")
	_,err := os.Open(filename)
	if err != nil && os.IsNotExist(err){
		t.Fatal(err)
	}
	os.Remove(filename)
	//os.RemoveAll(loger.LogPath)
}

//测试文件名模式为cate
func TestLogModeCate(t *testing.T) {
	loger := New()
	loger.LogMode = LOGMODE_CATE
	loger.Debug("d", "debug_message")
	filename := path.Join(loger.LogPath, "d")
	_,err := os.Open(filename)
	if err != nil && os.IsNotExist(err){
		t.Fatal(err)
	}
	os.Remove(filename)
	//os.RemoveAll(loger.LogPath)
}

//测试文件名模式为cate.level
func TestLogModeCateLevel(t *testing.T) {
	loger := New()
	loger.LogMode = LOGMODE_CATE_LEVEL
	loger.Debug("d", "debug_message")
	filename := path.Join(loger.LogPath, "d.debug")
	_,err := os.Open(filename)
	if err != nil && os.IsNotExist(err){
		t.Fatal(err)
	}
	os.Remove(filename)
	//os.RemoveAll(loger.LogPath)
}

//测试自定义日志输出的格式和顺序
func TestLogFlags(t *testing.T) {
	loger := New()
	loger.LogFlags = []int{LF_LEVEL,LF_CATE,LF_DATETIME}
	loger.Debug("d", "debug_message")
	filename := path.Join(loger.LogPath, loger.FileName)
	fh,err := os.Open(filename)
	if err != nil {
		t.Fatal(err)
	}
	b := bufio.NewReader(fh)
	line,_,err := b.ReadLine()
	if err != nil {
		t.Fatal(err)
	}

	defer os.Remove(filename)

	//按分隔符分隔line
	messages := strings.Split(string(line),loger.LogFlagSeparator)
	if len(messages) < 4 {
		t.Fatal("Message must have four part at least.")
	}

	if messages[0]!= "DEBUG" || messages[1] != "d" ||  messages[2]!= Date("Y-m-d") {
		t.Fatal("Message does not show as expected.",string(line))
	}
	//os.RemoveAll(loger.LogPath)
}

//测试自定义日志输出的分隔符
func TestLogFlagSeparator(t *testing.T) {
	loger := New()
	loger.LogFlags = []int{LF_LEVEL,LF_CATE,LF_DATETIME}
	loger.LogFlagSeparator = " | "
	loger.Debug("d", "debug_message")
	filename := path.Join(loger.LogPath, loger.FileName)
	fh,err := os.Open(filename)
	if err != nil {
		t.Fatal(err)
	}
	b := bufio.NewReader(fh)
	line,_,err := b.ReadLine()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(filename)
	//按分隔符分隔line
	messages := strings.Split(string(line)," | ")
	if len(messages) < 4 {
		t.Fatal("Message must have four part at least.")
	}

	if messages[0]!= "DEBUG" || messages[1] != "d" ||  !strings.Contains(messages[2],Date("Y-m-d")) {
		t.Fatal("Message does not show as expected.",string(line))
	}


	//os.RemoveAll(loger.LogPath)
}

//测试logFunCallDepth参数
func TestLogFunCallDepth(t *testing.T)  {
	loger := New()
	loger.LogFunCallDepth = 3
	loger.Debug("d","debug_message")
	filename := path.Join(loger.LogPath, loger.FileName)
	fh,err := os.Open(filename)
	if err != nil {
		t.Fatal(err)
	}
	b := bufio.NewReader(fh)
	line,_,err:=b.ReadLine()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(line),"flog_test.go"){
		t.Fatal("Get call func name failed, ",string(line))
	}
	os.Remove(filename)
	//os.RemoveAll(loger.LogPath)
}

//测试异步输出Async
func TestLogFunCallDepth2(t *testing.T)  {
	loger := New()
	loger.SetAsync(10)
	loger.LogFunCallDepth = 3
	loger.Debug("d","debug_message")
	loger.Close()
	filename := path.Join(loger.LogPath, loger.FileName)
	fh,err := os.Open(filename)
	if err != nil {
		t.Fatal(err)
	}
	b := bufio.NewReader(fh)
	line,_,err:=b.ReadLine()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(line),"flog_test.go"){
		t.Fatal("Get call func name failed, ",string(line))
	}
	os.Remove(filename)
	//os.RemoveAll(loger.LogPath)
}

//压力测试写入
func BenchmarkFile(b *testing.B)  {
	loger := New()
	for i:=0;i<b.N;i++{
		loger.Debug("ddd","bech test...")
	}
	os.Remove(path.Join(loger.LogPath, loger.FileName))
	//os.RemoveAll(loger.LogPath)
}

//压力测试异步写入
func BenchmarkFileAsync(b *testing.B)  {
	loger := New()
	loger.SetAsync(0)
	for i:=0;i<b.N;i++{
		loger.Debug("ddd","bech test...")
	}
	//loger.Close()
	os.Remove(path.Join(loger.LogPath, loger.FileName))
	//os.RemoveAll(loger.LogPath)
}