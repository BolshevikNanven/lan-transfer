package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"math"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

//go:embed static/*
var staticFiles embed.FS

var uploadDir = "uploads" // 上传文件保存目录
var port = "80"           // 监听端口

type PageData struct {
	Addresses []string // 改为字符串切片
	Files     []string
}

// 获取所有可能的局域网 IP 地址
func getLanIPs() []string {
	var ips []string
	interfaces, err := net.Interfaces()
	if err != nil {
		log.Printf("获取网络接口失败: %v", err)
		return []string{"无法获取IP"}
	}

	for _, i := range interfaces {
		// 忽略 down 的接口和环回接口
		if i.Flags&net.FlagUp == 0 || i.Flags&net.FlagLoopback != 0 {
			continue
		}
		// 尝试忽略虚拟机和 Docker 接口 (启发式方法，可能不完全准确)
		if strings.Contains(strings.ToLower(i.Name), "vmnet") ||
			strings.Contains(strings.ToLower(i.Name), "docker") ||
			strings.Contains(strings.ToLower(i.Name), "vbox") {
			continue
		}

		addrs, err := i.Addrs()
		if err != nil {
			log.Printf("获取接口 [%s] 地址失败: %v", i.Name, err)
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			// 只关心 IPv4, 非环回，并且最好是私有地址 (虽然也包括其他)
			if ip == nil || ip.IsLoopback() || ip.To4() == nil {
				continue
			}

			// 添加到列表
			ips = append(ips, ip.String())
		}
	}

	if len(ips) == 0 {
		return []string{"127.0.0.1"} // 如果找不到，返回本地回环
	}

	return ips
}

// 主页处理器
func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFS(staticFiles, "static/index.html")
	if err != nil {
		http.Error(w, "无法解析模板", http.StatusInternalServerError)
		log.Printf("模板解析错误: %v", err)
		return
	}

	files, err := listFiles(uploadDir)
	if err != nil {
		http.Error(w, "无法列出文件", http.StatusInternalServerError)
		log.Printf("列出文件错误: %v", err)
		return
	}

	localIPs := getLanIPs()
	var addresses []string
	for _, ip := range localIPs {
		addresses = append(addresses, fmt.Sprintf("http://%s:%s", ip, port))
	}

	data := PageData{
		Addresses: addresses, // 传递地址列表
		Files:     files,
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "无法执行模板", http.StatusInternalServerError)
		log.Printf("模板执行错误: %v", err)
	}
}

// 文件上传处理器 (保持不变)
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "只允许 POST 方法", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(math.MaxInt64)
	if err != nil {
		sendJSONResponse(w, false, fmt.Sprintf("文件太大或解析错误: %v", err))
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		sendJSONResponse(w, false, fmt.Sprintf("获取文件错误: %v", err))
		return
	}
	defer file.Close()

	_ = os.MkdirAll(uploadDir, os.ModePerm)
	dstPath := filepath.Join(uploadDir, handler.Filename)
	dst, err := os.Create(dstPath)
	if err != nil {
		sendJSONResponse(w, false, fmt.Sprintf("创建文件错误: %v", err))
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		sendJSONResponse(w, false, fmt.Sprintf("保存文件错误: %v", err))
		return
	}

	log.Printf("文件已上传: %s", handler.Filename)
	sendJSONResponse(w, true, "上传成功！")
}

// 发送 JSON 响应 (保持不变)
func sendJSONResponse(w http.ResponseWriter, success bool, message string) {
	response := map[string]any{
		"success": success,
		"message": message,
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("发送 JSON 响应错误: %v", err)
	}
}

// 列出上传目录中的文件 (保持不变)
func listFiles(dir string) ([]string, error) {
	var files []string
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return files, nil
		}
		return nil, err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}
	return files, nil
}

// 静态文件服务器
func staticHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/static/")
	file, err := staticFiles.Open("static/" + path)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		http.Error(w, "无法获取文件信息", http.StatusInternalServerError)
		return
	}

	// *** 关键: 确保 CSS 文件有正确的 MIME 类型 ***
	if strings.HasSuffix(path, ".css") {
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
	}

	// 使用 http.ServeContent, 但需要 ReadSeeker
	// embed.FS 的文件本身不直接实现 ReadSeeker, 但 http.FSFileServer 可以处理
	// 为了简单起见，我们直接读取并写入，或使用 http.ServeFileFS
	// 但这里既然已经打开，手动 ServeContent 更好控制
	// 注意: embed.File 实现了 ReadSeeker, 所以 file.(io.ReadSeeker) 应该是有效的
	http.ServeContent(w, r, stat.Name(), stat.ModTime(), file.(io.ReadSeeker))
}

func main() {
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		log.Fatalf("无法创建上传目录: %v", err)
	}

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/static/", staticHandler) // 确保这个处理器能正确服务 CSS
	http.Handle("/download/", http.StripPrefix("/download/", http.FileServer(http.Dir(uploadDir))))

	localIPs := getLanIPs()

	fmt.Println("=====================================")
	fmt.Println("🚀 LAN File Transfer (Linear Style) 🚀")
	fmt.Printf("💻 本地访问: http://127.0.0.1:%s\n", port)
	fmt.Println("🌐 局域网其他设备请尝试访问以下地址:")
	if len(localIPs) > 0 && localIPs[0] != "无法获取IP" {
		for _, ip := range localIPs {
			fmt.Printf("   -> http://%s:%s\n", ip, port)
		}
	} else {
		fmt.Println("   ! 未找到合适的局域网 IP, 请手动查询并访问。")
	}
	fmt.Println("📂 上传的文件将保存在 'uploads' 目录中。")
	fmt.Println("💡 按 Ctrl+C 停止服务。")
	fmt.Println("=====================================")

	log.Fatal(http.ListenAndServe(":"+port, nil))
}
