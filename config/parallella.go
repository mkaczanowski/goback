package config

import (
	"../step"
	"../util"
	"bufio"
	"fmt"
	"time"
)

func GetParallellaStepList(w *bufio.Writer, c *step.StepConfig) *step.StepList {
	// Initialize steps
	stEnterUboot := &step.Step{
		Trigger: "zynq_gem",
		OnTrigger: func() {
			util.MustSendCmd(w, "\n", true)
		},
		Expect:  "zynq-uboot>",
		Msg:     "Enter u-boot console",
		Timeout: 10 * time.Second,
	}

	stSetEnv := &step.Step{
		OnTrigger: func() {
			util.MustSendCmd(w, "setenv gatewayip 192.168.4.1;setenv ipaddr "+c.IpAddr, false)
			util.MustSendCmd(w, "setenv netmask 255.255.255.0;setenv serverip "+c.ServerAddr, false)
			util.MustSendCmd(w, "setenv ethaddr "+c.MacAddr+";printenv ipaddr", true)
		},
		Expect:    "ipaddr=" + c.IpAddr,
		Msg:       "Setup environment variables",
		SendProbe: true,
	}

	stLoadFpga := &step.Step{
		OnTrigger: func() {
			util.MustSendCmd(w, "tftp 0x4000000 parallella/parallella.bit.bin;fpga load 0 0x4000000 0x3dbafc;tftp 0x3000000 parallella/uImage;tftp 0x2A00000 parallella/devicetree.dtb;tftp 0x1100000 parallella/initrd;bootm 0x3000000 0x1100000 0x2A00000", true)
		},
		Expect:    "Welcome to Parallella",
		Msg:       "Load FPGA",
		SendProbe: true,
		Timeout:   20 * time.Second,
	}

	stLoginRoot := &step.Step{
		OnTrigger: func() {
			util.MustSendCmd(w, "root", true)
		},
		Expect:    "#",
		Msg:       "Login as root",
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

	stFinalReboot := &step.Step{
		OnTrigger: func() {
			util.MustSendCmd(w, "reboot", true)
		},
		Expect:    "root@linaro-nano",
		Msg:       "Final reboot",
		SendProbe: true,
	}

	stepList := step.NewList()
	stepList.Append(stEnterUboot)
	stepList.Append(stSetEnv)
	stepList.Append(stLoadFpga)
	stepList.Append(stLoginRoot)
	stepList.Append(stWriteImage)
	stepList.Append(stFinalReboot)

	return stepList
}
