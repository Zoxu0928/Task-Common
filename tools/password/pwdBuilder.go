package password

import (
	"bytes"
	"math/rand"
	"regexp"
	"time"
)

const (
	// 特殊单词
	SpecWord = `(jd)|(JD)|(360)|(bug)|(BUG)|(com)|(COM)|(cloud)|(CLOUD)|(password)|(PASSWORD)`

	// 连续数字
	SerialNumbers = `(123)|(234)|(345)|(456)|(567)|(678)|(789)|(890)|(098)|(987)|(876)|(765)|(654)|(543)|(432)|(321)`

	// 连续字母小写
	SerialLowercase = `(abc)|(bcd)|(cde)|(def)|(efg)|(fgh)|(ghi)|(hij)|(ijk)|(jkl)|(klm)|(lmn)|(mno)|(nop)|(opq)|(pqr)|(qrs)|(rst)|(stu)|(tuv)|(uvw)|(vwx)|(wxy)|(xyz)|(zyx)|(yxw)|(xwv)|(wvu)|(vut)|(uts)|(tsr)|(srq)|(rqp)|(qpo)|(pon)|(onm)|(nml)|(mlk)|(lkj)|(kji)|(jih)|(ihg)|(hgf)|(gfe)|(fed)|(edc)|(dcb)|(cba)|(qwe)|(wer)|(ert)|(rty)|(tyu)|(yui)|(uio)|(iop)|(poi)|(oiu)|(iuy)|(uyt)|(ytr)|(tre)|(rew)|(ewq)|(asd)|(sdf)|(dfg)|(fgh)|(ghj)|(hjk)|(jkl)|(lkj)|(kjh)|(jhg)|(hgf)|(gfd)|(fds)|(dsa)|(zxc)|(xcv)|(cvb)|(vbn)|(bnm)|(mnb)|(nbv)|(bvc)|(vcx)|(cxz)|(qaz)|(wsx)|(edc)|(rfv)|(tgb)|(yhn)|(ujm)|(zaq)|(xsw)|(cde)|(vfr)|(bgt)|(nhy)|(mju)`

	// 连续字母大写
	SerialUppercase = `(ABC)|(BCD)|(CDE)|(DEF)|(EFG)|(FGH)|(GHI)|(HIJ)|(IJK)|(JKL)|(KLM)|(LMN)|(MNO)|(NOP)|(OPQ)|(PQR)|(QRS)|(RST)|(STU)|(TUV)|(UVW)|(VWX)|(WXY)|(XYZ)|(ZYX)|(YXW)|(XWV)|(WVU)|(VUT)|(UTS)|(TSR)|(SRQ)|(RQP)|(QPO)|(PON)|(ONM)|(NML)|(MLK)|(LKJ)|(KJI)|(JIH)|(IHG)|(HGF)|(GFE)|(FED)|(EDC)|(DCB)|(CBA)|(QWE)|(WER)|(ERT)|(RTY)|(TYU)|(YUI)|(UIO)|(IOP)|(POI)|(OIU)|(IUY)|(UYT)|(YTR)|(TRE)|(REW)|(EWQ)|(ASD)|(SDF)|(DFG)|(FGH)|(GHJ)|(HJK)|(JKL)|(LKJ)|(KJH)|(JHG)|(HGF)|(GFD)|(FDS)|(DSA)|(ZXC)|(XCV)|(CVB)|(VBN)|(BNM)|(MNB)|(NBV)|(BVC)|(VCX)|(CXZ)|(QAZ)|(WSX)|(EDC)|(RFV)|(TGB)|(YHN)|(UJM)|(ZAQ)|(XSW)|(CDE)|(VFR)|(BGT)|(NHY)|(MJU)`
)

var (
	normal = [...]string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "g", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z",
		"A", "B", "C", "D", "E", "F", "G", "H", "I", "G", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z",
		"1", "2", "3", "4", "5", "6", "7", "8", "9", "0"}

	special = [...]string{"(", ")", "`", "~", "!", "@", "#", "$", "%", "^", "&", "*", "_", "-", "+", "=", "|", "{", "}", "[", "]", ":", ";", "'", "<", ">", ".", "?", "/"}
)

func GetPwd() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	var pwd string
	for i := 0; i < 7; i++ {
		pwd = pwd + normal[r.Intn(len(normal))]
	}
	pwd = pwd + special[r.Intn(len(special))]
	return pwd
}

//未设置则为随机产生，并通过站内信和邮件下发
//必须包含大写字母、小写字母、数字及特殊字符中三类，且不能少于8字符不能超过30字符
//特殊字符如下() ` ~ ! @ # $ % ^ & *_-+= {}[]: " ;'<>,.? /
//不能出现的字符或完整单词，如下：jd、JD、360、bug、BUG、com、COM、cloud、CLOUD、password、PASSWORD
//不能出现连续数字，例：123、987
//不能出现连续或键位连续字母，例：abc、CBA、bcde、qaz、tfc、zaq、qwer
//密码中不能出现自己的用户名。
//如果传入Password参数，请务必使用HTTPS协议调用API以避免密码泄露
//密码和SSH秘钥不能同时指定（暂时不限制）
func CheckPasswd(pin, pwd string) string {

	password := []byte(pwd)

	// 长度
	length := len(pwd)
	if length < 8 || length > 30 {
		return "Invalid password. The length must be between 8 and 30."
	}

	// 不能出现单词
	pattern_word := regexp.MustCompile(SpecWord)
	if pattern_word.Match(password) {
		return "Invalid password. Cannot appear the following words. jd、JD、360、bug、BUG、com、COM、cloud、CLOUD、password、PASSWORD"
	}

	// 不能出现pin
	pattern_pin := regexp.MustCompile("(" + pin + ")")
	if pattern_pin.Match(password) {
		return "Invalid password. Cannot appear pin."
	}

	// 包含不支持的字符
	if !regexp.MustCompile(`^[` + "`" + `|()~!@#$%^&*\-_+={}\[\]:";'<>,.?/0-9a-zA-Z]+$`).Match(password) {
		return "Invalid password. Found unsupported characters"
	}

	// 必须数字、大小写字母、特殊字符中的三类
	include := 0
	if regexp.MustCompile(`[0-9]`).Match(password) {
		include = include + 1
	}
	if regexp.MustCompile(`[a-z]`).Match(password) {
		include = include + 1
	}
	if regexp.MustCompile(`[A-Z]`).Match(password) {
		include = include + 1
	}
	if regexp.MustCompile(`[` + "`" + `|()~!@#$%^&*\-_+={}\[\]:";'<>,.?/]`).Match(password) {
		include = include + 1
	}
	if include < 3 {
		return "Invalid password. Must contains at least three types of numbers, uppercase letters, lowercase letters, and special characters"
	}

	// 不能含有三个连贯数字
	if regexp.MustCompile(SerialNumbers).Match(password) {
		return "Invalid password. Can't contain three consecutive number"
	}

	// 不能含有三个连续字母
	if regexp.MustCompile(SerialLowercase).Match([]byte(pwd)) || regexp.MustCompile(SerialUppercase).Match([]byte(pwd)) {
		return "Invalid password. Can't contain three consecutive character"
	}

	return ""
}

// 获得连续字符串的正则规则
func get_consecutive_reg() string {

	reg := ""

	// 字母连续
	s := "abcdefghijklmnopqrstuvwxyz"
	reg = get_consecutive_str(s)

	// 第一排连续
	s = "qwertyuiop"
	reg = reg + "|" + get_consecutive_str(s)

	// 第二排连续
	s = "asdfghjkl"
	reg = reg + "|" + get_consecutive_str(s)

	// 第三排连续
	s = "zxcvbnm"
	reg = reg + "|" + get_consecutive_str(s)

	// 纵向连续
	reg = reg + "|(qaz)|(wsx)|(edc)|(rfv)|(tgb)|(yhn)|(ujm)"
	reg = reg + "|(zaq)|(xsw)|(cde)|(vfr)|(bgt)|(nhy)|(mju)"

	return reg
}

// 获得连续字符串的正则规则，正序，反序
func get_consecutive_str(str string) string {
	str_buf := new(bytes.Buffer)
	for i := 0; i < len(str)-2; i++ {
		if i > 0 {
			str_buf.WriteString("|")
		}
		str_buf.WriteString("(")
		str_buf.WriteString(str[i : i+3])
		str_buf.WriteString(")")
	}
	for i := len(str); i > 2; i-- {
		str_buf.WriteString("|(")
		str_buf.WriteString(str[i-1 : i])
		str_buf.WriteString(str[i-2 : i-1])
		str_buf.WriteString(str[i-3 : i-2])
		str_buf.WriteString(")")
	}
	return str_buf.String()
}
