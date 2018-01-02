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

	"github.com/prettyyjnic/youjumpijump"
	"path/filepath"
	"strings"
	"image"
	"time"
	"errors"
	"image/draw"
	"image/color"
)

var adb string
var stepCount int // 步数
var basePath string
var tmpScreenShotPath = "/sdcard/screenshot.png"
var startCoor = image.Point{X: 336, Y: 1129} // 起点坐标
var minScanY = 610                           // 开始扫描的点的Y坐标，小于这个坐标的为显示分数的坐标
var defaultBgDistance float64 = 0.96400          // 背景颜色相似度
var defaultRoleDistance float64 = 0.93400          // 背景颜色相似度
var roleRgb = [3]int{55, 60, 102}

func init() {
	basePath, _ = filepath.Abs("./")
	basePath = strings.TrimRight(basePath, "/")
	//basePath, _ = filepath.Abs(filepath.Dir(os.Args[0]))
	fmt.Println(basePath)
	adb = basePath + "/platform-tools/adb.exe"
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func rediretCmd(cmd *exec.Cmd) {
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
}

// 截图到电脑
func shotImages() string {
	exec.Command(adb, "shell", "rm", tmpScreenShotPath).Run()

	var err error
	var cmd *exec.Cmd
	cmd = exec.Command(adb, "shell", "/system/bin/screencap", "-p", tmpScreenShotPath)
	rediretCmd(cmd)
	err = cmd.Run()
	checkErr(err)
	filename := basePath + "/debugger/screenshot" + strconv.Itoa(stepCount) + ".png"
	stepCount++
	//保存到电脑
	cmd = exec.Command(adb, "pull", tmpScreenShotPath, filename)
	err = cmd.Run()
	rediretCmd(cmd)
	checkErr(err)
	return filename
}

func main() {
	//var a = [3]int{53,53,63}
	//var b = [3]int{72, 60, 96}
	//var tmp = (255 - math.Abs(float64(a[0]-b[0]))*0.297 - math.Abs(float64(a[1]-b[1]))*0.593 - math.Abs(float64(
	//	a[2]-b[2]))*0.11 ) / 255
	//
	//fmt.Println(tmp)
	//return
	//return
	//img := decodeImg("./debugger/screenshot67.png")
	//nextCoor, err := findNextCoor(img) // 下一个要跳的点
	//target, err := findCurrentCoor(img)
	//checkErr(err)
	//fmt.Println(nextCoor)
	//fmt.Println(target)
	//return
	//
	//b := colorSimilar(jump.GetRGB(img.ColorModel(), img.At(674, 826)), [3]int{242,216,204}, defaultBgDistance)
	//fmt.Println(b)
	//return

	var ratio float64
	ratio = 1.38
	log.Printf("now jump ratio is %f", ratio)

	for {
		filename := shotImages()
		fmt.Println("保存截图", filename)

		dealFilename := basePath + "/debugger/screenshot" + strconv.Itoa(stepCount-1) + "_deal.png"
		var dealImg draw.Image
		dealFileFp, err := os.Create(dealFilename)
		checkErr(err)

		img := decodeImg(filename)
		nextCoor, err := findNextCoor(img) // 下一个要跳的点
		checkErr(err)
		dealImg = image.NewRGBA(img.Bounds())
		draw.Draw(dealImg, img.Bounds(), img, img.Bounds().Min, draw.Src)
		//描出下一个点
		for i := nextCoor.X - 5; i < nextCoor.X+5; i++ {
			for j := nextCoor.Y - 5; j < nextCoor.Y+5; j++ {
				dealImg.Set(i, j, color.RGBA{0, 255, 0, 255})
			}
		}


		// 保存图片
		currentCoor, err := findCurrentCoor(img)
		checkErr(err)
		//描出当前点
		for i := currentCoor.X - 5; i < currentCoor.X+5; i++ {
			for j := currentCoor.Y - 5; j < currentCoor.Y+5; j++ {
				dealImg.Set(i, j, color.RGBA{255, 0, 0, 255})
			}
		}
		err = png.Encode(dealFileFp, dealImg)
		checkErr(err)
		dealFileFp.Close()

		startCoor = currentCoor
		ms := int(math.Pow(math.Pow(float64(startCoor.X-nextCoor.X), 2)+math.Pow(float64(startCoor.Y-nextCoor.Y), 2), 0.5) * ratio)
		log.Printf("step %d, from:%v to:%v press:%vms", stepCount-1, startCoor, nextCoor, ms)
		_, err = exec.Command(adb, "shell", "/system/bin/input", "swipe", "320", "410", "320", "410", strconv.Itoa(ms)).Output()
		checkErr(err)

		time.Sleep(1500 * time.Millisecond)
	}
}

func calDistance(d1, d2 image.Point) float64 {
	return math.Sqrt(math.Pow(float64((d1.X - d2.X)), 2) + math.Pow(float64((d1.Y - d2.Y)), 2))
}

func findCurrentCoor(pngdec image.Image) (targetPoint image.Point, err error) {
	maxY := pngdec.Bounds().Max.Y
	maxX := pngdec.Bounds().Max.X
	//	扫描获取下一个跳的图形的点，和背影颜色 的相似度大的点
	for x := 10; x < maxX; x++ {
	nextY:
		for y := (maxY * 3 / 4); y > 0; y-- {
			if colorSimilar(jump.GetRGB(pngdec.ColorModel(), pngdec.At(x, y)), roleRgb, 0.95) { // 棋子坐标
				for i := x - 20; i < x+20; i++ {
					if !colorSimilar(jump.GetRGB(pngdec.ColorModel(), pngdec.At(i, y)), roleRgb, 0.95) {
						continue nextY
					}
				}
				for i := y - 20; i < y+20; i++ {
					if !colorSimilar(jump.GetRGB(pngdec.ColorModel(), pngdec.At(x, i)), roleRgb, 0.95) {
						continue nextY
					}
				}

				targetPoint.X = x
				targetPoint.Y = y+90
				fmt.Printf("找到当前点的点坐标(%d, %d)\n", x, y)
				return
			}
		}
	}
	err = errors.New("找不到当前点的坐标")
	return
}

func decodeImg(filename string) image.Image {
	infile, err := os.Open(filename)
	checkErr(err)
	pngdec, err := png.Decode(infile)
	checkErr(err)
	return pngdec
}

//查找下一个要跳的的坐标点
func findNextCoor(pngdec image.Image) (targetPoint image.Point, err error) {
	// 查找背景色
	// 取左上角和右下角的点的颜色
	maxY := pngdec.Bounds().Max.Y
	maxX := pngdec.Bounds().Max.X
	fmt.Println("maxY", maxY, "maxX", maxX)

	leftTopPointColor := jump.GetRGB(pngdec.ColorModel(), pngdec.At(maxX/2, 750))
	fmt.Println("左上角颜色", leftTopPointColor)
	//rightBottomPointColor := jump.GetRGB(pngdec.ColorModel(), pngdec.At(maxX-1, maxY-1))
	rightBottomPointColor := [3]int{leftTopPointColor[0] , leftTopPointColor[1] , leftTopPointColor[2] }
	fmt.Println("右下角颜色", rightBottomPointColor)

	//	因为是渐变颜色的背景，取中值作为背景颜色的相似度计算值
	var bgRgb [3]int
	for i := 0; i < 3; i++ {
		bgRgb[i] = ( leftTopPointColor[i] + rightBottomPointColor[i]) / 2
	}
	fmt.Println("背景颜色", bgRgb)

	var targetColor [3]int
	var targetGraphicsPointArr [4]image.Point // 目标图形 上右下左

	//	扫描获取下一个跳的图形的点，和背影颜色 的相似度大的点

	for y := minScanY; y < maxY; y++ {
		for x := 100; x < maxX; x++ {
			if colorSimilar(jump.GetRGB(pngdec.ColorModel(), pngdec.At(x, y)), bgRgb,
				defaultBgDistance) || colorSimilar(jump.GetRGB(pngdec.ColorModel(), pngdec.At(x, y)), roleRgb,
				defaultRoleDistance) { // 背景颜色 或者 角色
				continue;
			}
			//	扫描到的第一个点为 最上方的点
			if targetGraphicsPointArr[0].X == 0 || targetGraphicsPointArr[0].Y == 0 {
				targetGraphicsPointArr[0].X = x
				targetGraphicsPointArr[0].Y = y
				targetColor = jump.GetRGB(pngdec.ColorModel(), pngdec.At(x, y))
			}
			if targetGraphicsPointArr[0].Y > y {
				targetGraphicsPointArr[0].X = x
				targetGraphicsPointArr[0].Y = y
				targetColor = jump.GetRGB(pngdec.ColorModel(), pngdec.At(x, y))
			}
		}
	}
	fmt.Println("找到最上方的点坐标(", targetGraphicsPointArr[0].X, ",", targetGraphicsPointArr[0].Y, ") 颜色：", targetColor)
	if targetGraphicsPointArr[0].X == 0 || targetGraphicsPointArr[0].Y == 0 {
		err = errors.New("找不到要跳的上方点")
		return
	}
	targetPoint.X = targetGraphicsPointArr[0].X
	targetPoint.Y = targetGraphicsPointArr[0].Y + 70
	return

	// 纯色 目标图形使用 迷宫算法
	//走迷宫算法，获取 另外三个点,
	var nextPoint image.Point
	var hadMove = false
	nextPoint.X = targetGraphicsPointArr[0].X
	nextPoint.Y = targetGraphicsPointArr[0].Y
	var nextDirec int
	for {
		nextPoint, nextDirec = searchNextPoint(pngdec, nextPoint, targetColor, nextDirec)
		fmt.Println("nextPoint", nextPoint)
		fmt.Println("start", targetGraphicsPointArr[0])
		hadMove = true
		if targetGraphicsPointArr[1].X == 0 || nextPoint.X > targetGraphicsPointArr[1].X { // 最右边的点
			targetGraphicsPointArr[1].X = nextPoint.X
			targetGraphicsPointArr[1].Y = nextPoint.Y
		}
		if targetGraphicsPointArr[2].Y == 0 || nextPoint.Y > targetGraphicsPointArr[2].Y { // 最下边的点
			targetGraphicsPointArr[2].X = nextPoint.X
			targetGraphicsPointArr[2].Y = nextPoint.Y
		}
		if targetGraphicsPointArr[3].X == 0 || nextPoint.X < targetGraphicsPointArr[3].X { // 最左边的点
			targetGraphicsPointArr[3].X = nextPoint.X
			targetGraphicsPointArr[3].Y = nextPoint.Y
		}
		if nextPoint.X == targetGraphicsPointArr[0].X && nextPoint.Y == targetGraphicsPointArr[0].Y && hadMove {
			break
		}
	}
	fmt.Println(targetGraphicsPointArr)
	return

}

//迷宫算法，查找下一个点, 左侧搜索算法
func searchNextPoint(pngdec image.Image, start image.Point, targetColor [3]int, direc int) (image.Point, int) {
	if direc != 3 && colorSimilar(jump.GetRGB(pngdec.ColorModel(), pngdec.At(start.X-1, start.Y)), targetColor, 0.98) { // 向左移动一像素
		return image.Point{X: start.X - 1, Y: start.Y}, 1
	}
	if direc != 4 && colorSimilar(jump.GetRGB(pngdec.ColorModel(), pngdec.At(start.X, start.Y+1)), targetColor, 0.98) { // 向下移动一像素
		return image.Point{X: start.X, Y: start.Y + 1}, 2
	}
	if direc != 1 && colorSimilar(jump.GetRGB(pngdec.ColorModel(), pngdec.At(start.X+1, start.Y)), targetColor, 0.98) { // 向右移动一像素
		return image.Point{X: start.X + 1, Y: start.Y}, 3
	}
	if direc != 2 && colorSimilar(jump.GetRGB(pngdec.ColorModel(), pngdec.At(start.X, start.Y-1)), targetColor, 0.98) { // 向上移动一像素
		return image.Point{X: start.X, Y: start.Y - 1}, 4
	}
	return image.Point{X: start.X, Y: start.Y}, 0
}

//颜色相似度
func colorSimilar(a, b [3]int, distance float64) bool {
	//(255 - abs(r1 - r2) * 0.297 - abs(g1 - g2) * 0.593 - abs(b1 - b2) * 0.11) / 255 http://bbs.csdn.net/topics/391015532/ 论坛找到算法，，好像还可以的
	tmp := (255 - math.Abs(float64(a[0]-b[0]))*0.297 - math.Abs(float64(a[1]-b[1]))*0.593 - math.Abs(float64(
		a[2]-b[2]))*0.11 ) / 255
	return tmp > distance
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
