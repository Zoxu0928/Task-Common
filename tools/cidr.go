package tools

import (
	"bytes"
	"fmt"
	"net"
	"strconv"
	"strings"
)

func IpInCidr(ip, cidr string) (bool, string) {
	ipParsed := net.ParseIP(ip)

	firstIp, lastIp := GetCidrIpRange(cidr)
	firstIpParsed := net.ParseIP(firstIp)
	lastIpParsed := net.ParseIP(lastIp)

	if bytes.Compare(ipParsed, lastIpParsed) > 0 || bytes.Compare(ipParsed, firstIpParsed) < 0 {
		errmsg := fmt.Sprintf("ip %s is not in cidr %s", ip, cidr)
		return false, errmsg
	}
	return true, ""
}

func GetCidrIpRange(cidr string) (string, string) {
	ip := strings.Split(cidr, "/")[0]
	ipSegs := strings.Split(ip, ".")
	maskLen, _ := strconv.Atoi(strings.Split(cidr, "/")[1])
	seg1MinIp, seg1MaxIp := GetIpSegRange(ipSegs, maskLen, 0, 8)
	seg2MinIp, seg2MaxIp := GetIpSegRange(ipSegs, maskLen, 1, 16)
	seg3MinIp, seg3MaxIp := GetIpSegRange(ipSegs, maskLen, 2, 24)
	seg4MinIp, seg4MaxIp := GetIpSegRange(ipSegs, maskLen, 3, 32)
	return strconv.Itoa(seg1MinIp) + "." + strconv.Itoa(seg2MinIp) + "." + strconv.Itoa(seg3MinIp) + "." + strconv.Itoa(seg4MinIp+1),
		strconv.Itoa(seg1MaxIp) + "." + strconv.Itoa(seg2MaxIp) + "." + strconv.Itoa(seg3MaxIp) + "." + strconv.Itoa(seg4MaxIp-1)
}

func GetIpSegRange(ipSegs []string, maskLen int, index int, len int) (int, int) {
	if maskLen > len {
		segIp, _ := strconv.Atoi(ipSegs[index])
		return segIp, segIp
	}
	ipSeg, _ := strconv.Atoi(ipSegs[index])
	getIpSegRangeBase := func(userSegIp, offset uint8) (int, int) {
		var ipSegMax uint8 = 255
		netSegIp := ipSegMax << offset
		segMinIp := netSegIp & userSegIp
		segMaxIp := userSegIp&(255<<offset) | ^(255 << offset)
		return int(segMinIp), int(segMaxIp)
	}
	return getIpSegRangeBase(uint8(ipSeg), uint8(len-maskLen))
}

func CheckCidrOverlap(Cidr1, Cidr2 string) bool {
	_, net1, _ := net.ParseCIDR(Cidr1)
	_, net2, _ := net.ParseCIDR(Cidr2)

	return net2.Contains(net1.IP) || net1.Contains(net2.IP)
}

func CheckCidrInclusion(Cidr1, Cidr2 string) bool {
	_, net1, _ := net.ParseCIDR(Cidr1)
	_, net2, _ := net.ParseCIDR(Cidr2)

	if net1.Contains(net2.IP) == true {
		maskLen1, _ := net1.Mask.Size()
		maskLen2, _ := net2.Mask.Size()
		if maskLen1 <= maskLen2 {
			return true
		}
	}

	return false
}
