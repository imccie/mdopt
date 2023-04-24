package main

import (
	"fmt"
	"github.com/satori/go.uuid"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"
)

var wg = sync.WaitGroup{}

func main() {
	println("正在处理目录信息...")
	runtime.GOMAXPROCS(1)
	start := time.Now()
	files := checkExt(".md")
	println("当前目录及子目录中一共有" + strconv.Itoa(len(files)) + "个文件需要处理...")
	for j := 0; j < len(files); j++ {
		wg.Add(1)
		go workOnFile(files[j])
		time.Sleep(5 * 1000)
	}
	wg.Wait()
	elapsedTime := time.Since(start)
	fmt.Println("Total Time For Execution: " + elapsedTime.String())

}

func workOnFile(filestr string) {

	input, err := os.ReadFile(filestr)
	//input, err := ioutil.ReadFile(filestr)
	check(err)

	imagepath := filepath.Dir(filestr) + "/assert"
	//处理文件所属目录中是否有.images目录
	_, err = os.Stat(imagepath)
	if os.IsNotExist(err) {
		err = os.Mkdir(imagepath, 0750)
	}
	//匹配image URL
	imgurlre := regexp.MustCompile(`!\[[^\)]*\]\(http\S*\)`)
	//匹配图像文件名称
	imgext := regexp.MustCompile(`\.(jpg|jpeg|bmp|png|gif)`)
	//获取()中的内容,含()
	kuo := regexp.MustCompile(`\(.*\)`)

	lines := strings.Split(string(input), "\n")
	for lineindex, line := range lines {
		line = strings.ReplaceAll(line, ")", ") ")
		res := imgurlre.FindAllString(line, -1)
		//res,_ := imgurlre.FindStringMatch(line)
		var u1 uuid.UUID
		var toimage string
		if len(res) > 0 {
			for index, restr := range res {
				sur := strings.TrimPrefix(kuo.FindString(restr), "(")
				surl := strings.TrimRight(sur, ")")
				fext := imgext.FindString(surl)
				if fext == "" {
					fext = ".png"
				}
				u1 = uuid.Must(uuid.NewV4(), err)
				filename := imagepath + "/" + u1.String() + fext
				print("处理链接:" + surl)
				err := downloadFile(filename, surl)
				if err != nil {

					println("...下载完成")
					toimage = "![local](" + "assert/" + u1.String() + fext + ")"
					lines[lineindex] = strings.Replace(lines[lineindex], res[index], toimage, -1)
					println("图片本地化转换完成.")
				} else {
					println("下载文件：：" + surl + "---发生错误，将会保持互联网连接！")
				}
			}
		}
	}
	output := strings.Join(lines, "\n")
	err = ioutil.WriteFile(filestr, []byte(output), 0666)
	check(err)
	wg.Done()
}

func downloadFile(filepath string, url string) error {

	defer func() {
		if r := recover(); r != nil {
			log.Println(string(debug.Stack()))
		}
	}()
	// defer wg.Done()
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func checkExt(ext string) []string {
	pathS, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	//var pathS string
	//pathS = "/Users/wujing/Desktop/ceshi"
	var files []string
	filepath.Walk(pathS, func(path string, f os.FileInfo, _ error) error {
		if !f.IsDir() {
			if filepath.Ext(path) == ext {
				files = append(files, path)
			}
		}
		return nil
	})
	return files
}
