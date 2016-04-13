package flog

import (
	"sync"
	"log"
	"os"
	"fmt"
	"time"
	"path"
	"runtime"
	"strconv"
	"strings"
)

const (
	LEVEL_DEBUG = iota
	LEVEL_INFO
	LEVEL_WARNING
	LEVEL_ERROR
)

var levels = map[int]string{
	LEVEL_DEBUG:"debug",
	LEVEL_INFO:"info",
	LEVEL_WARNING:"warning",
	LEVEL_ERROR:"error",
}

//文件名模式
const (
	LOGMODE_FILE = iota        //以FileName做文件名
	LOGMODE_FILE_LEVEL        //以filename+level做文件名
	LOGMODE_CATE            //以分类做文件名
	LOGMODE_CATE_LEVEL        //以分类+level做文件名
)

//输出格式
const (
	LF_DATETIME = 1 << iota            //输出日期
	LF_SHORTFILE                //输出文件名+行号
	LF_LONGFILE                    //输出文件绝对路径+行号
	LF_CATE                        //输出分类
	LF_LEVEL                    //输出等级
)

//日志结构体
type LogMsg struct {
	logTime   time.Time
	level     int    //日志等级
	category  string //日志分类
	message   string //日志内容
	formatMsg string //格式化之后的内容
}

//用来格式化时间
var TimeFormatMap = map[string]string{
	"Y":"2006",
	"m":"01",
	"d":"02",
	"H":"15",
	"i":"04",
	"s":"05",
}

/**
 * 类似php的date()函数
 *
 * @param
 * @return
 *
 */
func Date(format string, timestamp ...int64) string {
	newFormat := format
	for k, v := range TimeFormatMap {
		newFormat = strings.Replace(newFormat, k, v, 1)
	}
	var tm time.Time
	if len(timestamp) > 0 {
		tm = time.Unix(timestamp[0], 0)
	}else {
		tm = time.Now()
	}
	return tm.Format(newFormat)
}

/**
 * 文件日志
 */
type Flog struct {
	mu               sync.Mutex
	Level            int                    //日志等级
	LogMode          int                    //日志文件名模式
	LogPath          string                 //日志文件的根目录
	FileName         string                 //文件名
	DateFormat       string                 //文件按格式化 YmdHis
	LogFlags         []int                  //日志输出的格式以及顺序
	LogFlagSeparator string                 //日志输出的分隔符
	ArchiveName      string                 //归档目录 default:archive
	LogFunCallDepth  int                    //获取调用函数的层级
											/**
											 * 日志logger相关
											 */
	logerMap         map[string]*log.Logger //filename:log.Logger
	fhMap            map[string]*os.File    //filename:os.File
											/**
											 * 异步写相关
											 */
	msgChan          chan *LogMsg           //日志chan
	signalChan       chan string            //信号chan 包括flush 和 close
	async            bool                   //是否开启异步
	wg               sync.WaitGroup

											/**
											 * 日志切割和归档相关
											 */
	LogRotateSize    int                    //日志切割的文件大小最大值,单位KB
}

/**
 * 实例化一个文件日志,并初始化属性
 *
 * @param logPath string 日志目录
 * @return *Flog
 *
 */
func New(logPath ...string) *Flog {
	flog := new(Flog)
	flog.init()
	if len(logPath) > 0 {
		flog.LogPath = logPath[0]
	}
	return flog
}

func (this *Flog ) init() {
	//对文件操作的map和日志处理map初始化
	if len(this.fhMap) == 0 {
		this.fhMap = make(map[string]*os.File)
		this.logerMap = make(map[string]*log.Logger)
	}
	if len(this.LogPath) == 0 {
		this.LogPath = "logs"
	}

	if len(this.ArchiveName) == 0 {
		this.ArchiveName = "archive"
	}

	if len(this.FileName) == 0 {
		this.FileName = "flog.log"
	}

	if this.LogMode == 0 {
		this.LogMode = LOGMODE_FILE
	}

	if this.LogFunCallDepth == 0 {
		this.LogFunCallDepth = 3
	}

	if len(this.LogFlags) == 0 {
		this.LogFlags = []int{LF_DATETIME, LF_LONGFILE, LF_CATE, LF_LEVEL}
	}

	if this.LogRotateSize == 0 {
		this.LogRotateSize = 100 << 10  //100M
	}
}

/**
 * 设置异步写日志
 *
 * @param capacity int 消息缓冲容量
 * @return *Flog
 *
 */
func (this *Flog ) SetAsync(capacity int64) *Flog {
	this.async = true
	if capacity <= 0 {
		capacity = 1 << 20  //1048576
		//fmt.Println(capacity)
	}
	//初始化chan
	this.msgChan = make(chan *LogMsg, capacity)
	this.signalChan = make(chan string, 1)
	//异步执行日志收集
	this.wg.Add(1)
	go this.collect()
	return this
}

func (this *Flog ) collect() {
	over := false

	for {
		select {
		//写入
		case msg := <-this.msgChan:
			this.writeMsg(msg)
		//接受flush 和 close 两个信号
		case signal := <-this.signalChan:
			this.flush()
			if signal == "close" {
				over = true
			}
			this.wg.Done()
		}
		if over {
			break
		}
	}
}

//将缓冲区的消息全部写入
func (this *Flog ) flush() {
	for {
		if len(this.msgChan) > 0 {
			msg := <-this.msgChan
			this.writeMsg(msg)
			continue
		}
		break
	}
}

//关闭日志并清空缓冲区消息
func (this *Flog ) Close() {
	if this.async {
		this.signalChan <- "close"
		//等待执行完成
		this.wg.Wait()
		close(this.msgChan)
		close(this.signalChan)
	}else {
		this.flush()
	}
	this.fhMap = nil
	this.logerMap = nil
}

//清空缓冲区消息
func (this *Flog ) Flush() {
	if this.async {
		this.signalChan <- "flush"
		this.wg.Wait()
		this.wg.Add(1)
		return
	}
	this.flush()
}

func (this *Flog ) Debug(category string, v ...interface{}) {
	if LEVEL_DEBUG >= this.Level {
		this.log(category, LEVEL_DEBUG, v...)
	}
}

func (this *Flog ) Info(category string, v ...interface{}) {
	if LEVEL_INFO >= this.Level {
		this.log(category, LEVEL_INFO, v...)
	}
}

func (this *Flog ) Warning(category string, v ...interface{}) {
	if LEVEL_WARNING >= this.Level {
		this.log(category, LEVEL_WARNING, v...)
	}
}

func (this *Flog ) Error(category string, v ...interface{}) {
	if LEVEL_ERROR >= this.Level {
		this.log(category, LEVEL_ERROR, v...)
	}
}

func (this *Flog ) log(category string, level int, v ...interface{}) {
	//执行初始化默认值
	this.init()
	msg := &LogMsg{
		logTime:time.Now(),
		level:level,
		category:category,
		message:fmt.Sprintln(v...),
	}
	//格式化message
	msg.formatMsg = this.formatMessage(msg)
	//如果是异步,先写入msgChan
	if this.async {
		this.msgChan <- msg
	}else {
		this.writeMsg(msg)
	}
}

//@todo 可能会出现多个人拿到同一个logger,但是某一个人拿到的时候执行了rotate导致其他人没法再写
func (this *Flog ) writeMsg(msg *LogMsg) {
	filename := this.getFilename(msg)
	//fmt.Println(filename)
	logger, err := this.getLogger(filename)
	if err != nil {
		fmt.Println("Error: fail to get logger by filename", filename)
		return
	}
	logger.Print(msg.formatMsg)
}

//格式化消息 日期 文件位置 等级 类别 消息
func (this *Flog ) formatMessage(msg *LogMsg) string {
	_, file, line, ok := runtime.Caller(this.LogFunCallDepth)
	if !ok {
		file = "???"
		line = 0
	}
	formatStr := make([]interface{}, 0)
	for _, flag := range this.LogFlags {
		switch flag {
		case LF_DATETIME:
			formatStr = append(formatStr, msg.logTime.Format("2006-01-02 15:04:05"))
		case LF_LEVEL:
			formatStr = append(formatStr, strings.ToUpper(this.getLevelName(msg.level)))
		case LF_CATE:
			formatStr = append(formatStr, msg.category)
		case LF_LONGFILE:
			formatStr = append(formatStr, file + ":" + strconv.Itoa(line))
		case LF_SHORTFILE:
			short := file
			for i := len(file) - 1; i > 0; i-- {
				if file[i] == '/' {
					short = file[i + 1:]
					break
				}
			}
			formatStr = append(formatStr, short + ":" + strconv.Itoa(line))
		}
	}
	formatStr = append(formatStr, msg.message)

	if len(this.LogFlagSeparator) == 0 {
		this.LogFlagSeparator = " "
	}

	s := strings.TrimPrefix(strings.Repeat(this.LogFlagSeparator + "%s", len(formatStr)), this.LogFlagSeparator)
	return fmt.Sprintf(s, formatStr...)
}

//根据等级获取等级的label
func (this *Flog ) getLevelName(level int) string {
	return levels[level]
}

//根据消息获取文件名
func (this *Flog ) getFilename(msg *LogMsg) string {
	filename := ""
	levelName := this.getLevelName(msg.level)
	switch this.LogMode {
	case LOGMODE_FILE:
		filename = this.FileName
	case LOGMODE_FILE_LEVEL:
		filename = this.FileName + "." + levelName
	case LOGMODE_CATE:
		filename = msg.category
	case LOGMODE_CATE_LEVEL:
		filename = msg.category + "." + levelName
	default:
		filename = this.FileName
	}
	if len(this.DateFormat) > 0 {
		nowDate := Date(this.DateFormat)
		//nowDate := time.Now().Format(this.DateFormat)
		filename = filename + "." + nowDate
	}
	return filename
}

//获取文件名对应的logger
func (this *Flog ) getLogger(filename string) (*log.Logger, error) {
	//如果目录不存在则创建
	os.MkdirAll(this.LogPath, os.ModePerm)
	filePath := path.Join(this.LogPath, filename)

	this.mu.Lock()
	defer this.mu.Unlock()

	//先去fhMap里面查看
	fh, ok := this.fhMap[filename]
	if !ok || (fh != nil && fh.Name() != filePath) {
		if fh != nil {
			fh.Close()
		}
		fh, err := os.OpenFile(filePath, os.O_RDWR | os.O_APPEND | os.O_CREATE, os.ModePerm)
		if err != nil {
			return nil, err
		}
		this.fhMap[filename] = fh
		this.logerMap[filename] = log.New(fh, "", 0)
	}else {
		err := this.rotate(fh)
		if err != nil {
			return nil, err
		}
	}
	//@todo check logger exist
	logger := this.logerMap[filename]
	return logger, nil

}

//执行切割
func (this *Flog ) rotate(file *os.File) error {
	//如果不需要切割
	if !this.needRotate(file) {
		return nil
	}
	//先关闭
	err := file.Close()
	if err != nil {
		return err
	}
	filePath := file.Name()
	_, filename := path.Split(filePath)
	//再重命名
	newPath := filePath + "." + Date("His")
	err = os.Rename(filePath, newPath)
	if err != nil {
		return err
	}
	//再生成新的logger和fh
	fh, err := os.OpenFile(filePath, os.O_RDWR | os.O_APPEND | os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}
	//替换旧的
	this.fhMap[filename] = fh
	this.logerMap[filename] = log.New(fh, "", 0)
	return nil
}

//是否需要切割日志
func (this *Flog ) needRotate(file *os.File) bool {
	//获取文件的大小
	info, err := file.Stat()
	if err != nil {
		log.Println("get file stat err,", file.Name(), err)
		return false
	}
	if info.Size() >= int64(this.LogRotateSize << 10) {
		return true
	}

	return false

}