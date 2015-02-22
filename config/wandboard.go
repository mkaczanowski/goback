package config

import (
	"../step"
	"../util"
	"bufio"
	"fmt"
	"time"
)

func GetWandboardStepList(w *bufio.Writer, c *step.StepConfig) *step.StepList {
	// Initialize steps
	stEnterUboot := &step.Step{
		Trigger: "Board: Wandboard",
		OnTrigger: func() {
			util.MustSendCmd(w, "\n", true)
		},
		Expect:  "=>",
		Msg:     "Enter u-boot console",
		Timeout: 10 * time.Second,
	}

	stSetEnv := &step.Step{
		OnTrigger: func() {
			util.MustSendCmd(w, "setenv ipaddr "+c.IpAddr, false)
			util.MustSendCmd(w, "setenv netmask 255.255.255.0;setenv serverip "+c.ServerAddr, false)
			util.MustSendCmd(w, "setenv ethaddr "+c.MacAddr+";printenv ipaddr", true)
		},
		Expect:    "ipaddr=" + c.IpAddr,
		Msg:       "Setup environment variables",
		SendProbe: true,
	}

	stLoadKernel := &step.Step{
		OnTrigger: func() {
			util.MustSendCmd(w, "tftp 0x12000000 wand/uImage;bootm 0x12000000", true)
		},
		Expect:    "Welcome to Buildroot",
		Msg:       "Get zImage via tftp",
		SendProbe: true,
		Timeout:   80 * time.Second,
	}

	stLoginRoot := &step.Step{
		OnTrigger: func() {
			util.MustSendCmd(w, "root", true)
		},
		Expect:    "#",
		Msg:       "Login as root",
		SendProbe: true,
	}

	stBringUpEth0 := &step.Step{
		OnTrigger: func() {
			util.MustSendCmd(w, "ifconfig eth0 up", true)
		},
		Expect:    "Link is Up",
		Msg:       "Bring up eth0",
		SendProbe: true,
	}

	stAssignIP := &step.Step{
		OnTrigger: func() {
			util.MustSendCmd(w, "echo performance > /sys/devices/system/cpu/cpu0/cpufreq/scaling_governor", false)
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
		Expect:    "records out",
		Msg:       "Image rootFS",
		SendProbe: true,
		Timeout:   60 * 15 * time.Second,
	}

	stFinalReboot := &step.Step{
		OnTrigger: func() {
			util.MustSendCmd(w, "reboot\n", false)
			util.MustSendCmd(w, "reboot", true)
		},
		Expect:    "ubuntu-armhf ttymxc0",
		Msg:       "Final reboot",
		SendProbe: true,
	}

	stepList := step.NewList()
	stepList.Append(stEnterUboot)
	stepList.Append(stSetEnv)
	stepList.Append(stLoadKernel)
	stepList.Append(stLoginRoot)
	stepList.Append(stBringUpEth0)
	stepList.Append(stAssignIP)
	stepList.Append(stWriteImage)
	stepList.Append(stFinalReboot)

	return stepList
}
