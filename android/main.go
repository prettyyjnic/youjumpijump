package main

import (
	"fmt"
	"image/png"
	"log"
	"math"
	"os"
	"os/exec"
	"strconv"

	"github.com/prettyyjnic/youjumpijump"
	"path/filepath"
	"strings"
	"image"
	"time"
	"errors"
	"image/draw"
	"image/color"
	"io/ioutil"
)

var adb string
var stepCount int // 步数
var basePath string
var tmpScreenShotPath = "/sdcard/screenshot.png"

var minScanY = 610                        // 开始扫描的点的Y坐标，小于这个坐标的为显示分数的坐标
var defaultBgDistance float64 = 0.96400   // 背景颜色相似度
var defaultRoleDistance float64 = 0.93400 // 背景颜色相似度
var roleRgb = [3]int{55, 60, 102}

var red = color.RGBA{255, 0, 0, 255}
var black = color.RGBA{0, 0, 0, 255}
var green = color.RGBA{0, 255, 0, 255}
var blue = color.RGBA{0, 0, 255, 255}

func init() {
	basePath, _ = filepath.Abs("./")
	basePath = strings.TrimRight(basePath, "/")
	fmt.Println("basePath:" + basePath)
	adb = basePath + "/platform-tools/adb.exe"

	if ok, _ := jump.Exists(basePath + "/debugger"); !ok {
		os.MkdirAll(basePath+"/debugger", os.ModePerm)
	}
	files, err := ioutil.ReadDir(basePath + "/debugger")
	checkErr(err)
	for _, f := range files {
		fname := f.Name()
		os.Remove(basePath + "/debugger/" + fname)
	}
}

//删除已经没有用了的图片
func dealTmpImages(){
	go func() {
		for{
			<- time.After(time.Second*2)
			files, err := ioutil.ReadDir(basePath + "/debugger")
			checkErr(err)

			for _, f := range files {
				fname := f.Name()
				fname = strings.TrimRight(fname, "_deal.png")
				if i, _:= strconv.Atoi(fname ) ; i + 2 < stepCount {// 只保存最近两张
					os.Remove(basePath + "/debugger/" + fname)
				}
			}
		}
	}()

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
	filename := basePath + "/debugger/" + strconv.Itoa(stepCount) + ".png"
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
	dealTmpImages()
	for {
		filename := shotImages()
		fmt.Println("保存截图", filename)


		infile, err := os.Open(filename)
		checkErr(err)
		img, err := png.Decode(infile)
		checkErr(err)
		// 当前点
		startCoor, err := findCurrentCoor(img)
		checkErr(err)

		nextCoor, err := findNextCoor(img, startCoor) // 下一个要跳的点
		checkErr(err)

		var dealImg draw.Image
		dealImg = image.NewRGBA(img.Bounds())

		draw.Draw(dealImg, img.Bounds(), img, img.Bounds().Min, draw.Src)
		//描出下一个点
		drawPoint(dealImg, nextCoor, green)
		//描出当前点
		drawPoint(dealImg, startCoor, red)
		//保存debug 的图
		dealFilename := basePath + "/debugger/" + strconv.Itoa(stepCount-1) + "_deal.png"
		dealFileFp, err := os.Create(dealFilename)
		checkErr(err)
		err = png.Encode(dealFileFp, dealImg)
		checkErr(err)
		dealFileFp.Close()

		ms := CalSwipeMs(startCoor, nextCoor, ratio)

		log.Printf("step %d, from:%v to:%v press:%vms", stepCount-1, startCoor, nextCoor, ms)
		_, err = exec.Command(adb, "shell", "/system/bin/input", "swipe", "320", "410", "320", "410", strconv.Itoa(ms)).Output()
		checkErr(err)
		infile.Close()

		os.Remove(filename)// 删除临时文件
		time.Sleep(1500 * time.Millisecond)
	}
}

// 计算要按住屏幕的时间
func CalSwipeMs(startCoor, nextCoor image.Point, ratio float64) int {
	ms := int(math.Pow(math.Pow(float64(startCoor.X-nextCoor.X), 2)+math.Pow(float64(startCoor.Y-nextCoor.Y), 2), 0.5) * ratio)
	return ms
}

func drawPoint(dealImg draw.Image, currentCoor image.Point, rgbColor color.RGBA) {
	for i := currentCoor.X - 5; i < currentCoor.X+5; i++ {
		for j := currentCoor.Y - 5; j < currentCoor.Y+5; j++ {
			dealImg.Set(i, j, rgbColor)
		}
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
				targetPoint.Y = y + 90
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
func findNextCoor(pngdec image.Image, startCoor image.Point) (targetPoint image.Point, err error) {
	// 查找背景色
	// 取左上角和右下角的点的颜色
	maxY := pngdec.Bounds().Max.Y
	maxX := pngdec.Bounds().Max.X
	fmt.Println("maxY", maxY, "maxX", maxX)

	bgRgb := jump.GetRGB(pngdec.ColorModel(), pngdec.At(maxX/2, 39 / 100 * maxY))
	fmt.Println("背景颜色", bgRgb)

	var targetColor [3]int
	var targetGraphicsPointArr [4]image.Point // 目标图形 上右下左

	//	扫描获取下一个跳的图形的点，和背影颜色 的相似度大的点

	for y := minScanY; y < maxY; y++ {
		for x := 100; x < maxX; x++ {
			//colorSimilar(jump.GetRGB(pngdec.ColorModel(), pngdec.At(x, y)), roleRgb,
			//	defaultRoleDistance
			if colorSimilar(jump.GetRGB(pngdec.ColorModel(), pngdec.At(x, y)), bgRgb,
				defaultBgDistance) ||  (( startCoor.X -20 ) < x && (startCoor.X + 20 ) > x ) { // 背景颜色 或者 跟角色的 Y 轴 + - 20 px
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



}

//没有用
func maze(pngdec image.Image, targetColor [3]int){
	var targetGraphicsPointArr [4]image.Point
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

//迷宫算法，查找下一个点, 左侧搜索算法, 没有用
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
		a[2]-b[2])) * 0.11 ) / 255
	return tmp > distance
}
