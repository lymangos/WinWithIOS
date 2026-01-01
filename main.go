package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	"github.com/atotto/clipboard"
	"github.com/go-toast/toast"
)

// 配置：文件保存目录（请根据需要修改）
const uploadPath = "C:\\Users\\lymangos\\Desktop\\wiiserver"

// 页面模板：包含上传表单和文件列表链接
const htmlTmpl = `
<!DOCTYPE html>
<html>
<head>
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Win10 USB Link</title>
    <style>
        body { font-family: sans-serif; padding: 20px; text-align: center; }
        .btn { display: block; width: 100%; padding: 15px; margin: 10px 0; background: #007bff; color: white; text-decoration: none; border-radius: 5px; font-size: 1.2em; border: none; cursor: pointer;}
        .btn-green { background: #28a745; }
        input[type="file"] { display: none; }
    </style>
</head>
<body>
    <h1>🔗 USB Link</h1>
    
    <form action="/upload" method="post" enctype="multipart/form-data">
        <label for="file-upload" class="btn">📤 选择文件上传</label>
        <input id="file-upload" type="file" name="file" onchange="this.form.submit()">
    </form>

    <a href="/files/" class="btn btn-green">📂 浏览电脑文件</a>
</body>
</html>
`

type SmsPayload struct {
	Sender  string `json:"sender"`
	Content string `json:"content"`
}

var codeRegex = regexp.MustCompile(`\d{4,6}`)

func handleIndex(w http.ResponseWriter, r *http.Request) {
	t, _ := template.New("index").Parse(htmlTmpl)
	t.Execute(w, nil)
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 限制上传大小 (例如 100MB)
	r.ParseMultipartForm(100 << 20)

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 确保目录存在
	os.MkdirAll(uploadPath, os.ModePerm)

	// [优化]：使用 filepath.Base 清洗文件名，防止路径穿越
	filename := filepath.Base(handler.Filename)
	dstPath := filepath.Join(uploadPath, filename)

	// 创建目标文件
	dst, err := os.Create(dstPath)
	if err != nil {
		http.Error(w, "Error creating file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// 写入
	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}

	fmt.Printf("已接收文件: %s\n", handler.Filename)

	// 上传成功后简单的提示并跳转回主页
	w.Write([]byte("<script>alert('上传成功!'); window.location.href='/';</script>"))
}

func handleSms(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}
	body, _ := io.ReadAll(r.Body)
	var payload SmsPayload
	json.Unmarshal(body, &payload)

	// 提取并复制验证码
	code := codeRegex.FindString(payload.Content)
	if code != "" {
		clipboard.WriteAll(code)
	}

	// 发送通知
	notification := toast.Notification{
		AppID:   "USB Link",
		Title:   "验证码已同步",
		Message: payload.Content,
		// 👇 这里就是加回按钮的关键代码
		Actions: []toast.Action{
			// type: "protocol" 表示这是一个点击后执行协议的按钮
			// label: "查看" 是按钮上显示的文字
			// arguments: "" 这里留空表示不打开特定URL，只激活通知中心
			{"protocol", "浏览文件", "http://localhost:8080/files/"},
		},
	}
	notification.Push()
}
func main() {
	// 路由绑定
	http.HandleFunc("/", handleIndex)        // 主页
	http.HandleFunc("/upload", handleUpload) // 上传接口
	http.HandleFunc("/api/sms", handleSms)   // 短信接口

	// 下载服务 (记得修改为你想要共享的目录)
	fs := http.FileServer(http.Dir(uploadPath))
	http.Handle("/files/", http.StripPrefix("/files/", fs))

	fmt.Println("🚀 服务已启动! 监听端口 :8080")
	fmt.Println("📂 上传目录:", uploadPath)

	// 监听 0.0.0.0 从而允许 USB IP 访问
	log.Fatal(http.ListenAndServe("0.0.0.0:8080", nil))
}
