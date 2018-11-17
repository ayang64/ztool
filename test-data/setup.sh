#! /usr/local/bin/bash

mdconfig -l | xargs -n1 mdconfig -d -u

ndevices=${1:-2}

setup(){
	datafile=devices.$$
	devs=""
	for ((x=0; x<${ndevices};x++)); do
		vol=vol${dev}.$$.zfs
		truncate -s1G ${vol}
		echo "created ${vol}."
		d=$(mdconfig -f ${vol})
		echo "associated ${d} with ${vol}."
		devs="${d} ${devs}"
	done
	echo >${datafile} ${devs}
	echo "wrote ${datafile}"
	echo "created ${devs}"
}

teardown(){
	datafile=devices.${1:-"000"}
	cat ${datafile} | xargs -n1 mdconfig -d -u
}

setup
