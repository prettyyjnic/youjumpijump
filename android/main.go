package main

import (
	"fmt"
	"image/png"
	"log"
	"math"
	"os"
	"os/exec"
	"runtime/debug"
	"strconv"

	jump "github.com/prettyyjnic/youjumpijump"
	"time"
	"path/filepath"
)

var adb string
var basePath string
var tmpScreenShotPath = "/sdcard/screenshot.png"

func init(){
	basePath, _ =  filepath.Abs("./")
	//basePath, _ = filepath.Abs(filepath.Dir(os.Args[0]))
	fmt.Println(basePath)
	adb = basePath + "/platform-tools/adb.exe"
}

func checkErr(err error){
	if err !=nil {
		panic(err)
	}
}

// 截图到电脑
func shotImages() string {
	var err error
	err = exec.Command(adb, "shell", "/system/bin/screencap", "-p", tmpScreenShotPath).Run()
	//err = cmd.Run()
	checkErr(err)
	filename := basePath + "debugger/screenshot" + strconv.FormatInt( time.Now().UnixNano(),
		10) + ".png"
	//保存到电脑
	fmt.Println(filename)
	err = exec.Command(adb, "pull", tmpScreenShotPath, filename).Run()
	checkErr(err)
	return filename
}

func main(){
	var ratio float64
	//var err error
	//if len(os.Args) > 1 {
	//	ratio, err = strconv.ParseFloat(os.Args[1], 10)
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//} else {
	//	fmt.Print("input jump ratio (recommend 2.04):")
	//	_, err = fmt.Scanln(&ratio)
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//}
	ratio = 1
	log.Printf("now jump ratio is %f", ratio)

	filename := shotImages()
	fmt.Println(filename)
	// 查找点

}

//查找下一个要跳的的坐标点
func findNextCoor(filename string){
	infile, err := os.Open(filename)
	checkErr(err)
	pngdec, err := png.Decode(infile)
	checkErr(err)
	// 查找背景色
	// 取左上角和右下角的点的颜色
	maxY := pngdec.Bounds().Max.Y
	maxX := pngdec.Bounds().Max.X

}

func bak() {
	defer func() {
		jump.Debugger()
		if e := recover(); e != nil {
			log.Printf("%s: %s", e, debug.Stack())
			fmt.Print("the program has crashed, press any key to exit")
			var c string
			fmt.Scanln(&c)
		}
	}()

	var ratio float64
	var err error
	if len(os.Args) > 1 {
		ratio, err = strconv.ParseFloat(os.Args[1], 10)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		fmt.Print("input jump ratio (recommend 2.04):")
		_, err = fmt.Scanln(&ratio)
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Printf("now jump ratio is %f", ratio)

	for {
		jump.Debugger()

		_, err := exec.Command("/system/bin/screencap", "-p", "jump.png").Output()
		if err != nil {
			panic(err)
		}
		infile, err := os.Open("jump.png")
		if err != nil {
			panic(err)
		}
		src, err := png.Decode(infile)
		if err != nil {
			panic(err)
		}

		start, end := jump.Find(src)
		if start == nil {
			log.Print("can't find the starting point，please export the debugger directory")
			break
		} else if end == nil {
			log.Print("can't find the end point，please export the debugger directory")
			break
		}

		ms := int(math.Pow(math.Pow(float64(start[0]-end[0]), 2)+math.Pow(float64(start[1]-end[1]), 2), 0.5) * ratio)
		log.Printf("from:%v to:%v press:%vms", start, end, ms)

		_, err = exec.Command("/system/bin/sh", "/system/bin/input", "swipe", "320", "410", "320", "410", strconv.Itoa(ms)).Output()
		if err != nil {
			panic(err)
		}

		infile.Close()
		//time.Sleep(time.Millisecond * 1500)
		//time.Sleep(time.Millisecond * 100)
	}
}
