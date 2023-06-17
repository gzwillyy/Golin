package dirscan

import (
	"fmt"
	"github.com/parnurzeal/gorequest"
	"github.com/spf13/cobra"
	"golin/global"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	spinnerChars = []string{"|", "/", "-", "\\"} //进度条更新动画
	code         []string                        //扫描的状态
	mu           sync.Mutex
	counter      = 0 //当前已扫描的数量，计算百分比
	ch           = make(chan struct{}, 30)
	wg           = sync.WaitGroup{}
	proxyurl     = ""     //代理地址
	file         = ""     //读取的字典文件
	checkurl     []string //扫描的url切片
	countall     = 0
	request      *gorequest.SuperAgent
)

func ParseFlags(cmd *cobra.Command, args []string) {
	url, _ := cmd.Flags().GetString("url")
	url = strings.TrimSuffix(url, "/") // 删除最后的/
	chcount, _ := cmd.Flags().GetInt("chan")
	timeout, _ := cmd.Flags().GetInt("timeout")
	ch = make(chan struct{}, chcount)

	proxyurl, _ = cmd.Flags().GetString("proxy") //不为空则设置代理
	if proxyurl != "" {
		request = gorequest.New().Proxy(proxyurl).Timeout(time.Duration(timeout) * time.Second)
	} else {
		request = gorequest.New().Timeout(time.Duration(timeout) * time.Second)
	}
	Agent, _ := cmd.Flags().GetString("Agent") //如果Agent不为空则自定义User-Agent
	if Agent != "" {
		request.Set("User-Agent", Agent)
	}

	file, _ = cmd.Flags().GetString("file") //读取字典文件
	if !global.PathExists(file) {
		fmt.Printf("[-] url字典文件不存在！通过-f 指定！\n")
		os.Exit(0)
	} else {
		data, _ := os.ReadFile(file)
		str := strings.ReplaceAll(string(data), "\r\n", "\n")
		for _, u := range strings.Split(str, "\n") {
			checkurl = append(checkurl, u)
		}
		countall = len(removeDuplicates(checkurl)) //去重
		if countall == 0 {
			fmt.Printf("[-] url为空！\n")
			os.Exit(0)
		}
	}
	waittime, _ := cmd.Flags().GetInt("wait")      //循环超时
	codese, _ := cmd.Flags().GetString("code")     //搜索的状态码
	for _, s := range strings.Split(codese, ",") { //根据分隔符写入到状态切片
		code = append(code, s)
	}
	codestr := strings.Join(code, ", ")

	fmt.Printf("[*] 开始运行dirsearch模式 字典位置%s 共计尝试:%d次 超时等待:%d/s 循环等待:%d/s 并发数:%d 寻找状态码:%s 代理地址:%s\n ", file, countall, timeout, waittime, chcount, codestr, proxyurl)
	for _, checku := range removeDuplicates(checkurl) {
		if checku[0] != '/' { //判断第一个字符是不是/不是的话则增加
			checku = "/" + checku
		}
		ch <- struct{}{}
		wg.Add(1)
		go isStatusCodeOk(fmt.Sprintf("%s%s", url, checku)) //传递完整url\

		if waittime > 0 { //延迟等待
			time.Sleep(time.Duration(waittime) * time.Second)
		}
	}
	wg.Wait()
	time.Sleep(time.Second * 1) //等待1秒是因为并发问题，等待进度条。
	percent()
	fmt.Println()

}

// removeDuplicates 切片去重
func removeDuplicates(slice []string) []string {
	keys := make(map[string]bool)
	var list []string
	for _, entry := range slice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}