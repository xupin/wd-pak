package core

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path"
	"sync"
	"time"

	"github.com/xupin/wd-pak/lzss"
	"github.com/xupin/wd-pak/utils"
)

type reader struct {
	File      *os.File
	OutDir    string
	wg        *sync.WaitGroup
	startedAt time.Time
	files     int64
}

func NewReader(file, outDir string) *reader {
	f, err := os.Open(file)
	if err != nil {
		panic(err)
	}
	fs, _ := os.ReadDir(outDir)
	for _, f1 := range fs {
		os.RemoveAll(path.Join([]string{outDir, f1.Name()}...))
	}
	return &reader{
		File:      f,
		OutDir:    outDir,
		wg:        &sync.WaitGroup{},
		startedAt: time.Now(),
	}
}

func (r *reader) DeCompress() {
	// 头信息
	if r.byte2uint32() != HEAD {
		panic("file format not supported")
	}
	// 文件大小
	fileSize := r.byte2uint32()
	// 跳过8字节
	ignore := make([]byte, 8)
	if _, err := r.File.Read(ignore); err != nil {
		panic(err)
	}
	// 创建文件夹
	if !utils.Exists(r.OutDir) {
		os.Mkdir(r.OutDir, os.ModePerm)
	}
	// 读文件夹内容
	fileNum := fileSize / uint32(BLOCK_SIZE)
	for i := 1; i <= int(fileNum); i++ {
		r.process(r.OutDir)
		// 指针重新移动至文件信息区
		r.File.Seek(int64(HEAD_SIZE+(i)*BLOCK_SIZE), io.SeekStart)
	}
}

func (r *reader) Close() {
	r.wg.Wait()
	r.File.Close()
	fmt.Printf("处理完成！%d 个文件，耗时：%v \n", r.files, time.Since(r.startedAt))
}

func (r *reader) process(curDir string) {
	// 文件类型
	fileType := r.byte2uint32()
	// 地址
	filePath := r.byte2uint32()
	// 大小
	fileSize := r.byte2uint32()
	// 原始大小
	fileOriginSize := r.byte2uint32()
	// 时间戳
	r.byte2uint32()
	// 文件名
	fileName := r.readName()
	// 指针移动至文件内容区
	r.File.Seek(int64(filePath), io.SeekStart)
	// 处理
	if fileType == TYPE_FILE {
		compBytes := make([]byte, fileSize)
		if _, err := r.File.Read(compBytes); err != nil {
			panic(err)
		}
		r.files += 1
		r.wg.Add(1)
		go func() {
			defer r.wg.Done()
			fmt.Printf("->解压文件：%s/%s <%s>\n", curDir, fileName, utils.Byte2Str(int64(fileOriginSize)))
			// 解压写文件
			deCompBytes := make([]byte, fileOriginSize)
			reader := lzss.NewReader(compBytes)
			reader.Read(deCompBytes)
			f1, err := os.OpenFile(curDir+"/"+fileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
			if err != nil {
				panic(err)
			}
			defer f1.Close()
			for i, b := range deCompBytes {
				if b != byte(0) {
					continue
				}
				r := rune(' ')
				sBytes := []byte(string(r))
				deCompBytes[i] = sBytes[0]
			}
			f1.Write(deCompBytes)
		}()
	} else {
		// 子文件数量
		subFileNum := fileSize / uint32(BLOCK_SIZE)
		// 创建文件夹
		curDir += "/" + fileName
		if !utils.Exists(curDir) {
			os.Mkdir(curDir, os.ModePerm)
			fmt.Printf("=>解压文件夹：%s <%d 个文件> \n", fileName, subFileNum)
		}
		// 读文件夹内容
		for i := 1; i <= int(subFileNum); i++ {
			r.process(curDir)
			// 指针重新移动至文件信息区
			r.File.Seek(int64(filePath+uint32(i*BLOCK_SIZE)), io.SeekStart)
		}
	}
}

func (r *reader) byte2uint32() uint32 {
	bytes := make([]byte, 4)
	if _, err := r.File.Read(bytes); err != nil {
		panic(err)
	}
	return binary.LittleEndian.Uint32(bytes)
}

// 剩余44字节
func (r *reader) readName() string {
	bytes := make([]byte, 44)
	if _, err := r.File.Read(bytes); err != nil {
		panic(err)
	}
	for i, b := range bytes {
		if b == byte(0) {
			bytes = bytes[:i]
			break
		}
	}
	return string(bytes)
}
