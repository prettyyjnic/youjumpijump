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
	"runtime/debug"
	rand2 "math/rand"
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
		log.Printf("%s", debug.Stack())
		fmt.Print("the program has crashed, press any key to exit")
		var c string
		fmt.Scanln(&c)
		os.Exit(0)
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

// 测试用
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
	log.Println("current (850,1139)", target)

	img = decodeImg("./special/119_deal.png")
	target, err = jump.FindCurrentCoor(img)
	checkErr(err)
	maxY = img.Bounds().Max.Y
	maxX = img.Bounds().Max.X
	bgCoor = image.Point{maxX / 2, 35 * maxY / 100}
	nextCoor, err = jump.FindNextCoor(img, target, bgCoor) // 下一个要跳的点
	checkErr(err)

	log.Println("nextCoor (306,835)", nextCoor)
	log.Println("current (838,1007)", target)

	img = decodeImg("./special/50_deal.png")
	target, err = jump.FindCurrentCoor(img)
	checkErr(err)
	maxY = img.Bounds().Max.Y
	maxX = img.Bounds().Max.X
	bgCoor = image.Point{maxX / 2, 35 * maxY / 100}
	nextCoor, err = jump.FindNextCoor(img, target, bgCoor) // 下一个要跳的点
	checkErr(err)

	log.Println("nextCoor (438,888)", nextCoor)
	log.Println("current (694,1071)", target)

}

// 小米5s可以正常运行
func main() {
	//var a = [3]int{243, 244, 188}
	//var b = [3]int{246, 246, 246}
	//var tmp = (255 - math.Abs(float64(a[0]-b[0]))*0.297 - math.Abs(float64(a[1]-b[1]))*0.593 - math.Abs(float64(
	//	a[2]-b[2]))*0.11 ) / 255
	//
	//fmt.Println(tmp)
	//return
	// check()
	 //return
	var ratio float64
	var err error
	var stepCount int // 步数
	ratio = 1.44

	r := rand2.New(rand2.NewSource(time.Now().UnixNano()))

	fmt.Print("input jump ratio (recommend 1.44):")
	_, err = fmt.Scanln(&ratio)
	if err != nil {
		log.Printf("input is empty, will use 1.44 as default ratio")
		ratio = 1.44
	}
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
		// 获取截图
		infile, err := os.Open(filename)
		checkErr(err)
		img, err := png.Decode(infile)
		checkErr(err)
		//计算应该跳的时间
		ms, err := jump.CalSwipeMs(basePath , stepCount , ratio, img )
		checkErr(err)
		x:= strconv.Itoa( r.Intn(50)+200 )
		y:= strconv.Itoa( r.Intn(50)+300 )
		_, err = exec.Command(adb, "shell", "/system/bin/input", "swipe", x, y, x, y, strconv.Itoa(ms)).Output()
		checkErr(err)
		infile.Close()

		os.Remove(filename) // 删除临时文件

		<-time.After( time.Duration(ms * 2 + r.Intn(100)) * time.Millisecond  )
	}
}


//测试用
func decodeImg(filename string) image.Image {
	infile, err := os.Open(filename)
	checkErr(err)
	pngdec, err := png.Decode(infile)
	checkErr(err)
	return pngdec
}
