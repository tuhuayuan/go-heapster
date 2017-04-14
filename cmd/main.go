package main

import (
	"flag"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"zonst/qipai/logagent/utils"

	_ "zonst/qipai/gamehealthysrv/daemons"
)

var (
	configDir = flag.String("configs", "./configs", "Directory of config files.")
	level     = flag.Int("v", 3, "Logger level 0(panic)~5(debug).")
	help      = flag.Bool("help", false, "Print this message.")
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	var (
		err      error
		hostname string
		ag       *utils.Agent
	)

	flag.Parse()

	utils.SetLoggerLevel(*level)
	if *help {
		flag.Usage()
		os.Exit(0)
	}

	hostname, err = os.Hostname()
	if err != nil {
		hostname = "gamehealthy"
	}

	// trace system signal.
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	go func() {
		<-signalChan
		ag.Stop()
	}()

	ag = utils.NewAgent()
	ag.ConfigDir = *configDir
	ag.Name = hostname

	if err = ag.Run(); err != nil {
		os.Exit(-1)
	}
}
