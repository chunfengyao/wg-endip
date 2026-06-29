# WireGuard Endpoint Scanner (wg-endip)

A high-performance WireGuard endpoint scanner for testing connectivity and RTT.

## Features
- **Concurrent Scanning:** Uses multiple goroutines to test targets in parallel.
- **Protocol Simulation:** Constructs valid WireGuard handshake packets (Noise_IKpsk2_25519_ChaChaPoly_BLAKE2s).
- **Flexible Targets:** Supports individual IPs or CIDR blocks.
- **Detailed Reporting:** Outputs real-time scan logs and generates a CSV summary.
- **Cross-Platform:** Supports cross-compilation for various router architectures (ARM, MIPS, etc.).

## Usage
```bash
# Basic usage
./wg-endip -priStr <BASE64_PRIVATE_KEY>

# Advanced usage
./wg-endip -targets 1.1.1.0/24 -ports 51820 -concurrency 500 -repeat 5 -priStr <BASE64_PRIVATE_KEY>
```

## How It Works & Reporting
The tool performs a simulated WireGuard handshake against specified targets. 

### Data Processing
1. **Concurrency:** Scans run in parallel based on the `-concurrency` setting.
2. **RTT & Loss Calculation:** For each target, it sends `-repeat` number of packets. Loss and RTT (min/max/avg) are calculated from these attempts.
3. **Sorting:** The final report (both CSV and console output) is sorted by:
   - **Loss Percentage (Ascending):** Lower loss first.
   - **Average Latency (Ascending):** For equal loss, faster response time first.

### Reporting
The CSV output (`-output`) contains:
- `ip:port`: Target endpoint (IPv6 addresses are wrapped in `[]`).
- `loss_percent`: Percentage of failed handshakes.
- `min_rtt_ms`: Minimum Round Trip Time.
- `max_rtt_ms`: Maximum Round Trip Time.
- `avg_rtt_ms`: Average Round Trip Time.

## Command Line Arguments

| Parameter | Default | Description |
| :--- | :--- | :--- |
| `-help` | - | Show this help message |
| `-concurrency` | 200 | Number of concurrent targets to test simultaneously. Increase for faster scanning of large ranges. |
| `-repeat` | 20 | Number of handshake packets to send to each target to calculate accurate loss and RTT. |
| `-targets` | 162.159.193.0/24,2606:4700:100::/48 | Comma-separated list of IP addresses or CIDR ranges. |
| `-ports` | 2408,500,1701,4500 | Comma-separated list of UDP ports to test for each IP. |
| `-timeout` | 3s | Time to wait for a response from each packet before considering it lost. |
| `-max-tests` | 5000 | Maximum number of IP:Port combinations to test. Prevents accidental massive scans. |
| `-output` | report.csv | Path to save the final results in CSV format. |
| `-priStr` | - | **Required:** Base64 encoded WireGuard client private key. |
| `-pubStr` | bmXOC+F1FxEMF9dyiK2H5/1SUtzH0JuVo51h2wPfgyo= | Base64 encoded WireGuard server public key. |
| `-reserved` | [0,0,0] | 3-byte reserved field in array format (e.g., [0,0,0]) to include in the handshake packet. |

## Building
Run `./build.sh` to generate binaries.
- `./build.sh --all` : Build all supported architectures.
- `./build.sh <arch>` : Build for a specific architecture (e.g., `./build.sh linux-amd64`).
