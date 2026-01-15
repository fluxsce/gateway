package net

import (
	"net"
	"os"
	"strings"
)

// GetAllMACAddresses 获取所有可用网卡的MAC地址
// 返回所有非回环、有MAC地址的网卡MAC，以逗号分隔
func GetAllMACAddresses() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return ""
	}
	var macs []string
	for _, iface := range interfaces {
		// 只跳过回环和无MAC的接口
		if iface.Flags&net.FlagLoopback != 0 || len(iface.HardwareAddr) == 0 {
			continue
		}
		macs = append(macs, iface.HardwareAddr.String())
	}
	return strings.Join(macs, ",")
}

// GetFirstIPv4Address 获取第一个非回环的IPv4地址
func GetFirstIPv4Address() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "127.0.0.1"
}

// GetFirstIPv6Address 获取第一个非回环的IPv6地址
func GetFirstIPv6Address() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "::1"
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() == nil && ipnet.IP.To16() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "::1"
}

// GetAllIPv4Addresses 获取所有非回环的IPv4地址
func GetAllIPv4Addresses() []string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil
	}
	var ips []string
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ips = append(ips, ipnet.IP.String())
			}
		}
	}
	return ips
}

// GetHostname 获取主机名
func GetHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return ""
	}
	return hostname
}
