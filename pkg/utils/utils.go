package utils

import (
	"encoding/csv"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/chunfengyao/wg-endip/pkg/types"
)

func formatTarget(t types.Target) string {
	ip := net.ParseIP(t.IP)
	if ip != nil && ip.To4() == nil {
		return fmt.Sprintf("[%s]:%d", t.IP, t.Port)
	}
	return fmt.Sprintf("%s:%d", t.IP, t.Port)
}

func ParseTargets(targetsStr, portsStr string, maxTests int) []types.Target {
	var targets []types.Target
	ports := []int{}
	for _, p := range strings.Split(portsStr, ",") {
		if port, err := strconv.Atoi(p); err == nil {
			ports = append(ports, port)
		}
	}

	targetItems := strings.Split(targetsStr, ",")
	for _, item := range targetItems {
		if strings.Contains(item, "/") {
			_, ipNet, err := net.ParseCIDR(item)
			if err != nil {
				continue
			}
			for i := 0; i < 20; i++ {
				ip := GenerateRandomIP(ipNet)
				for _, p := range ports {
					targets = append(targets, types.Target{IP: ip.String(), Port: p})
				}
			}
		} else {
			for _, p := range ports {
				targets = append(targets, types.Target{IP: item, Port: p})
			}
		}
	}

	if len(targets) > maxTests {
		return targets[:maxTests]
	}
	return targets
}

func GenerateRandomIP(ipNet *net.IPNet) net.IP {
	ip := make(net.IP, len(ipNet.IP))
	copy(ip, ipNet.IP)

	mask := ipNet.Mask
	ones, bits := mask.Size()

	for i := ones; i < bits; i++ {
		if rand.Intn(2) == 1 {
			ip[i/8] |= 1 << uint(7-i%8)
		} else {
			ip[i/8] &= ^(1 << uint(7-i%8))
		}
	}
	return ip
}

func WriteCSV(path string, results []types.Result) {
	file, err := os.Create(path)
	if err != nil {
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"ip:port", "loss_percent", "min_rtt_ms", "max_rtt_ms", "avg_rtt_ms"})

	for _, r := range results {
		writer.Write([]string{
			formatTarget(r.Target),
			fmt.Sprintf("%.0f", r.Loss*100),
			fmt.Sprintf("%.2f", float64(r.Min.Microseconds())/1000.0),
			fmt.Sprintf("%.2f", float64(r.Max.Microseconds())/1000.0),
			fmt.Sprintf("%.2f", float64(r.Avg.Microseconds())/1000.0),
		})
	}
}
