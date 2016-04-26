# flog
file logger for go
这是一个go语言写入文件的日志库,习作,如果有问题或者有更好的想法欢迎联系,kowloonzh@gmail.com

可以通过下面的方式进行安装

```go get github.com/KowloonZh/flog```


## 如何使用

### 最简单的方式

首先引入包

```
import (
	"github.com/KowloonZh/flog"
)
```

无需任何设置,然后我们就可以使用了

```
package main

import (
	"github.com/KowloonZh/flog"
)

func main()  {
	loger := flog.New()
	//第一个参数是分类名,后面的参数是日志内容
	loger.Debug("test","debug message")
	loger.Info("test","info message")
	loger.Warning("test","warning message")
	loger.Error("test","error message")
}
```

默认的日志路径是当前目录下的logs目录,不存在会自动创建,请确保有写权限,默认的文件名是flog.log


### 设置日志目录
```
....
func main()  {
    //设置日志目录为/data/logs
	loger := flog.New("/data/logs")

	loger.Debug("test","debug message")

}
```


### 设置日志文件名
```
....
func main()  {
	loger := flog.New("/data/logs")

	//设置日志文件名为app.log
    loger.FileName = "app.log"

	loger.Debug("test","debug message")

}
```

### 设置日志文件时间后缀
时间后缀支持 YmdHis ,分别表示年月日时分秒,eg. Y-m-d, Ymd, YmdH, etc...
```
....
func main()  {
	loger := flog.New("/data/logs")
    loger.FileName = "app.log"

    //设置时间格式之后,产生的日志文件后带上时间后缀,eg. app.log.20160410
    loger.DateFormat = "Ymd"

	loger.Debug("test","debug message")

}
```

### 设置日志文件名模式
属性 LogMode 用来指定文件名模式,目前支持四种文件名

- LOGMODE_FILE eg. app.log
- LOGMODE_FILE_LEVEL eg. app.log.debug
- LOGMODE_CATE eg. test test为下例中Debug的第一个参数
- LOGMODE_CATE_LEVEL eg. test.debug

```
....
func main()  {
	loger := flog.New("/data/logs")
    loger.DateFormat = "Ymd"

    //设置文件名模式为cate+level,最后生成日志文件名 test.debug.20160410
    loger.LogMode = flog.LOGMODE_CATE_LEVEL
	loger.Debug("test","debug message")

}
```

### 设置日志异步写入


```
....
func main()  {
	loger := flog.New("/data/logs")

	//开启异步,参数为支持的并发写入数
	loger.SetAsync(100000)
	//如果执行的go的脚本,需要在脚本结束的时候关闭日志把缓冲区的日志写完
	defer loger.Close()

	loger.Debug("test","debug message")

}
```

### 设置日志中显示调用调用日志的文件名以及行数


```
....
func main()  {
	loger := flog.New("/data/logs")
    //默认值是3,直接使用flog不需要设置该值,如果你对flog进行了封装的话,可自行调整
    loger.LogFunCallDepth = 3

	loger.Debug("test","debug message")

}
```

### 设置日志内容的格式以及顺序以及分隔符
日志内容的格式 LogFlags 数组里面的元素可以自定义以下几种

-	LF_DATETIME   输出日期 eg. 2016-04-06 03:05:01
-	LF_SHORTFILE  输出文件名和行号 eg. test.go:22
-	LF_LONGFILE   输出文件绝对路径和行号  eg. /tmp/test.go:22
-	LF_CATE       输出分类 eg. test
-	LF_LEVEL      输出等级 eg. DEBUG

LogFlags 默认为 [LF_DATETIME, LF_LONGFILE, LF_CATE, LF_LEVEL]

日志内容分隔符 LogFlagSeparator 默认为空格,可以自定义其他字串,eg. "|"

```
....
func main()  {
	loger := flog.New("/data/logs")

    //设置日志内容以及顺序为 等级 分类 时间
	loger.LogFlags = []int{flog.LF_LEVEL,flog.LF_CATE,flog.LF_DATETIME}
	//设置日志内容的分隔符为 " | "
    loger.LogFlagSeparator = " | "

    loger.Debug("d", "debug_message")

}
```

### 命令行日志

```
	loger := flog.New("/data/logs")

    //设置命令行同步输出日志,默认为false,不同级别会输出不同的颜色
    loger.OpenConsoleLog = true

    loger.Debug("d", "debug_message")
    loger.Info("i", "info_message")
    loger.Warning("w", "warning_message")
    loger.Error("e", "error_message")

```

### 日志切割

```
....
func main()  {
	loger := flog.New("/data/logs")

    //设置切割日志的大小,单位KB
    loger.LogRotateSize = 10*1024  //10MB

    loger.Debug("d", "debug_message")

}
```

### 归档
归档涉及三个参数
NeedArchive  是否需要归档,默认为false
ArchivePath  归档目录,默认为archive,表示LogPath目录下的archive目录
LogKeepDay   日志保留的天数,默认为7天

> 考虑到性能问题,归档采用的是goroutine的方式调用,测试的时候可能会出现主线程先退出,归档未完成的情况,可以在主程序中加time.Sleep()来查看归档效果


```
....
func main()  {
	loger := flog.New("/data/logs")

    //开启归档功能
    loger.NeedArchive = true

    //设置归档目录
    loger.ArchivePath = "archive"

    //设置归档日志保留的天数
    loger.LogKeepDay = 30

    loger.Debug("d", "debug_message")

}
```