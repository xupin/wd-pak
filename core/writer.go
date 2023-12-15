package core

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"time"

	"github.com/xupin/wd-pak/lzss"
	"github.com/xupin/wd-pak/utils"
)

type writer struct {
	File      *os.File
	InDir     string
	size      int64
	startedAt time.Time
	files     int64
}

type fragFile struct {
	Name string
	Path int64
}

var fragFiles []*fragFile

func NewWriter(inDir, file string) *writer {
	f, err := os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		panic(err)
	}
	return &writer{
		File:      f,
		InDir:     inDir,
		startedAt: time.Now(),
	}
}

func (r *writer) Compress() {
	// 取出文件列表
	dirs, files := r.findFiles(r.InDir)
	// 头信息
	r.writeUint32(287454020)
	// 文件数量
	r.writeUint32(uint32((len(dirs) + len(files)) * BLOCK_SIZE))
	// 跳过8字节
	r.writeIgnore()
	// 包大小
	stat, _ := r.File.Stat()
	r.size = stat.Size()
	// 处理打包
	r.process(r.InDir, dirs, files)
}

func (r *writer) Close() {
	r.File.Close()
	fmt.Printf("处理完成！%d 个文件，耗时：%v \n", r.files, time.Since(r.startedAt))
}

func (r *writer) process(curDir string, dirs []string, files []string) {
	curFragFile := r.findFragFile(curDir)
	if curFragFile != nil {
		subFileNum := len(dirs) + len(files)
		fmt.Printf("=>写入文件夹: %s <%d 个文件>\n", curFragFile.Name, subFileNum)
		r.writeFileInfo(curFragFile, uint32(subFileNum)*BLOCK_SIZE, 0)
	}
	// 目录组
	for _, dir := range dirs {
		fragFile := &fragFile{
			Name: curDir + "/" + dir,
			Path: r.size,
		}
		fragFiles = append(fragFiles, fragFile)
		r.writeBlock(TYPE_FOLDER, dir)
	}
	// 文件组
	for _, file := range files {
		fragFile := &fragFile{
			Name: curDir + "/" + file,
			Path: r.size,
		}
		fragFiles = append(fragFiles, fragFile)
		r.writeBlock(TYPE_FILE, file)
	}
	for _, file := range files {
		// 读文件
		f, err := os.Open(curDir + "/" + file)
		if err != nil {
			panic(err)
		}
		fStat, _ := f.Stat()
		fSize := fStat.Size()

		fileBytes := make([]byte, fSize)
		f.Read(fileBytes)
		f.Close()

		// 压缩
		buffer := new(bytes.Buffer)
		writer := lzss.NewWriter(buffer)
		writer.Write(fileBytes)
		writer.Close()

		// 文件大小
		fileCompBytes := buffer.Bytes()

		// 取文件信息
		curFragFile := r.findFragFile(curDir + "/" + file)
		if curFragFile != nil {
			fileCompSize := len(fileCompBytes)
			fmt.Printf("->写入文件: %s <%s>\n", curFragFile.Name, utils.Byte2Str(int64(fileCompSize)))
			r.writeFileInfo(curFragFile, uint32(fileCompSize), uint32(fSize))
		}

		// 写
		r.files += 1
		r.write(fileCompBytes)
	}
	// 递归处理
	for _, dir := range dirs {
		subDirs, subFiles := r.findFiles(curDir + "/" + dir)
		r.process(curDir+"/"+dir, subDirs, subFiles)
	}
}

func (r *writer) findFiles(dir string) ([]string, []string) {
	dirs := make([]string, 0)
	files := make([]string, 0)
	fs, err := os.ReadDir(dir)
	if err != nil {
		panic(err)
	}
	for _, f := range fs {
		if f.Name()[:1] == "." {
			continue
		}
		if f.IsDir() {
			dirs = append(dirs, f.Name())
		} else {
			files = append(files, f.Name())
		}
	}
	return dirs, files
}

// 查询文件
func (r *writer) findFragFile(dir string) *fragFile {
	for _, fragFile := range fragFiles {
		if fragFile.Name != dir {
			continue
		}
		return fragFile
	}
	return nil
}

func (r *writer) writeFileInfo(fragFile *fragFile, size, originSize uint32) {
	// 地址
	r.writeAtUint32(uint32(r.size), fragFile.Path+4)
	// 大小
	r.writeAtUint32(size, fragFile.Path+8)
	// 原始大小
	if originSize > 0 {
		r.writeAtUint32(originSize, fragFile.Path+12)
	} else {
		r.writeAtUint32(size, fragFile.Path+12)
	}
}

func (r *writer) writeUint32(v uint32) {
	bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytes, v)
	r.write(bytes)
}

func (r *writer) writeAtUint32(v uint32, off int64) {
	bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytes, v)
	r.File.WriteAt(bytes, off)
}

func (r *writer) writeIgnore() {
	ignore := make([]byte, 8)
	r.write(ignore)
}

func (r *writer) writeBlock(t uint32, name string) {
	buffer := bytes.Buffer{}
	// 文件类型
	fileType := make([]byte, 4)
	binary.LittleEndian.PutUint32(fileType, t)
	buffer.Write(fileType)
	// 地址、大小、原始大小
	fill := make([]byte, 12)
	buffer.Write(fill)
	// 时间戳
	fileTs := make([]byte, 4)
	binary.LittleEndian.PutUint32(fileTs, uint32(time.Now().Unix()))
	buffer.Write(fileTs)
	// 文件名称
	fileName := []byte(name)
	buffer.Write(fileName)
	fill = make([]byte, 44-len(fileName))
	buffer.Write(fill)
	// 写
	r.write(buffer.Bytes())
}

func (r *writer) write(v []byte) {
	r.File.Write(v)
	r.size += int64(len(v))
}
