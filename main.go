package main

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
    "regexp"
    "strings"
    "sync"
    "time"
)

const numPages = 100 // 总共需要抓取的页面数
const itemsPerPage = 5 // 每页数据项数量

type Comment struct {
    Email string `json:"email"`
}

func main() {
    start := time.Now() // 记录程序开始时间
    var wg sync.WaitGroup
    comments := make(chan Comment, numPages*itemsPerPage) // 创建一个缓冲通道，用于存储抓取到的邮件地址

    // 循环抓取每一页的数据
    for i := 1; i <= numPages; i++ {
        url := fmt.Sprintf("https://jsonplaceholder.typicode.com/posts/%d/comments", i)
        wg.Add(1) // 每次开启协程前将 WaitGroup 计数器加一
        go getComments(url, &wg, comments) // 开启一个协程来抓取数据
    }

    wg.Wait() // 等待所有协程完成任务

    // 将抓取到的邮件地址写入文件
    fileContents := make([]string, 0, numPages*itemsPerPage)
    for i := 0; i < numPages*itemsPerPage; i++ {
        comment := <-comments
        fileContents = append(fileContents, comment.Email)
    }
    ioutil.WriteFile("emails.txt", []byte(strings.Join(fileContents, "\n")), 0644)

    elapsed := time.Since(start) // 计算整个程序运行耗时
    fmt.Printf("整个程序运行耗时 %s\n", elapsed)
}

func getComments(url string, wg *sync.WaitGroup, comments chan<- Comment) {
    defer wg.Done() // 协程完成任务后将 WaitGroup 计数器减一

    resp, err := http.Get(url)
    if err != nil {
        fmt.Println("Error fetching URL:", url)
        return
    }
    defer resp.Body.Close()

    bodyBytes, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        fmt.Println("Error reading response body:", err)
        return
    }

    var commentData []Comment
    err = json.Unmarshal(bodyBytes, &commentData)
    if err != nil {
        fmt.Println("Error decoding JSON data:", err)
        return
    }

    for _, comment := range commentData {
        if isValidEmail(comment.Email) {
            comments <- comment
        }
    }
}

func isValidEmail(email string) bool {
    // 正则表达式判断邮件地址是否合法
    emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
    return emailRegex.MatchString(email)
}
