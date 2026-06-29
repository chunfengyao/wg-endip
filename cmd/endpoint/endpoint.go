package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"math/rand"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/chunfengyao/wg-endip/pkg/scanner"
	"github.com/chunfengyao/wg-endip/pkg/types"
	"github.com/chunfengyao/wg-endip/pkg/utils"
	"github.com/chunfengyao/wg-endip/pkg/wg"
)

func formatTarget(t types.Target) string {
	ip := net.ParseIP(t.IP)
	if ip != nil && ip.To4() == nil {
		return fmt.Sprintf("[%s]:%d", t.IP, t.Port)
	}
	return fmt.Sprintf("%s:%d", t.IP, t.Port)
}

func main() {
	concurrency := flag.Int("concurrency", 200, "Number of concurrent target tests")
	repeat := flag.Int("repeat", 20, "Number of packets to send per target")
	targetsArg := flag.String("targets", "162.159.193.0/24,2606:4700:100::/48", "Comma-separated IP lists or CIDRs")
	portsArg := flag.String("ports", "2408,500,1701,4500", "Comma-separated ports")
	timeout := flag.Duration("timeout", 3*time.Second, "Timeout for each packet")
	maxTests := flag.Int("max-tests", 5000, "Max IP:Port combinations to test")
	outputFile := flag.String("output", "report.csv", "Path to save CSV report")
	priStr := flag.String("priStr", "", "Base64 encoded client private key")
	pubStr := flag.String("pubStr", "bmXOC+F1FxEMF9dyiK2H5/1SUtzH0JuVo51h2wPfgyo=", "Base64 encoded server public key")
	reservedStr := flag.String("reserved", "[0,0,0]", "Comma-separated 3 bytes in array format, e.g., [0,0,0]")

	flag.Parse()

	if *priStr == "" {
		fmt.Println("Error: -priStr parameter is required.")
		flag.Usage()
		os.Exit(1)
	}

	clientPrivateBytes, err := base64.StdEncoding.DecodeString(*priStr)
	if err != nil {
		fmt.Printf("Error decoding -priStr (Base64): %v\n", err)
		os.Exit(1)
	}
	responderStaticPublic, err := base64.StdEncoding.DecodeString(*pubStr)
	if err != nil {
		fmt.Printf("Error decoding -pubStr (Base64): %v\n", err)
		os.Exit(1)
	}

	reservedBytes := make([]byte, 3)
	cleaned := strings.Trim(strings.Trim(*reservedStr, "["), "]")
	reservedParts := strings.Split(cleaned, ",")
	if len(reservedParts) == 3 {
		for i := 0; i < 3; i++ {
			b, err := strconv.Atoi(strings.TrimSpace(reservedParts[i]))
			if err == nil {
				reservedBytes[i] = byte(b)
			}
		}
	}

	handshakePacket := wg.GenerateHandshakePacket(clientPrivateBytes, responderStaticPublic, reservedBytes)

	rand.Seed(time.Now().UnixNano())
	targets := utils.ParseTargets(*targetsArg, *portsArg, *maxTests)

	fmt.Println("Starting tests...")

	sem := make(chan struct{}, *concurrency)
	results := make([]types.Result, 0, len(targets))
	resultsChan := make(chan types.Result, len(targets))

	for _, t := range targets {
		go func(target types.Target) {
			sem <- struct{}{}
			resultsChan <- scanner.RunTest(target, handshakePacket, *repeat, *timeout)
			<-sem
		}(t)
	}

	for i := 0; i < len(targets); i++ {
		results = append(results, <-resultsChan)
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].Loss != results[j].Loss {
			return results[i].Loss < results[j].Loss
		}
		return results[i].Avg < results[j].Avg
	})

	utils.WriteCSV(*outputFile, results)

	fmt.Println("\nTop 20 results:")
	fmt.Printf("%-40s %-10s %-10s %-10s %-10s\n", "ip:port", "loss(%)", "min(ms)", "max(ms)", "avg(ms)")
	limit := 20
	if len(results) < limit {
		limit = len(results)
	}
	for i := 0; i < limit; i++ {
		r := results[i]
		fmt.Printf("%-40s %-10.0f %-10.2f %-10.2f %-10.2f\n",
			formatTarget(r.Target),
			r.Loss*100,
			float64(r.Min.Microseconds())/1000.0,
			float64(r.Max.Microseconds())/1000.0,
			float64(r.Avg.Microseconds())/1000.0,
		)
	}

	fmt.Println("\nTests completed. Report saved to", *outputFile)
}
