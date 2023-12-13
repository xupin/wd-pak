package main

import (
	"flag"
	"fmt"

	"github.com/xupin/wd-pak/core"
	"github.com/xupin/wd-pak/utils"
)

var (
	t                    int
	filePath, folderPath string
)

func main() {
	flag.IntVar(&t, "t", 0, "1:解包\n2:打包")
	flag.StringVar(&filePath, "f", "", "pak文件路径")
	flag.StringVar(&folderPath, "o", "", "输出路径")
	flag.Parse()

	if t < 1 || len(filePath) == 0 || len(folderPath) == 0 {
		fmt.Printf("请输入 -h 查看说明（via 90gm.vip） \n")
		return
	}

	if t == 1 {
		if !utils.Exists(filePath) {
			fmt.Printf("pak文件 %s 不存在 \n", filePath)
			return
		}
		pakReader := core.NewReader(filePath, folderPath)
		pakReader.DeCompress()
		pakReader.Close()
	} else if t == 2 {
		if !utils.Exists(folderPath) {
			fmt.Printf("目录 %s 不存在 \n", folderPath)
			return
		}
		pakReader := core.NewWriter(folderPath, filePath)
		pakReader.Compress()
		pakReader.Close()
	}
}
