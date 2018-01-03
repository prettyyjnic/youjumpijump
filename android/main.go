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

var basePath string
var tmpScreenShotPath = "/sdcard/screenshot.png"

var minScanY = 610                                 // 开始扫描的点的Y坐标，小于这个坐标的为显示分数的坐标
var defaultBgDistance float64 = 0.9723254901960784 // 背景颜色相似度
var defaultRoleDistance float64 = 0.94500          // 角色
var roleRgb = [3]int{54, 60, 102}
var minDistinceFromRole = 80
var roleR = 15 // 角色半径
var minPoinDistanct float64 = 39
var red = color.RGBA{255, 0, 0, 255}
var black = color.RGBA{0, 0, 0, 255}
var green = color.RGBA{0, 255, 0, 255}
var blue = color.RGBA{0, 0, 255, 255}
var errCanNotMoveRight = errors.New("不能继续向右移动了")
func init() {
	basePath, _ = filepath.Abs("./")
	basePath = strings.TrimRight(basePath, "/")
	fmt.Println("basePath:" + basePath)
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
		fmt.Println(fname)
		//os.Remove(basePath + "/debugger/" + fname)
	}
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

func main() {
	//var a = [3]int{243, 244, 188}
	//var b = [3]int{246, 246, 246}
	//var tmp = (255 - math.Abs(float64(a[0]-b[0]))*0.297 - math.Abs(float64(a[1]-b[1]))*0.593 - math.Abs(float64(
	//	a[2]-b[2]))*0.11 ) / 255
	//
	//fmt.Println(tmp)
	//return
	//return
	//img := decodeImg("./special/46_deal.png")
	//target, err := findCurrentCoor(img)
	//checkErr(err)
	//maxY := img.Bounds().Max.Y
	//maxX := img.Bounds().Max.X
	//bgCoor := image.Point{maxX / 2, 35 * maxY / 100}
	//nextCoor, err := findNextCoor(img, target, bgCoor) // 下一个要跳的点
	//checkErr(err)
	//fmt.Println(nextCoor)
	//fmt.Println(target)
	//return
	//
	//b := colorSimilar(jump.GetRGB(img.ColorModel(), img.At(674, 826)), [3]int{242,216,204}, defaultBgDistance)
	//fmt.Println(b)
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
					//os.Remove(basePath + "/debugger/" + fname)
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
		//背景点
		// 查找背景色
		// 取左上角和右下角的点的颜色
		maxY := img.Bounds().Max.Y
		maxX := img.Bounds().Max.X
		bgCoor := image.Point{maxX / 2, 39 * maxY / 100}
		bgRgb := jump.GetRGB(img.ColorModel(), img.At(bgCoor.X, bgCoor.Y))
		log.Printf("背景颜色点 %v\n", bgCoor)
		log.Printf("背景颜色: %v\n", bgRgb)

		// 当前点
		startCoor, err := findCurrentCoor(img)
		checkErr(err)
		log.Printf("当前点坐标: %v\n", startCoor)
		// 下一个要跳的点
		nextCoor, err := findNextCoor(img, startCoor, bgCoor)
		checkErr(err)
		log.Printf("要挑的点坐标: %v\n", nextCoor)

		var dealImg draw.Image
		dealImg = image.NewRGBA(img.Bounds())

		draw.Draw(dealImg, img.Bounds(), img, img.Bounds().Min, draw.Src)
		//描出下一个点
		drawPoint(dealImg, nextCoor, green)
		//描出当前点
		drawPoint(dealImg, startCoor, red)
		//描出背景
		drawPoint(dealImg, bgCoor, blue)
		//保存debug 的图
		dealFilename := basePath + "/debugger/" + strconv.Itoa(stepCount) + "_deal.png"
		dealFileFp, err := os.Create(dealFilename)
		checkErr(err)
		err = png.Encode(dealFileFp, dealImg)
		checkErr(err)
		dealFileFp.Close()

		ms := CalSwipeMs(startCoor, nextCoor, ratio)

		log.Printf("step %d, from:%v to:%v press:%vms\n", stepCount, startCoor, nextCoor, ms)
		_, err = exec.Command(adb, "shell", "/system/bin/input", "swipe", "320", "410", "320", "410", strconv.Itoa(ms)).Output()
		checkErr(err)
		infile.Close()

		os.Remove(filename) // 删除临时文件
		<-time.After(1500 * time.Millisecond)
	}
}

// 计算要按住屏幕的时间
func CalSwipeMs(startCoor, nextCoor image.Point, ratio float64) int {
	distance := math.Pow(math.Pow(float64(startCoor.X-nextCoor.X), 2)+math.Pow(float64(startCoor.Y-nextCoor.Y), 2), 0.5)
	ms := int(distance * ratio)

	log.Printf("距离：%f 时间: %d", distance, ms)
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

	for y := (maxY * 60 / 100); y > 0; y-- {
	nextY:
		for x := maxX/10; x < maxX; x++ {
			if colorSimilar(jump.GetRGB(pngdec.ColorModel(), pngdec.At(x, y)), roleRgb, defaultRoleDistance) { // 棋子坐标
				for i := x - roleR; i < x+roleR; i++ {
					if !colorSimilar(jump.GetRGB(pngdec.ColorModel(), pngdec.At(i, y)), roleRgb, defaultRoleDistance) {
						continue nextY
					}
				}
				for i := y - roleR; i < y; i++ {
					if !colorSimilar(jump.GetRGB(pngdec.ColorModel(), pngdec.At(x, i)), roleRgb, defaultRoleDistance) {
						continue nextY
					}
				}

				targetPoint.X = x + 4
				targetPoint.Y = y
				//log.Printf("找到当前点的点坐标(%d, %d)\n", x, y)
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
func findNextCoor(pngdec image.Image, startCoor, bgCoor image.Point) (targetPoint image.Point, err error) {
	// 查找背景色
	// 取左上角和右下角的点的颜色
	maxY := pngdec.Bounds().Max.Y
	maxX := pngdec.Bounds().Max.X

	bgRgb := jump.GetRGB(pngdec.ColorModel(), pngdec.At(bgCoor.X, bgCoor.Y))
	log.Printf("背影颜色：%v", bgRgb)
	var targetColor [3]int
	var leftTopPoint image.Point // 目标图形 上右下左

	//	扫描获取下一个跳的图形的点，和背影颜色 的相似度大的点

	for y := minScanY; y < maxY; y++ {
		for x := maxX / 10; x < maxX; x++ {
			//colorSimilar(jump.GetRGB(pngdec.ColorModel(), pngdec.At(x, y)), roleRgb,
			//	defaultRoleDistance
			if colorSimilar(jump.GetRGB(pngdec.ColorModel(), pngdec.At(x, y)), bgRgb,
				defaultBgDistance) || (( startCoor.X-minDistinceFromRole ) < x && (startCoor.X+minDistinceFromRole ) > x ) { // 背景颜色 或者 跟角色的 Y 轴 + - 20 px
				continue;
			}
			//	扫描到的第一个点为 最上方的点
			if leftTopPoint.X == 0 || leftTopPoint.Y == 0 {
				leftTopPoint.X = x
				leftTopPoint.Y = y
				targetColor = jump.GetRGB(pngdec.ColorModel(), pngdec.At(x, y))
			}
			if leftTopPoint.Y > y {
				leftTopPoint.X = x
				leftTopPoint.Y = y
				targetColor = jump.GetRGB(pngdec.ColorModel(), pngdec.At(x, y))
			}
		}
	}
	log.Println("找到最上方的点坐标(", leftTopPoint.X, ",", leftTopPoint.Y, ") 颜色：", targetColor)
	if leftTopPoint.X == 0 || leftTopPoint.Y == 0 {
		err = errors.New("找不到要跳的上方点")
		return
	}
	//使用纯色搜索算法
	targetPoint, err = Maze(pngdec, leftTopPoint, targetColor)
	log.Printf("纯色搜索算法结果：%v", targetPoint)
	if err != nil {
		targetPoint.X = leftTopPoint.X
		targetPoint.Y = leftTopPoint.Y + 70
		log.Printf("使用模糊结果：%v", targetPoint)
		err = nil
	}
	return

}

// 向左下角进行搜索
func Maze(pngdec image.Image, startPoint image.Point, targetColor [3]int) (point image.Point, err error) {
	var targetGraphicsPointArr [4]image.Point
	// 纯色 目标图形使用 迷宫算法
	//走迷宫算法，获取 另外三个点,
	var nextPoint image.Point
	for i := 0; i < 4; i++ {
		targetGraphicsPointArr[i].X = startPoint.X
		targetGraphicsPointArr[i].Y = startPoint.Y
	}

	for {
		var errTmp error
		nextPoint, errTmp = searchLeftNextPoint(pngdec, startPoint, targetColor)
		if nextPoint.X == startPoint.X && nextPoint.Y == startPoint.Y { // 没有进行移动
			break
		}
		startPoint.X = nextPoint.X
		startPoint.Y = nextPoint.Y
		if nextPoint.X > targetGraphicsPointArr[1].X { // 最右边的点
			targetGraphicsPointArr[1].X = nextPoint.X
			targetGraphicsPointArr[1].Y = nextPoint.Y
		}
		if nextPoint.Y > targetGraphicsPointArr[2].Y { // 最下边的点
			targetGraphicsPointArr[2].X = nextPoint.X
			targetGraphicsPointArr[2].Y = nextPoint.Y
		}
		if nextPoint.X < targetGraphicsPointArr[3].X { // 最左边的点
			targetGraphicsPointArr[3].X = nextPoint.X
			targetGraphicsPointArr[3].Y = nextPoint.Y
		}
		if errTmp !=nil && errTmp == errCanNotMoveRight {
			break
		}
	}
	// 获取中点
	point.X = (targetGraphicsPointArr[0].X + targetGraphicsPointArr[2].X) / 2
	point.Y = (targetGraphicsPointArr[0].Y + targetGraphicsPointArr[2].Y) / 2

	//检查点是否正确
	leftTopDistinct := math.Abs(float64(targetGraphicsPointArr[3].X-targetGraphicsPointArr[0].X)) // 最左距离最上的x 距离
	leftBottomDistinct := math.Abs(float64(targetGraphicsPointArr[3].X-targetGraphicsPointArr[2].X)) // 最左距离最下的x 距离
	topDistinct := math.Abs(float64(targetGraphicsPointArr[2].Y-targetGraphicsPointArr[0].Y))
	log.Printf("纯色算法结果 %v left: %d top: %d\n", targetGraphicsPointArr, leftTopDistinct, topDistinct)
	if leftTopDistinct < minPoinDistanct || topDistinct < minPoinDistanct || leftBottomDistinct < minPoinDistanct {
		err = errors.New("点距过小，不合适")
	}
	return
}

//迷宫算法，查找下一个点, 向下方进行搜索
func searchLeftNextPoint(pngdec image.Image, start image.Point, targetColor [3]int) (image.Point, error) {
	var distinct = 0.995555
	if colorSimilar(jump.GetRGB(pngdec.ColorModel(), pngdec.At(start.X-1, start.Y)), targetColor, distinct) { // 向左移动一像素
	//fmt.Printf("左移动 %v\n", image.Point{X: start.X - 1, Y: start.Y})
		return image.Point{X: start.X - 1, Y: start.Y}, nil
	}
	if colorSimilar(jump.GetRGB(pngdec.ColorModel(), pngdec.At(start.X, start.Y+1)), targetColor, distinct) { // 向下移动一像素
		//fmt.Printf("下移动 %v\n", image.Point{X: start.X , Y: start.Y+1})
		return image.Point{X: start.X, Y: start.Y + 1}, nil
	}
	//向右下方移动
	for x := start.X; x < pngdec.Bounds().Max.X; x++ {
		//fmt.Printf("右移动 %v\n", image.Point{X: x, Y: start.Y})
		if !colorSimilar(jump.GetRGB(pngdec.ColorModel(), pngdec.At(x, start.Y)), targetColor, distinct) { // 右边不是可以移动的点
			return image.Point{X: x, Y: start.Y}, errCanNotMoveRight
		}
		if colorSimilar(jump.GetRGB(pngdec.ColorModel(), pngdec.At(x, start.Y+1)), targetColor, distinct) {
			return image.Point{X: x, Y: start.Y + 1},nil
		}
	}
	return image.Point{X: start.X, Y: start.Y}, nil
}

//颜色相似度
func colorSimilar(a, b [3]int, distance float64) bool {
	//(255 - abs(r1 - r2) * 0.297 - abs(g1 - g2) * 0.593 - abs(b1 - b2) * 0.11) / 255 http://bbs.csdn.net/topics/391015532/ 论坛找到算法，，好像还可以的
	tmp := (255 - math.Abs(float64(a[0]-b[0]))*0.297 - math.Abs(float64(a[1]-b[1]))*0.593 - math.Abs(float64(
		a[2]-b[2])) * 0.11 ) / 255
	return tmp > distance
}
