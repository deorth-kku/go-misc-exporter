package hwmon

import (
	"bufio"
	"os"
	"strings"
)

const (
	venderAMD     = "AuthenticAMD"
	venderIntel   = "GenuineIntel"
	venderUnknown = "Unknown"
)

func getCPUVendor() string {
	file, err := os.Open("/proc/cpuinfo")
	if err != nil {
		return venderUnknown
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// 查找 vendor_id 这一行
		if strings.HasPrefix(line, "vendor_id") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return venderUnknown
}
