package logger

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"
)

var defaultlog *logBean = getdefaultLogger()
var skip int = 4

type logger struct {
	lb *logBean
}

func (this *logger) SetConsole(isConsole bool) {
	this.lb.setConsole(isConsole)
}

func (this *logger) SetLevel(_level level) {
	this.lb.setLevel(_level)
}

func (this *logger) SetRollingFile(fileDir, fileName string, maxNumber int32, maxSize int64, _unit unit) {
	this.lb.setRollingFile(fileDir, fileName, maxNumber, maxSize, _unit)
}

func (this *logger) SetRollingDaily(fileDir, fileName string, maxFileCount int32) {
	this.lb.setRollingDaily(fileDir, fileName, maxFileCount)
}

func (this *logger) SetLevelFile(level level, dir, fileName string) {
	this.lb.setLevelFile(level, dir, fileName)
}

type logBean struct {
	mu              *sync.Mutex
	logLevel        level
	maxFileSize     int64
	maxFileCount    int32
	consoleAppender bool
	rolltype        _ROLLTYPE
	id              string
	d, i, w, e, f   string //id
	lg              *log.Logger
}

type fileBeanFactory struct {
	fbs map[string]*fileBean
	mu  *sync.RWMutex
}

var fbf = &fileBeanFactory{fbs: make(map[string]*fileBean, 0), mu: new(sync.RWMutex)}

func (this *fileBeanFactory) add(dir, filename string, maxsize int64, maxfileCount int32) {
	this.mu.Lock()
	defer this.mu.Unlock()
	id := md5str(fmt.Sprint(dir, filename))
	if _, ok := this.fbs[id]; !ok {
		this.fbs[id] = newFileBean(dir, filename, maxsize, maxfileCount)
	}
}

func (this *fileBeanFactory) get(id string) *fileBean {
	this.mu.RLock()
	defer this.mu.RUnlock()
	return this.fbs[id]
}

type fileBean struct {
	id           string
	dir          string
	filename     string
	_date        *time.Time
	mu           *sync.RWMutex
	logfile      *os.File
	lg           *log.Logger
	filesize     int64
	maxFileSize  int64
	maxFileCount int32
}

func getdefaultLogger() (lb *logBean) {
	lb = &logBean{}
	lb.mu = new(sync.Mutex)
	lb.setConsole(true)
	lb.lg = &log.Logger{}
	lb.lg.SetOutput(os.Stdout)
	return
}

func (this *logBean) setConsole(isConsole bool) {
	this.consoleAppender = isConsole
}

func (this *logBean) setLevelFile(level level, dir, fileName string) {
	key := md5str(fmt.Sprint(dir, fileName))
	switch level {
	case _DEBUG:
		this.d = key
	case _INFO:
		this.i = key
	case _WARN:
		this.w = key
	case _ERROR:
		this.e = key
	case _FATAL:
		this.f = key
	default:
		return
	}
	fbf.add(dir, fileName, this.maxFileSize, this.maxFileCount)
}

func (this *logBean) setLevel(_level level) {
	this.logLevel = _level
}

func (this *logBean) setRollingFile(fileDir, fileName string, maxNumber int32, maxSize int64, _unit unit) {
	this.mu.Lock()
	defer this.mu.Unlock()
	if maxNumber > 0 {
		this.maxFileCount = maxNumber
	} else {
		this.maxFileCount = 1<<31 - 1
	}
	this.maxFileSize = maxSize * int64(_unit)
	this.rolltype = _ROLLFILE
	mkdirlog(fileDir)

	this.id = md5str(fmt.Sprint(fileDir, fileName))
	fbf.add(fileDir, fileName, this.maxFileSize, this.maxFileCount)
}

func (this *logBean) setRollingDaily(fileDir, fileName string, maxFileCount int32) {
	this.rolltype = _DAILY
	mkdirlog(fileDir)
	this.id = md5str(fmt.Sprint(fileDir, fileName))
	fbf.add(fileDir, fileName, 0, maxFileCount)
}

func (this *logBean) console(lvl, s string) {
	if this.consoleAppender {
		_, file, line, _ := runtime.Caller(skip)
		short := file
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' {
				short = file[i+1:]
				break
			}
		}
		file = short
		this.lg.Println(time.Now().Format(_TIMEFORMAT), lvl, fmt.Sprintf("%s:%d", file, line), "-", s)
	}
}

func (this *logBean) log(level string, text string, v ...interface{}) {
	defer catchError("log")
	s := fmt.Sprintf(text, v...)
	var lg *fileBean = fbf.get(this.id)
	var _level = _ALL
	switch level {
	case "[DEBUG]":
		if this.d != "" {
			lg = fbf.get(this.d)
		}
		_level = _DEBUG
	case "[INFO]":
		if this.i != "" {
			lg = fbf.get(this.i)
		}
		_level = _INFO
	case "[WARN]":
		if this.w != "" {
			lg = fbf.get(this.w)
		}
		_level = _WARN
	case "[ERROR]":
		if this.e != "" {
			lg = fbf.get(this.e)
		}
		_level = _ERROR
	case "[FATAL]":
		if this.f != "" {
			lg = fbf.get(this.f)
		}
		_level = _FATAL
	}
	if lg != nil {
		this.fileCheck(lg)
		if this.logLevel <= _level {
			if lg != nil {
				writeLen := lg.write(level, s)
				lg.addsize(int64(writeLen))
			}
		}
	} else {
		if this.logLevel <= _level {
			this.console(level, s)
		}
	}
}

func (this *logBean) debug(text string, v ...interface{}) {
	this.log("[DEBUG]", text, v...)
}
func (this *logBean) info(text string, v ...interface{}) {
	this.log("[INFO]", text, v...)
}
func (this *logBean) warn(text string, v ...interface{}) {
	this.log("[WARN]", text, v...)
}
func (this *logBean) error(text string, v ...interface{}) {
	this.log("[ERROR]", text, v...)
}
func (this *logBean) fatal(text string, v ...interface{}) {
	this.log("[FATAL]", text, v...)
}

func (this *logBean) fileCheck(fb *fileBean) {
	defer catchError("fileCheck")
	if this.isMustRename(fb) {
		this.mu.Lock()
		defer this.mu.Unlock()
		if this.isMustRename(fb) {
			fb.rename(this.rolltype)
		}
	}
}

//--------------------------------------------------------------------------------

func (this *logBean) isMustRename(fb *fileBean) bool {
	switch this.rolltype {
	case _DAILY:
		t, _ := time.Parse(_DATEFORMAT, time.Now().Format(_DATEFORMAT))
		if t.After(*fb._date) {
			return true
		}
	case _ROLLFILE:
		return fb.isOverSize()
	}
	return false
}

func newFileBean(fileDir, fileName string, maxSize int64, maxfileCount int32) (fb *fileBean) {
	t, _ := time.Parse(_DATEFORMAT, time.Now().Format(_DATEFORMAT))
	fb = &fileBean{dir: fileDir, filename: fileName, _date: &t, mu: new(sync.RWMutex)}
	fb.logfile, _ = os.OpenFile(fileDir+"/"+fileName, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	fb.lg = &log.Logger{}
	fb.lg.SetOutput(fb.logfile)
	fb.maxFileSize = maxSize
	fb.maxFileCount = maxfileCount
	fb.filesize = fileSize(fileDir + "/" + fileName)
	fb._date = &t
	return
}

func (this *fileBean) rename(rolltype _ROLLTYPE) {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.close()
	nextfilename := ""
	switch rolltype {
	case _DAILY:
		// 当前文件变更的文件名应该为：文件名日期为昨天
		nextfilename = fmt.Sprint(this.dir, "/", this.filename, ".", this._date.Format(_DATEFORMAT))
		if isExist(nextfilename) {
			os.Remove(nextfilename)
		}
	case _ROLLFILE:
		// 当前文件变更的文件名应该为：文件名后缀为1
		nextfilename = fmt.Sprint(this.dir, "/", this.filename, ".1")
		// 如果后缀序号最大的文件存在，删除它
		if maxNoFilename := fmt.Sprint(this.dir, "/", this.filename, ".", this.maxFileCount); isExist(maxNoFilename) {
			os.Remove(maxNoFilename)
		}
		// 滚动文件，后缀序号依次增加1
		for lastNo := this.maxFileCount - 1; lastNo >= 1; lastNo-- {
			if currentFilename := fmt.Sprint(this.dir, "/", this.filename, ".", lastNo); isExist(currentFilename) {
				os.Rename(currentFilename, fmt.Sprint(this.dir, "/", this.filename, ".", (lastNo+1)))
			}
		}
	}
	os.Rename(this.dir+"/"+this.filename, nextfilename)
	this.logfile, _ = os.OpenFile(this.dir+"/"+this.filename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	this.lg = &log.Logger{}
	this.lg.SetOutput(this.logfile)
	this.filesize = fileSize(this.dir + "/" + this.filename)
	t, _ := time.Parse(_DATEFORMAT, time.Now().Format(_DATEFORMAT))
	this._date = &t
	// 按日期滚动的日志文件，进行清理
	if rolltype == _DAILY {
		go func() {
			defer catchError("clean")
			this.clean()
		}()
	}
}

func (this *fileBean) clean() {
	timeFormat := this.filename + "." + _DATEFORMAT
	nowTime, _ := time.Parse(_DATEFORMAT, time.Now().Format(_DATEFORMAT))
	lastTime := nowTime.Add(time.Duration(this.maxFileCount) * 24 * time.Hour * -1)
	files, _ := ioutil.ReadDir(this.dir)
	for _, f := range files {
		if f.IsDir() || f.Name() == this.filename {
			continue
		}
		logTime, err := time.Parse(timeFormat, f.Name())
		if err != nil {
			continue
		}
		if logTime.Before(lastTime) {
			fname := fmt.Sprint(this.dir, "/", f.Name())
			fmt.Println("clean expired log file", fname)
			os.Remove(fname)
		}
	}
}

func (this *fileBean) addsize(size int64) {
	atomic.AddInt64(&this.filesize, size)
}

func (this *fileBean) write(lvl, s string) int {
	this.mu.RLock()
	defer this.mu.RUnlock()
	_, file, line, _ := runtime.Caller(skip)
	short := file
	for i := len(file) - 1; i > 0; i-- {
		if file[i] == '/' {
			short = file[i+1:]
			break
		}
	}
	file = short
	content := fmt.Sprintln(time.Now().Format(_TIMEFORMAT), lvl, fmt.Sprintf("%s:%d", file, line), "-", s)
	this.lg.Output(skip+1, content)
	return len([]byte(content))
}

func (this *fileBean) isOverSize() bool {
	return this.filesize >= this.maxFileSize
}

func (this *fileBean) close() {
	this.logfile.Close()
}

//-----------------------------------------------------------------------------------------------

func mkdirlog(dir string) (e error) {
	_, er := os.Stat(dir)
	b := er == nil || os.IsExist(er)
	if !b {
		if err := os.MkdirAll(dir, 0666); err != nil {
			if os.IsPermission(err) {
				e = err
			}
		}
	}
	return
}

func fileSize(file string) int64 {
	f, e := os.Stat(file)
	if e != nil {
		fmt.Println(e.Error())
		return 0
	}
	return f.Size()
}

func isExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

func md5str(s string) string {
	m := md5.New()
	m.Write([]byte(s))
	return hex.EncodeToString(m.Sum(nil))
}

func catchError(id string) {
	if err := recover(); err != nil {
		fmt.Println("logger error", id, string(debug.Stack()))
	}
}
