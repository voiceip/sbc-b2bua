package main

import (
	crand "crypto/rand"
	"flag"
	mrand "math/rand"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/voiceip/sbc-b2bua/sbc"
	"sippy"
	"sippy/conf"
	"sippy/log"
	"sippy/net"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	buf := make([]byte, 8)
	crand.Read(buf)
	var salt int64
	for _, c := range buf {
		salt = (salt << 8) | int64(c)
	}
	mrand.Seed(salt)

	var laddr, nh_addr, logfile string
	var lport int
	var err error

	flag.StringVar(&laddr, "l", "", "Local addr")
	flag.IntVar(&lport, "p", -1, "Local port")
	flag.StringVar(&nh_addr, "n", "", "Next hop address")
	flag.StringVar(&logfile, "L", "/var/log/sip.log", "Log file")
	flag.Parse()

	error_logger := sippy_log.NewErrorLogger()
	sip_logger, err := sippy_log.NewSipLogger("b2bua", logfile)
	if err != nil {
		error_logger.Error(err)
		return
	}
	config := &sbc.Config{
		Config:  sippy_conf.NewConfig(error_logger, sip_logger),
		NH_addr: sippy_net.NewHostPort("192.168.0.102", "5060"), // next hop address
	}
	//config.SetIPV6Enabled(false)
	if nh_addr != "" {
		var parts []string
		var addr string

		if strings.HasPrefix(nh_addr, "[") {
			parts = strings.SplitN(nh_addr, "]", 2)
			addr = parts[0] + "]"
			if len(parts) == 2 {
				parts = strings.SplitN(parts[1], ":", 2)
			}
		} else {
			parts = strings.SplitN(nh_addr, ":", 2)
			addr = parts[0]
		}
		port := "5060"
		if len(parts) == 2 {
			port = parts[1]
		}
		config.NH_addr = sippy_net.NewHostPort(addr, port)
	}
	config.SetMyUAName("Sippy B2BUA (Simple)")
	config.SetAllowFormats([]int{0, 8, 18, 100, 101})
	if laddr != "" {
		config.SetMyAddress(sippy_net.NewMyAddress(laddr))
	}
	config.SetSipAddress(config.GetMyAddress())
	if lport > 0 {
		config.SetMyPort(sippy_net.NewMyPort(strconv.Itoa(lport)))
	}
	config.SetSipPort(config.GetMyPort())
	cmap := sbc.NewCallManager(config, error_logger)
	sipTransactionManager, err := sippy.NewSipTransactionManager(config, cmap)
	if err != nil {
		error_logger.Error(err)
		return
	}
	cmap.Sip_TM = sipTransactionManager
	cmap.Proxy = sippy.NewStatefulProxy(sipTransactionManager, config.NH_addr, config)
	go sipTransactionManager.Run()

	signal_chan := make(chan os.Signal, 1)
	signal.Notify(signal_chan, syscall.SIGTERM, syscall.SIGINT)
	signal.Ignore(syscall.SIGHUP, syscall.SIGPIPE, syscall.SIGUSR1, syscall.SIGUSR2)
	select {
	case <-signal_chan:
		cmap.Shutdown()
		sipTransactionManager.Shutdown()
		time.Sleep(time.Second)
		break
	}
}
