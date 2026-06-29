package scanner

import (
	"fmt"
	"net"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/chunfengyao/wg-endip/pkg/types"
)

func formatTarget(t types.Target) string {
	ip := net.ParseIP(t.IP)
	if ip != nil && ip.To4() == nil {
		return fmt.Sprintf("[%s]:%d", t.IP, t.Port)
	}
	return fmt.Sprintf("%s:%d", t.IP, t.Port)
}

func RunTest(t types.Target, handshakePacket []byte, repeat int, timeout time.Duration) types.Result {
	totalPackets := repeat
	numCPU := runtime.NumCPU()

	numStreams := numCPU
	if totalPackets < numCPU {
		numStreams = totalPackets
	}
	if numStreams == 0 {
		numStreams = 1
	}

	packetsPerStream := totalPackets / numStreams

	var wg sync.WaitGroup
	var mu sync.Mutex

	var successPackets int
	var rtts []time.Duration

	for i := 0; i < numStreams; i++ {
		wg.Add(1)
		p := packetsPerStream
		if i == numStreams-1 {
			p += totalPackets % numStreams
		}

		go func(packets int) {
			defer wg.Done()
			conn, err := net.DialTimeout("udp", net.JoinHostPort(t.IP, strconv.Itoa(t.Port)), timeout)
			if err != nil {
				return
			}
			defer conn.Close()

			for j := 0; j < packets; j++ {
				start := time.Now()
				conn.SetDeadline(start.Add(timeout))
				_, err := conn.Write(handshakePacket)
				if err != nil {
					continue
				}

				buf := make([]byte, 1024)
				n, err := conn.Read(buf)

				rtt := time.Since(start)

				mu.Lock()
				targetStr := formatTarget(t)
				if err == nil && n > 0 {
					successPackets++
					rtts = append(rtts, rtt)
					fmt.Printf("%s %s 收到UDP响应: %x, RTT: %.2fms\n", time.Now().Format("15:04:05.000"), targetStr, buf[:n], float64(rtt.Microseconds())/1000.0)
				} else {
					fmt.Printf("%s %s 超时或无响应\n", time.Now().Format("15:04:05.000"), targetStr)
				}
				mu.Unlock()
			}
		}(p)
	}
	wg.Wait()

	loss := float64(totalPackets-successPackets) / float64(totalPackets)

	if len(rtts) == 0 {
		return types.Result{Target: t, Loss: loss}
	}

	var min, max, sum time.Duration
	min = rtts[0]
	max = rtts[0]
	for _, r := range rtts {
		if r < min {
			min = r
		}
		if r > max {
			max = r
		}
		sum += r
	}

	return types.Result{
		Target: t,
		Loss:   loss,
		Min:    min,
		Max:    max,
		Avg:    sum / time.Duration(len(rtts)),
	}
}
