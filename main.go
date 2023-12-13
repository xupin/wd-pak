package main

import "github.com/xupin/wd-pak/pak"

func main() {
	pakWriter := pak.NewWriter("./out", "./aaa.pak")
	pakWriter.Compress()
	pakWriter.Close()

	pakReader := pak.NewReader("./aaa.pak", "./out1")
	pakReader.DeCompress()
	pakReader.Close()
}
