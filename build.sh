#!/bin/sh

os=`uname`
echo "Detected OS: $os"

PWD=`pwd`

case $os in
	Linux | MINGW32_NT-* | MINGW64_NT-*)
		export GOPATH=$PWD
		export GOBIN=$PWD/bin
		;;
	"Darwin")
		export GOPATH=~/source/nats
		export GOBIN=~/source/nats/bin
		;;
	*)
		echo "ERROR: Unknown OS."
		exit
		;;
esac

if [ ! -d "src/github.com/bocajim/evaler" ]; then
	go get github.com/bocajim/evaler
fi

build_all() {
	if [ -d "pkg" ]; then
    	rm -Rf pkg/*
	fi
	go install src/nagios-checks.go

}

build_for_windows() {
	# Builds 64bit Windows binaries
	echo "Building 64bit Windows nats"

	export GOOS=windows
	export GOARCH=amd64
	export CGO_ENABLED=0
	build_all
}

build_for_linux() {
	echo "Building 64bit Linux nats"

	export GOOS=linux
	build_all
}

case $1 in
	"all")
		build_all
		;;
	"linux")
		build_for_linux
		;;
	"windows")
		build_for_windows
		;;
	*)
		build_all
		;;
esac

echo "Done"
