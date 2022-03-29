package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var urllist []string

//扫描请求
func httpres(url string, port int, c chan string, wgscan *sync.WaitGroup) {
	//参数1，扫描使用的协议，参数2，IP+端口号，参数3，设置连接超时的时间
	_, err := net.DialTimeout("tcp", url+":"+strconv.Itoa(port), time.Second)
	if err == nil {
		c <- (url + "\t" + strconv.Itoa(port))
	}
	wgscan.Done()
}

//文本读取成ip
func fileread(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("open file err:", err.Error())
		return
	}
	defer file.Close()
	r := bufio.NewReader(file) //建立缓冲区，把文件内容放到缓冲区中
	for {
		// 分行读取文件  ReadLine返回单个行，不包括行尾字节(\n  或 \r\n)
		data, _, err := r.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("read err", err.Error())
			break
		}
		// 打印出内容
		//fmt.Printf("%v", string(data))
		urllist = append(urllist, string(data))
	}
}

//获取所有端口
func getAllPort(port *string) ([]int, error) {
	var ports []int
	//处理 ","号 如 80,81,88 或 80,88-100
	portArr := strings.Split(strings.Trim(*port, ","), ",")
	for _, v := range portArr {
		portArr2 := strings.Split(strings.Trim(v, "-"), "-")
		startPort, err := filterPort(portArr2[0])
		if err != nil {
			continue
		}
		//第一个端口先添加
		ports = append(ports, startPort)
		if len(portArr2) > 1 {
			//添加第一个后面的所有端口
			endPort, _ := filterPort(portArr2[1])
			if endPort > startPort {
				for i := 1; i <= endPort-startPort; i++ {
					ports = append(ports, startPort+i)
				}
			}
		}
	}
	//去重复
	ports = arrayUnique(ports)

	return ports, nil
}

//端口合法性过滤
func filterPort(str string) (int, error) {
	port, err := strconv.Atoi(str)
	if err != nil {
		return 0, err
	}
	if port < 1 || port > 65535 {
		return 0, errors.New("端口号范围超出")
	}
	return port, nil
}

//数组去重
func arrayUnique(arr []int) []int {
	var newArr []int
	for i := 0; i < len(arr); i++ {
		repeat := false
		for j := i + 1; j < len(arr); j++ {
			if arr[i] == arr[j] {
				repeat = true
				break
			}
		}
		if !repeat {
			newArr = append(newArr, arr[i])
		}
	}
	return newArr
}

func main() {
	ip := flag.String("u", "", "ip or url")
	file := flag.String("r", "", "url list file")
	ports := flag.String("p", "80-1000", "端口号范围 例如:-p=80,81,88-1000")
	flag.Parse()
	var wg sync.WaitGroup
	portall, _ := getAllPort(ports)
	c := make(chan string, 500) //通道定义
	start := time.Now()
	if *ip != "" {
		urllist = append(urllist, *ip)
		wg.Add(len(portall)) //计数器，只有带那个计数器为0才执行某个操作
		/*
			遍历url数组进行扫描
		*/
		for i := range urllist {
			for a := range portall {
				go func(url string, port int) {
					httpres(url, port, c, &wg)
				}(urllist[i], portall[a])
			}
		}
		go func() {
			wg.Wait() //计数器等待，为0即关闭通道
			close(c)
		}()
		for i := range c {
			fmt.Println(i)
		}
		end := time.Since(start)
		fmt.Println("花费时间为:", end)
	}
	if *file != "" {
		fileread(*file)
		wg.Add(len(portall) * len(urllist)) //计数器，只有带那个计数器为0才执行某个操作
		/*
			遍历url数组进行扫描
		*/
		for i := range urllist {
			for a := range portall {
				go func(url string, port int) {
					httpres(url, port, c, &wg)
				}(urllist[i], portall[a])
			}
		}
		go func() {
			wg.Wait() //计数器等待，为0即关闭通道
			close(c)
		}()
		for i := range c {
			fmt.Println(i)
		}
		end := time.Since(start)
		fmt.Println("花费时间为:", end)
	}
	if *file == "" && *ip == "" {
		flag.Usage()
	}
}
