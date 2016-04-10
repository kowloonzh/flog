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
	LEVEL_ERROR:"erro",
}

//文件名模式
const (
	LOGMODE_FILE = iota        //以FileName做文件名
	LOGMODE_FILE_LEVEL        //以filename+level做文件名
	LOGMODE_CATE            //以分类做文件名
	LOGMODE_CATE_LEVEL        //以分类+level做文件名
)

//日志结构体
type LogMsg struct {
	logTime  time.Time
	level    int    //日志等级
	category string //日志分类
	message  string //日志内容
}

/**
 * 文件日志
 */
type Flog struct {
	mu              sync.Mutex
	Level           int    //日志等级
	LogMode         int    //日志文件名模式
						   //LogFlag         int    //日志内容模式
	LogPath         string //日志文件的根目录
	DateFormat      string //文件按格式化
	ArchiveName     string //归档目录 default:archive
	FileName        string //文件名
	LogFunCallDepth int    //获取调用函数的层级
						   //日志logger相关
	logerMap        map[string]*log.Logger
	fhMap           map[string]*os.File
						   //异步写相关
	msgChan         chan *LogMsg
	signalChan      chan string
	async           bool
	wg              sync.WaitGroup
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
}

/**
 * 设置异步写日志
 *
 * @param capacity int 消息缓冲容量
 * @return *Flog
 *
 */
func (this *Flog ) SetAsync(capacity int) *Flog {
	this.async = true
	if capacity <= 0 {
		capacity = 1 << 16  //65536
		fmt.Println(capacity)
	}
	//初始化chan
	this.msgChan = make(chan *LogMsg, capacity)
	this.signalChan = make(chan string, 1)
	//阻塞goroutine
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
	//如果是异步,先写入msgChan
	if this.async {
		this.msgChan <- msg
	}else {
		this.writeMsg(msg)
	}
	//logger.Output(3, this.formatMessage(msg))
}

func (this *Flog ) writeMsg(msg *LogMsg) {
	filename := this.getFilename(msg)
	//fmt.Println(filename)
	logger, err := this.getLogger(filename)
	if err != nil {
		fmt.Println("Error: fail to get logger by filename", filename)
		return
	}
	logger.Print(this.formatMessage(msg))
}

//格式化消息 日期 文件位置 [等级] [类别] 消息  @todo 定制格式化输出
func (this *Flog ) formatMessage(msg *LogMsg) string {
	_, file, line, ok := runtime.Caller(this.LogFunCallDepth)
	if !ok {
		file = "???"
		line = 0
	}
	fileMsg := file + ":" + strconv.FormatInt(int64(line), 10)
	return fmt.Sprintf("%s %s [%s][%s]\n%s\n", msg.logTime.Format("2006-01-02 15:04:05"), fileMsg, this.getLevelName(msg.level), msg.category, msg.message)
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
		nowDate := time.Now().Format(this.DateFormat)
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
	}
	//@todo check logger exist
	logger := this.logerMap[filename]
	return logger, nil

}