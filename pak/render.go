package pak

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/xupin/wd-pak/lzss"
	"github.com/xupin/wd-pak/utils"
)

type reader struct {
	File   *os.File
	OutDir string
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
		File:   f,
		OutDir: outDir,
	}
}

func (r *reader) DeCompress() {
	// 头信息
	if r.byte2uint32() != HEAD {
		panic("file format not supported")
	}
	// 文件数量
	fileNum := r.byte2uint32() / uint32(BLOCK_SIZE)
	// 跳过8字节
	ignore := make([]byte, 8)
	if _, err := r.File.Read(ignore); err != nil {
		panic(err)
	}
	// 创建文件夹
	if !utils.Exists(r.OutDir) {
		os.Mkdir(r.OutDir, os.ModePerm)
	}
	for i := 1; i <= int(fileNum); i++ {
		// 文件类型
		fileType := r.byte2uint32()
		// 文件夹
		if fileType == TYPE_FOLDER {
			r.deFolder(r.OutDir)
		} else {
			r.deFile(r.OutDir)
		}
		// 移动指针至文件信息区
		r.File.Seek(int64(HEAD_SIZE+(i)*BLOCK_SIZE), io.SeekStart)
	}
}

func (r *reader) Close() {
	r.File.Close()
}

func (r *reader) byte2uint32() uint32 {
	bytes := make([]byte, 4)
	if _, err := r.File.Read(bytes); err != nil {
		panic(err)
	}
	return binary.LittleEndian.Uint32(bytes)
}

func (r *reader) deFolder(curDir string) {
	// 地址
	filePath := r.byte2uint32()
	// 大小
	fileSize := r.byte2uint32()
	// 原始大小
	fileOriginSize := r.byte2uint32()
	// 时间戳
	fileTs := r.byte2uint32()
	// 目录名
	fileName := r.readName()
	// 创建文件夹
	folderPath := curDir + "/" + fileName
	if !utils.Exists(folderPath) {
		os.Mkdir(folderPath, os.ModePerm)
		fmt.Printf("->创建文件夹: %s [%d %d] \n", fileName, fileOriginSize, fileTs)
	}
	// 移动指针至文件内容区
	r.File.Seek(int64(filePath), io.SeekStart)
	// 文件数量
	fileNum := fileSize / uint32(BLOCK_SIZE)
	for i := 1; i <= int(fileNum); i++ {
		// 文件类型
		fileType := r.byte2uint32()
		// 文件夹
		if fileType == TYPE_FOLDER {
			r.deFolder(folderPath)
		} else {
			r.deFile(folderPath)
		}
		// 移动指针至文件信息区
		r.File.Seek(int64(filePath+uint32(i*BLOCK_SIZE)), io.SeekStart)
	}
}

func (r *reader) deFile(curDir string) {
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
	// 移动指针至文件内容区
	r.File.Seek(int64(filePath), io.SeekStart)
	compBytes := make([]byte, fileSize)
	if _, err := r.File.Read(compBytes); err != nil {
		panic(err)
	}
	fmt.Printf("->正在解压: %s \n", fileName)
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
