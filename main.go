package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	. "log"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
)

type conf struct {
	ResDir      string
	CcbListFile string
}

func doPick(resDir string, ccbPath string, group string, output string,loger *Logger) {
	fmt.Println("\n查找到ccb："+ccbPath)
	loger.Println("\n查找到ccb："+ccbPath)
	fmt.Println("开始处理")
	loger.Println("开始处理")
	ccbPath = strings.Replace(ccbPath, "\\", "/", -1)
	ccbFile, err := ioutil.ReadFile(ccbPath)
	if err != nil {
		loger.Println("[ERROR]:读取CCB文件失败：" + ccbPath, err.Error())
	}
	ccbstr := string(ccbFile)
	//3.筛选ccb中的png文件
	var reg, _ = regexp.Compile(`[\w|/]+.png`)
	pngs := reg.FindAllString(ccbstr, -1)
	tmps := strings.Split(ccbPath, "/")
	ccbn := tmps[len(tmps)-1]
	ccbnamestr := strings.Split(ccbn, ".")
	ccbname := ccbnamestr[0]

	//4.从png目录中拉取目标文件
	if group != "" {
		output = output + group + "/"
	}
	output = output + ccbname + "/"
	for _, png := range pngs {
		path := resDir + "/" + png
		fmt.Println("匹配到图片-"+png)
		loger.Println("匹配到图片-"+png)
		pngFileData, err := ioutil.ReadFile(path)
		if err != nil {
			fmt.Println("[ERROR]:读取png文件失败", err.Error())
			loger.Println("[ERROR]:读取png文件失败", err.Error())
		} else {
			if err := os.MkdirAll(output, 0777);err != nil{
				fmt.Println("[ERROR]:创建目录失败失败-" + output, err.Error())
				loger.Println("[ERROR]:创建目录失败失败-" + output, err.Error())
			}
			index := strings.LastIndex(png,"/")
			pngName := png[index+1:]
			outFile := output + pngName
			if err := ioutil.WriteFile(outFile, pngFileData, os.ModePerm); err != nil {
				fmt.Println("[ERROR]:写入png文件失败-" + outFile, err.Error())
				loger.Println("[ERROR]:写入png文件失败-" + outFile, err.Error())
			}
		}
	}
	fmt.Println("处理完毕")
	loger.Println("处理完毕")
}

func main() {
    pwd, err := filepath.Abs(filepath.Dir(os.Args[0]))
    if err != nil {
        panic(err)
    }
	var loger *Logger
	if _, e := os.Stat(pwd+"/log"); e != nil {
		logFile, _ := os.Create(pwd+"/log")
		defer logFile.Close()
		loger = New(logFile, "", 0)
	} else {
		logFile, _ := os.OpenFile(pwd+"/log", os.O_APPEND, 0666)
		defer logFile.Close()
		loger = New(logFile, "", 0)
	}

	resDir := "."
	ccbListFile := ""
	//1.从配置文件读取png目录和输出目录
	fmt.Println("读取配置文件...")
	loger.Println("读取配置文件...")
	conFile := pwd+"/ini.json"
	file, err := ioutil.ReadFile(conFile)
	if err != nil {
		fmt.Println("[ERROR]:读取配置文件失败...", err.Error())
		loger.Panicln("[ERROR]:读取配置文件失败...", err.Error())
	} else {
		var cfst conf
		if err = json.Unmarshal(file, &cfst); err != nil {
			fmt.Println("[ERROR]:解析配置文件失败...", err.Error())
			loger.Panicln("[ERROR]:解析配置文件失败...", err.Error())
		} else {
			resDir = cfst.ResDir
			ccbListFile = cfst.CcbListFile
		}
	}
	if strings.LastIndex(resDir,"./") >= 0 {
		resDir = strings.Replace(resDir,"./",pwd+"/",1)
	}
	if strings.LastIndex(ccbListFile,"./") >= 0 {
		ccbListFile = strings.Replace(ccbListFile,"./",pwd+"/",1)
	}
	fmt.Println("resDir:"+resDir, "ccbListFile:"+ccbListFile)
	loger.Println("resDir:"+resDir, "ccbListFile:"+ccbListFile)

	//2.解析ccblist
	fmt.Println("读取CCB列表文件...")
	loger.Println("读取CCB列表文件...")
	ccblist := map[string]interface{}{}
	ccblistfileObj, err := ioutil.ReadFile(ccbListFile)
	if err != nil {
		fmt.Println("[ERROR]:读取CCB列表文件失败...", err.Error())
		loger.Panicln("[ERROR]:读取CCB列表文件失败...", err.Error())
	} else {
		if err = json.Unmarshal(ccblistfileObj, &ccblist); err != nil {
			fmt.Println("[ERROR]:解析CCB列表文件失败...", err.Error())
			loger.Panicln("[ERROR]:解析CCB列表文件失败...", err.Error())
		}
	}

	//删除输出目录
	if info, err := os.Stat(pwd+"/output/"); err == nil && info.IsDir() {
		fmt.Println("删除output目录")
		loger.Println("删除output目录")
		if err := os.RemoveAll(pwd+"/output/");err != nil{
			fmt.Println("[ERROR]:删除output目录失败",err.Error())
			loger.Panicln("[ERROR]:删除output目录失败",err.Error())
		}
	}

	//3.读取ccb
	fmt.Println("查找ccb...")
	loger.Println("查找ccb...")
	for entry := range ccblist {
		fmt.Println(entry, ccblist[entry])
		if reflect.Slice == reflect.TypeOf(ccblist[entry]).Kind() {
			s := reflect.ValueOf(ccblist[entry])
			for i := 0; i < s.Len(); i++ {
				ele := s.Index(i)
				ccbName := ele.Interface().(string)
				ccbPath := resDir + "/layout/" + ccbName + ".ccb"
				fmt.Println("###", ccbPath)
				doPick(resDir, ccbPath, entry,pwd+"/output/",loger)
			}
		} else if reflect.String == reflect.TypeOf(ccblist[entry]).Kind() {
			var ccbName = ccblist[entry].(string)
			ccbPath := resDir + "/layout/" + ccbName + ".ccb"
			fmt.Println("###", ccbPath)
			doPick(resDir, ccbPath, "",pwd+"/output/",loger)
		}
	}
}
