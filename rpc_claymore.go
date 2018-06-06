package rpcclaymore

import (
	"fmt"
	"net/rpc/jsonrpc"
	"strconv"
	"strings"
)

const (
	methodGetInfo      = "miner_getstat1"
	methodRestartMiner = "miner_restart"
	methodReboot       = "miner_reboot"
)

var args = struct {
	id      string
	jsonrpc string
	psw     string
}{"0", "2.0", ""}

// Crypto Information about this concrete crypto-currency
type Crypto struct {
	HashRate       int
	Shares         int
	RejectedShares int
	InvalidShares  int
}

func (c Crypto) String() (s string) {
	s += fmt.Sprintf("HashRate: %d\n", c.HashRate)
	s += fmt.Sprintf("Shares: %d\n", c.Shares)
	s += fmt.Sprintf("RejectedShares: %d\n", c.RejectedShares)
	s += fmt.Sprintf("InvalidShares: %d\n", c.InvalidShares)
	return s
}

// PoolInfo Information about the miner's connected pool
type PoolInfo struct {
	Address  string
	Switches int
}

func (p PoolInfo) String() (s string) {
	s += fmt.Sprintf("Address: %s\n", p.Address)
	s += fmt.Sprintf("Switches: %d\n", p.Switches)
	return s
}

// GPU Information about each concrete GPU
type GPU struct {
	HashRate    int
	AltHashRate int
	Temperature int
	FanSpeed    int
}

func (gpu GPU) String() (s string) {
	s += fmt.Sprintf("Hash Rate: %d\n", gpu.HashRate)
	s += fmt.Sprintf("Alt Hash Rate: %d\n", gpu.AltHashRate)
	s += fmt.Sprintf("Temperature: %d\n", gpu.Temperature)
	s += fmt.Sprintf("Fan Speed: %d\n", gpu.FanSpeed)
	return s
}

// MinerInfo Information about the miner
type MinerInfo struct {
	Version    string
	UpTime     int
	MainCrypto Crypto
	AltCrypto  Crypto
	MainPool   PoolInfo
	AltPool    PoolInfo
	GPUS       []GPU
}

func (m MinerInfo) String() string {
	var s string
	s += fmt.Sprintf("Version: %s\n", m.Version)
	s += fmt.Sprintf("Up Time: %d\n", m.UpTime)
	s += "\n"
	s += fmt.Sprintf("Main Crypto\n%s\n", m.MainCrypto)
	s += fmt.Sprintf("Alt Crypto\n%s\n", m.AltCrypto)
	s += fmt.Sprintf("Main Pool\n%s\n", m.MainPool)
	s += fmt.Sprintf("Alt Pool\n%s\n", m.AltPool)
	for i, gpu := range m.GPUS {
		s += fmt.Sprintf("GPU %d\n%s\n", i, gpu)
	}
	return s
}

// Miner creates an instance to get info of a miner
type Miner struct {
	Address  string
	Password string
}

func (m Miner) String() (s string) {
	return fmt.Sprintf("Miner {Address: %s}\n", m.Address)
}

// Restart Stop and start the miner
func (m Miner) Restart() error {
	client, err := jsonrpc.Dial("tcp", m.Address)
	if err != nil {
		return err
	}
	args.psw = m.Password
	return client.Call(methodRestartMiner, args, nil)
}

// Reboot Turn off and on again the computer
func (m Miner) Reboot() error {
	client, err := jsonrpc.Dial("tcp", m.Address)
	if err != nil {
		return err
	}
	args.psw = m.Password
	return client.Call(methodReboot, args, nil)
}

// GetInfo return the status of the miner or throw and error if it is not reachable
func (m Miner) GetInfo() (MinerInfo, error) {
	var mi MinerInfo
	var reply []string
	client, err := jsonrpc.Dial("tcp", m.Address)
	if err != nil {
		return mi, err
	}
	args.psw = m.Password
	fmt.Println(args)
	err = client.Call(methodGetInfo, args, &reply)
	if err != nil {
		return mi, err
	}
	return parseResponse(reply), nil
}

func parseResponse(info []string) MinerInfo {
	var mi MinerInfo
	var group []string

	mi.Version = strings.Replace(info[0], " - ETH", "", 1)
	mi.UpTime = toInt(info[1])

	group = splitGroup(info[2])
	mi.MainCrypto.HashRate = toInt(group[0])
	mi.MainCrypto.Shares = toInt(group[1])
	mi.MainCrypto.RejectedShares = toInt(group[2])

	group = splitGroup(info[4])
	mi.AltCrypto.HashRate = toInt(group[0])
	mi.AltCrypto.Shares = toInt(group[1])
	mi.AltCrypto.RejectedShares = toInt(group[2])

	group = splitGroup(info[7])
	mi.MainPool.Address = group[0]
	if len(group) > 1 {
		mi.AltPool.Address = group[1]
	}

	group = splitGroup(info[8])
	mi.MainCrypto.InvalidShares = toInt(group[0])
	mi.MainPool.Switches = toInt(group[1])
	mi.AltCrypto.InvalidShares = toInt(group[2])
	mi.AltPool.Switches = toInt(group[3])

	for _, hashrate := range splitGroup(info[3]) {
		mi.GPUS = append(mi.GPUS, GPU{HashRate: toInt(hashrate)})
	}

	for i, val := range splitGroup(info[6]) {
		if i%2 == 0 {
			mi.GPUS[i/2].Temperature = toInt(val)
		} else {
			mi.GPUS[(i-1)/2].FanSpeed = toInt(val)
		}
	}

	if mi.AltPool.Address != "" {
		for i, val := range splitGroup(info[5]) {
			hashrate, err := strconv.Atoi(val)
			if err == nil {
				mi.GPUS[i].AltHashRate = hashrate
			}
		}
	}

	return mi
}

func splitGroup(s string) []string {
	return strings.Split(s, ";")
}

func toInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}