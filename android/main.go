package main

import (
	"fmt"
	"image/png"
	"log"
	"os"
	"os/exec"
	"strconv"

	"github.com/prettyyjnic/youjumpijump"
	"path/filepath"
	"strings"
	"image"
	"time"
	"io/ioutil"
)

var basePath string

var adb string
var tmpScreenShotPath = "/sdcard/screenshot.png"



func init() {
	basePath, _ = filepath.Abs("./")
	basePath = strings.TrimRight(basePath, "/")
	adb = basePath + "/platform-tools/adb.exe"

	if ok, _ := jump.Exists(basePath + "/debugger"); !ok {
		os.MkdirAll(basePath+"/debugger", os.ModePerm)
	}
}

func initDebuggerDir() {
	files, err := ioutil.ReadDir(basePath + "/debugger")
	checkErr(err)
	for _, f := range files {
		fname := f.Name()
		//fmt.Println(fname)
		os.Remove(basePath + "/debugger/" + fname)
	}
}

func checkErr(err error) {
	if err != nil {
		log.Printf("发生错误：%v",err)
		os.Exit(1)
	}
}

func rediretCmd(cmd *exec.Cmd) {
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
}

// 截图到电脑
func shotImages(stepCount int) string {
	exec.Command(adb, "shell", "rm", tmpScreenShotPath).Run()
	var err error
	var cmd *exec.Cmd
	cmd = exec.Command(adb, "shell", "/system/bin/screencap", "-p", tmpScreenShotPath)
	rediretCmd(cmd)
	err = cmd.Run()
	checkErr(err)
	filename := basePath + "/debugger/" + strconv.Itoa(stepCount) + ".png"

	//保存到电脑
	cmd = exec.Command(adb, "pull", tmpScreenShotPath, filename)
	err = cmd.Run()
	rediretCmd(cmd)
	checkErr(err)
	return filename
}

func check(){
	var img image.Image
	var err error
	var target, bgCoor, nextCoor image.Point
	var maxY, maxX int
	img = decodeImg("./special/46_deal.png")
	target, err = jump.FindCurrentCoor(img)
	checkErr(err)
	maxY = img.Bounds().Max.Y
	maxX = img.Bounds().Max.X
	bgCoor = image.Point{maxX / 2, 35 * maxY / 100}
	nextCoor, err = jump.FindNextCoor(img, target, bgCoor) // 下一个要跳的点
	checkErr(err)

	log.Println("nextCoor (444,919)", nextCoor)
	log.Println("current (680,1063)", target)

	img = decodeImg("./special/16_deal.png")
	target, err = jump.FindCurrentCoor(img)
	checkErr(err)
	maxY = img.Bounds().Max.Y
	maxX = img.Bounds().Max.X
	bgCoor = image.Point{maxX / 2, 35 * maxY / 100}
	nextCoor, err = jump.FindNextCoor(img, target, bgCoor) // 下一个要跳的点
	checkErr(err)

	log.Println("nextCoor (275,821)", nextCoor)
	log.Println("current (843,1139)", target)

	img = decodeImg("./special/119_deal.png")
	target, err = jump.FindCurrentCoor(img)
	checkErr(err)
	maxY = img.Bounds().Max.Y
	maxX = img.Bounds().Max.X
	bgCoor = image.Point{maxX / 2, 35 * maxY / 100}
	nextCoor, err = jump.FindNextCoor(img, target, bgCoor) // 下一个要跳的点
	checkErr(err)

	log.Println("nextCoor (306,835)", nextCoor)
	log.Println("current (856,1152)", target)
}


func main() {
	//var a = [3]int{243, 244, 188}
	//var b = [3]int{246, 246, 246}
	//var tmp = (255 - math.Abs(float64(a[0]-b[0]))*0.297 - math.Abs(float64(a[1]-b[1]))*0.593 - math.Abs(float64(
	//	a[2]-b[2]))*0.11 ) / 255
	//
	//fmt.Println(tmp)
	//return
	//check()
	//return


	var ratio float64
	var stepCount int // 步数
	ratio = 1.44
	log.Printf("now jump ratio is %f\n", ratio)
	go func() {
		for {
			<-time.After(time.Second * 2)
			files, err := ioutil.ReadDir(basePath + "/debugger")
			checkErr(err)

			for _, f := range files {
				fname := f.Name()
				step := strings.TrimRight(fname, "_deal.png")
				//fmt.Println(fname)
				if i, _ := strconv.Atoi(step); i+3 < stepCount { // 只保存最近两张
					os.Remove(basePath + "/debugger/" + fname)
				}
			}
		}
	}()
	initDebuggerDir()

	for {
		stepCount++
		log.Printf("step: %d\n", stepCount)
		filename := shotImages(stepCount)

		infile, err := os.Open(filename)
		checkErr(err)
		img, err := png.Decode(infile)
		checkErr(err)
		ms, err := jump.CalSwipeMs(basePath , stepCount , ratio, img )
		checkErr(err)
		_, err = exec.Command(adb, "shell", "/system/bin/input", "swipe", "320", "410", "320", "410", strconv.Itoa(ms)).Output()
		checkErr(err)
		infile.Close()

		os.Remove(filename) // 删除临时文件
		<-time.After(1500 * time.Millisecond)
	}
}

func decodeImg(filename string) image.Image {
	infile, err := os.Open(filename)
	checkErr(err)
	pngdec, err := png.Decode(infile)
	checkErr(err)
	return pngdec
}
