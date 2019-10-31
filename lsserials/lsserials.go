/*
Tool for listing serial ports.
*/
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/hjkoskel/listserialports"
)

func main() {
	pPort := flag.String("f", "", "is port free")
	pCertain := flag.Bool("c", false, "port is free only if it certain")
	pMonitor := flag.Bool("m", false, "monitor serial port status by polling")
	pPollInterval := flag.Int("p", 1500, "poll delay in ms. If monitor is selected")
	flag.Parse()

	if 0 < len(*pPort) { //Check only one port
		pidlist, certain, errProbe := listserialports.FileIsInUse(*pPort)
		if errProbe != nil {
			fmt.Printf("%v\n", errProbe)
			os.Exit(-2)
		}
		if *pCertain && !certain {
			fmt.Printf("Uncertain is free\n")
			os.Exit(-1)
		}
		if len(pidlist) == 0 {
			fmt.Printf("%v is free\n", *pPort)
			os.Exit(0)
		}
		if len(pidlist) == 1 {
			fmt.Printf("File %v is used by %v\n", *pPort, pidlist[0])
			os.Exit(-1)
		}
		fmt.Printf("File %v is used by %v\n", *pPort, pidlist)
		os.Exit(-1)
	}

	entries, errProbe := listserialports.ProbeSerialports()
	if errProbe != nil {
		fmt.Printf("%v", errProbe)
		os.Exit(-1)
	}
	for i, entry := range entries {
		fmt.Printf("%v: %v", i, entry.ToPrintoutFormat())
	}

	if !*pMonitor {
		os.Exit(0)
	}
	//Monitor by polling

	pollDelay := time.Duration(*pPollInterval) * time.Millisecond
	for {
		time.Sleep(pollDelay)
		nextEntries, errNextProbe := listserialports.ProbeSerialports()
		if errNextProbe != nil {
			fmt.Printf("%v", errProbe)
			os.Exit(-1)
		}

		added := listserialports.NewEntries(entries, nextEntries)
		for _, e := range added {
			fmt.Printf("\nADDED:\n%v\n", e.ToPrintoutFormat())
		}

		removed := listserialports.NewEntries(nextEntries, entries) //reverse time and adding becomes remove
		for _, e := range removed {
			fmt.Printf("\n%v removed\n", e.DeviceFile)
		}

		//Status upgrades? taken by someone?
		updated := listserialports.Updates(entries, nextEntries)
		for _, e := range updated {
			fmt.Printf("\nUPDATED:\n%v", e.ToPrintoutFormat())
		}

		entries = nextEntries
	}

}
