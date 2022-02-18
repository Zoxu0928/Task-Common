package tools

import (
	"net"
)

// 获得本地IP地址
func GetIntranetIp() ([]string, error) {
	local_ips := make([]string, 0)
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return local_ips, err
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && !ipnet.IP.IsLinkLocalUnicast() && !ipnet.IP.IsLinkLocalMulticast() {
			if ipnet.IP.To4() != nil {
				local_ips = append(local_ips, ipnet.IP.String())
			}
		}
	}
	return local_ips, nil
}
