package config

import (
	"../step"
	"../util"
	"bufio"
	"fmt"
	"time"
)

func GetOdroidStepList(w *bufio.Writer, c *step.StepConfig) *step.StepList {
	// Initialize steps
	stEnterUboot := &step.Step{
		Trigger: "ModeKey Check...",
		OnTrigger: func() {
			util.MustSendCmd(w, "\n", true)
		},
		Expect:  "Exynos4412 #",
		Msg:     "Enter u-boot console",
		Timeout: 10 * time.Second,
	}

	stStartEthernet := &step.Step{
		OnTrigger: func() {
			// Double \n -> uboot bug
			util.MustSendCmd(w, "usb start\n", true)
		},
		Expect:    "1 Ethernet Device(s) found",
		Msg:       "USB Ethernet start",
		SendProbe: true,
	}

	stSetEnv := &step.Step{
		OnTrigger: func() {
			util.MustSendCmd(w, "setenv ethact sms0;setenv ipaddr "+c.IpAddr, false)
			util.MustSendCmd(w, "setenv netmask 255.255.255.0;setenv serverip "+c.ServerAddr, false)
			util.MustSendCmd(w, "setenv ethaddr "+c.MacAddr+";printenv ipaddr", true)
		},
		Expect:    "ipaddr=" + c.IpAddr,
		Msg:       "Setup environment variables",
		SendProbe: true,
	}

	stLoadKernel := &step.Step{
		OnTrigger: func() {
			util.MustSendCmd(w, "tftp 0x40008000 odroid/zImage; bootm", true)
		},
		Expect:    "Welcome to Odroid",
		Msg:       "Get zImage via tftp",
		SendProbe: true,
		Timeout:   80 * time.Second,
	}

	stLoginRoot := &step.Step{
		OnTrigger: func() {
			util.MustSendCmd(w, "root", true)
		},
		Expect:    "",
		Msg:       "Login as root",
		SendProbe: true,
	}

	stLoadModules := &step.Step{
		OnTrigger: func() {
			util.MustSendCmd(w, "modprobe usbnet; modprobe smsc95xx", true)
		},
		Expect:    "registered new interface driver smsc95xx",
		Msg:       "Load modules",
		SendProbe: true,
	}

	stBringUpEth0 := &step.Step{
		OnTrigger: func() {
			util.MustSendCmd(w, "ifconfig eth0 up", true)
		},
		Expect:    "eth0: link up",
		Msg:       "Bring up eth0",
		SendProbe: true,
	}

	stAssignIP := &step.Step{
		OnTrigger: func() {
			util.MustSendCmd(w, "ip addr add "+c.IpAddr+"/24 dev eth0", false)
			util.MustSendCmd(w, "ip addr show eth0", true)
		},
		Expect:    "inet " + c.IpAddr + "/24",
		Msg:       "Assign IP address",
		SendProbe: true,
	}

	stWriteImage := &step.Step{
		OnTrigger: func() {
			params := "--progress=dot:kilo"
			command := "wget %s http://%s/%s -O -|xzcat -|dd of=%s bs=4M"

			command = fmt.Sprintf(command, params, c.ServerAddr,
				c.ImageFilename, c.Device)
			util.MustSendCmd(w, command, true)
		},
		Expect:    "100%",
		Msg:       "Image rootFS",
		SendProbe: true,
		Timeout:   60 * 3 * time.Second,
	}

	//stUpdateUboot := &step.Step{
	//OnTrigger: func() {
	//params := "--progress=dot:dots"
	//command := "wget %s http://%s/%s -O - | tar -xz && ./update.sh"
	//command = fmt.Sprintf(command, params, c.ServerAddr, c.UbootTar)
	//util.MustSendCmd(w, command, true)
	//},
	//Expect:    "U-boot image is fused successfully",
	//Msg:       "Flash U-boot",
	//SendProbe: true,
	//Timeout:   60 * 2 * time.Second,
	//}

	stFinalReboot := &step.Step{
		OnTrigger: func() {
			util.MustSendCmd(w, "reboot", true)
		},
		Expect:    "Mounting root file system",
		Msg:       "Final reboot",
		SendProbe: true,
	}

	stepList := step.NewList()
	stepList.Append(stEnterUboot)
	stepList.Append(stSetEnv)
	stepList.Append(stStartEthernet)
	stepList.Append(stLoadKernel)
	stepList.Append(stLoginRoot)
	stepList.Append(stLoadModules)
	stepList.Append(stBringUpEth0)
	stepList.Append(stAssignIP)
	stepList.Append(stWriteImage)
	//stepList.Append(stUpdateUboot)
	stepList.Append(stFinalReboot)

	return stepList
}
