#!/bin/sh

pf=${1}

workdir=$(cd $(dirname $0); pwd)
out=$workdir/build/pak-tool
CGO_ENABLED=0
GOARCH=${2-amd64}
GOOS=darwin


if [ "$pf" = "windows" ];then
    GOOS=windows out="$out-win.exe"
elif [ "$pf" = "linux" ];then
    GOOS=linux out="$out-linux"
else    
    GOOS=darwin out="$out-osx"
fi

go build -ldflags "-w -s" -gcflags="all=-trimpath=${PWD}" -asmflags="all=-trimpath=${PWD}" -o $out $workdir/.