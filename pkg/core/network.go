package core

import (
	"fmt"
	"net"
	"time"

	"github.com/mdlayher/wifi"
)

func GetCurrentSSID() (string, error) {
	c, err := wifi.New()
	if err != nil {
		return "", fmt.Errorf("failed to open wifi handle: %w", err)
	}
	defer c.Close()

	interfaces, err := c.Interfaces()
	if err != nil {
		return "", fmt.Errorf("failed to get interfaces: %w", err)
	}

	for _, iface := range interfaces {
		bss, err := c.BSS(iface)
		if err != nil {
			continue
		}

		if bss != nil {
			return bss.SSID, nil
		}
	}

	return "", fmt.Errorf("no active Wi-Fi connection found")
}

func CheckConnectivity() bool {
	for i := 0; i < 3; i++ {
		if checkOnce() {
			return true
		}
		time.Sleep(2 * time.Second)
	}
	return false
}

func checkOnce() bool {
	conn, err := net.DialTimeout("tcp", "8.8.8.8:53", 2*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
