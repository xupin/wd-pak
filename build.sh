#!/bin/sh

pf=${1}

workdir=$(cd $(dirname $0); pwd)
out=$workdir/build/pak-tool
arch=${2-amd64}

if [ "$pf" = "windows" ];then
    os=windows out="$out-win.exe"
elif [ "$pf" = "linux" ];then
    os=linux out="$out-linux"
else    
    os=darwin out="$out-osx"
fi

CGO_ENABLED=0 GOARCH=$arch GOOS=$os go build -ldflags "-w -s" -gcflags="all=-trimpath=${PWD}" -asmflags="all=-trimpath=${PWD}" -o $out $workdir/.