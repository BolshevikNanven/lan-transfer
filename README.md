# LAN Transfer / 局域网传输

[English](#english) | [中文](#中文)

---

![home](home.png)

---

<a name="english"></a>
## LAN Transfer (English)

### Description

LAN Transfer is a simple, web-based utility designed for easy file sharing across a local area network (LAN). Built with Go, it offers a clean, modern web interface allowing users to upload files to a central server and download them from any device on the same network. It automatically detects and displays the server's potential LAN IP addresses, simplifying access for other users.

### Features

* **Simple Web Interface:** Clean and intuitive UI for uploading and downloading files.
* **Drag & Drop Uploads:** Supports dragging files directly into the browser for uploading.
* **Upload Progress:** Displays a real-time progress bar during uploads.
* **Automatic IP Detection:** Identifies and shows possible LAN IP addresses for easy access from other devices.
* **Direct Downloads:** Lists uploaded files for straightforward downloading.
* **Self-Contained:** Uses Go's `embed` package to bundle static assets (HTML, CSS), making deployment easy.
* **Modern Design:** Features a UI inspired by Linear, with a focus on usability.

### Project Structure
```
lan-transfer/
├── main.go           # The main Go application file (web server, handlers).
├── static/
│   ├── index.html    # The main HTML template for the web interface.
│   └── style.css     # CSS styles for the web interface.
├── uploads/          # Default directory where uploaded files are stored (created automatically).
└── README.md         # This file.
```

### Getting Started

#### Option 1: Using Pre-compiled Executable (Recommended for most users)

1.  Go to the **Releases** section of this GitHub repository.
2.  Download the latest executable file for your operating system (e.g., `lan-transfer.exe` for Windows).
3.  Run the executable file. It will start the server and display the access addresses in the console.

#### Option 2: Running from Source

1.  **Prerequisites:** Go (version 1.16 or later to support `embed`).
2.  **Clone the repository (or download the files).**
3.  **Navigate to the project directory:**
    ```bash
    cd lan-transfer
    ```
4.  **Run the Go application:**
    ```bash
    go run main.go
    ```

### How to Use

1.  **Access the application:**
    * Open your web browser and go to `http://127.0.0.1:80` (or the port you configure).
    * The application will also print a list of detected LAN IP addresses. Use one of these addresses (e.g., `http://192.168.1.100:80`) to access the application from other devices on the same network.
2.  **Upload:** Drag a file onto the "Upload Files" area or click "Click to select files", then click "Start Upload".
3.  **Download:** Click on any file name in the "Download Files" list.
4.  **Refresh:** Click the refresh button in the "Download Files" section to see newly uploaded files.

---

<a name="中文"></a>
## 局域网传输 (中文)

### 描述

局域网传输 (LAN Transfer) 是一个简单的、基于 Web 的实用工具，专为在局域网 (LAN) 内轻松共享文件而设计。它使用 Go 构建，提供了一个简洁、现代的 Web 界面，允许用户将文件上传到中央服务器，并从同一网络上的任何设备下载它们。它会自动检测并显示服务器可能的局域网 IP 地址，从而简化其他用户的访问。

### 特性

* **简单的 Web 界面**：用于上传和下载文件的简洁直观的用户界面。
* **拖放上传**：支持将文件直接拖到浏览器中进行上传。
* **上传进度**：在上传过程中显示实时进度条。
* **自动 IP 检测**：识别并显示可能的局域网 IP 地址，以便从其他设备轻松访问。
* **直接下载**：列出上传的文件以便直接下载。
* **独立部署**：使用 Go 的 `embed` 包捆绑静态资源 (HTML, CSS)，使部署变得简单。
* **现代设计**：具有受 Linear 启发的 UI，注重可用性。

### 项目结构
```
lan-transfer/
├── main.go           # Go 应用程序主文件 (Web 服务器, 处理器等)。
├── static/
│   ├── index.html    # Web 界面的主 HTML 模板。
│   └── style.css     # Web 界面的 CSS 样式。
├── uploads/          # 存储上传文件的默认目录 (自动创建)。
└── README.md         # 本文档。
```

### 开始使用

#### 方式一：使用预编译的可执行文件 (推荐大多数用户)

1.  前往本 GitHub 仓库的 **Releases** (发布) 部分。
2.  下载适用于您操作系统的最新可执行文件 (例如，Windows 系统的 `lan-transfer.exe`)。
3.  运行该可执行文件。它将启动服务器并在控制台中显示访问地址。

#### 方式二：从源代码运行

1.  **先决条件**：Go (版本 1.16 或更高，以支持 `embed`)。
2.  **克隆仓库 (或下载文件)。**
3.  **进入项目目录:**
    ```bash
    cd lan-transfer
    ```
4.  **运行 Go 应用程序:**
    ```bash
    go run main.go
    ```

### 如何使用

1.  **访问应用程序:**
    * 打开您的 Web 浏览器并访问 `http://127.0.0.1:80` (或您配置的端口)。
    * 应用程序还将打印出检测到的局域网 IP 地址列表。 请使用其中一个地址 (例如 `http://192.168.1.100:80`) 从同一网络上的其他设备访问该应用程序。
2.  **上传**：将文件拖放到“上传文件”区域，或点击“点击选择文件”，然后点击“开始上传”。
3.  **下载**：点击“下载文件”列表中任何文件的名称。
4.  **刷新**：点击“下载文件”部分的刷新按钮以查看新上传的文件。