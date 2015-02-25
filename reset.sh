#!/bin/bash

declare -A ports
ports=( ["atx"]=60 ["odroid0"]=46 ["odroid1"]=47
	["odroid2"]=26 ["odroid3"]=44 ["parallella0"]=45
	["parallella1"]=115 ["wandboard0"]=49 ["wandboard1"]=48)

function exportPin {
	pin=$1	
	path="/sys/class/gpio/gpio$pin/value"
	if [ ! -f $path ]; then
		echo $pin > /sys/class/gpio/export	
		echo "out" > /sys/class/gpio/gpio$pin/direction
	fi
}

function parseMach {
	mach=$1
	NO=${mach//[^0-9]/}
	MACH=${mach/[0-9]/}
}

function checkATX {
	pin=${ports["atx"]}
	exportPin $pin
	
	pinVal=$(cat /sys/class/gpio/gpio$pin/value)
	if [ $pinVal -eq 1 ]; then 
		echo 0 > /sys/class/gpio/gpio$pin/value
		sleep 0.6
	fi
}

function powerOn {
	if [ "$1" != "atx" ]; then 
		checkATX
	fi	

	pin=${ports[$1]}
	exportPin $pin

	pinVal=$(cat /sys/class/gpio/gpio$pin/value)
	if [ $pinVal -eq 0 ]; then 
		echo 1 > /sys/class/gpio/gpio$pin/value
	fi
}

function powerOff {
	turnoffState=0
	isOn=1
	if [ "$1" == "atx" ]; then
		turnoffState=1	
		isOn=0
	fi

	pin=${ports[$1]}
	exportPin $pin

	pinVal=$(cat /sys/class/gpio/gpio$pin/value)
	if [ $pinVal -eq $isOn ]; then 
		echo $turnoffState > /sys/class/gpio/gpio$pin/value
	fi
}

function switchPower {
	if [ "$1" != "atx" ]; then
		checkATX
	fi

        pin=${ports[$1]}
        exportPin $pin

	pinVal=$(cat /sys/class/gpio/gpio$pin/value)
	if [ $pinVal -eq 0 ]; then 
		echo 1 > /sys/class/gpio/gpio$pin/value
	else
		echo 0 > /sys/class/gpio/gpio$pin/value
	fi
}

function iterate {
	funcName=$1
	target=$2
	i=0	

	parseMach $target	
	if [ "$NO" == "" -a "$target" != "atx" ]; then
		while true; do
			key="$MACH$i"
			if [[ ${ports[$key]} ]]; then
				echo "$funcName: $key";
				$funcName $key
				i=$((i+1))
				sleep 0.6
			else
				break
			fi
		done	
	else
		echo "$funcName: $target"
		$funcName $target
		sleep 0.6
	fi
}

ACTION=$1
IFS=","

for v in $2; do
  if [ "$ACTION" == "on" ]; then
	iterate "powerOn" $v
  elif [ "$ACTION" == "off" ]; then
	iterate "powerOff" $v
  elif [ "$ACTION" == "switch" ]; then
	iterate "switchPower" $v
  else
	echo "Unrecognized action"
  fi
done

