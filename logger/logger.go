package logger

const (
	_ver string = "1.0.3"
)

type level int32
type unit int64
type _ROLLTYPE int //dailyrolling ,rollingfile

const _DATEFORMAT = "2006-01-02"
const _TIMEFORMAT = "2006-01-02 15:04:05.000"

var logLevel level = 1

const (
	_       = iota
	kb unit = 1 << (iota * 10)
	mb
	gb
	tb
)

const (
	_ALL level = iota
	_DEBUG
	_INFO
	_WARN
	_ERROR
	_FATAL
	_OFF
)

const (
	_DAILY _ROLLTYPE = iota
	_ROLLFILE
)

func setConsole(isConsole bool) {
	defaultlog.setConsole(isConsole)
}
func setLevel(_level level) {
	defaultlog.setLevel(_level)
}
func setRollingFile(fileDir, fileName string, maxNumber int32, maxSize int64, _unit unit) {
	defaultlog.setRollingFile(fileDir, fileName, maxNumber, maxSize, _unit)
}
func setRollingDaily(fileDir, fileName string, maxFileCount int32) {
	defaultlog.setRollingDaily(fileDir, fileName, maxFileCount)
}
func setLevelFile(level level, dir, fileName string) {
	defaultlog.setLevelFile(level, dir, fileName)
}
