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

var uploadDir = "uploads"
var port = "80"

type PageData struct {
	Addresses []string
	Files     []string
}

// getLanIPs attempts to find likely LAN IP addresses for the host.
func getLanIPs() []string {
	var ips []string
	interfaces, err := net.Interfaces()
	if err != nil {
		log.Printf("Failed to get network interfaces: %v", err)
		return []string{"Unable to get IP"}
	}

	for _, i := range interfaces {
		// Ignore down interfaces and loopback
		if i.Flags&net.FlagUp == 0 || i.Flags&net.FlagLoopback != 0 {
			continue
		}
		// Heuristic to ignore common VM/Docker interfaces
		if strings.Contains(strings.ToLower(i.Name), "vmnet") ||
			strings.Contains(strings.ToLower(i.Name), "docker") ||
			strings.Contains(strings.ToLower(i.Name), "vbox") {
			continue
		}

		addrs, err := i.Addrs()
		if err != nil {
			log.Printf("Failed to get addresses for interface [%s]: %v", i.Name, err)
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

			// We only care about IPv4, non-loopback addresses.
			if ip == nil || ip.IsLoopback() || ip.To4() == nil {
				continue
			}

			ips = append(ips, ip.String())
		}
	}

	if len(ips) == 0 {
		return []string{"127.0.0.1"} // Fallback to loopback if no suitable IP is found
	}

	return ips
}

// indexHandler serves the main HTML page.
func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFS(staticFiles, "static/index.html")
	if err != nil {
		http.Error(w, "Could not parse template", http.StatusInternalServerError)
		log.Printf("Template parsing error: %v", err)
		return
	}

	files, err := listFiles(uploadDir)
	if err != nil {
		http.Error(w, "Could not list files", http.StatusInternalServerError)
		log.Printf("File listing error: %v", err)
		return
	}

	localIPs := getLanIPs()
	var addresses []string
	for _, ip := range localIPs {
		addresses = append(addresses, fmt.Sprintf("http://%s:%s", ip, port))
	}

	data := PageData{
		Addresses: addresses,
		Files:     files,
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Could not execute template", http.StatusInternalServerError)
		log.Printf("Template execution error: %v", err)
	}
}

// uploadHandler handles file uploads via POST request.
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(math.MaxInt64) // Use a large limit
	if err != nil {
		sendJSONResponse(w, false, fmt.Sprintf("File too large or form parsing error: %v", err))
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		sendJSONResponse(w, false, fmt.Sprintf("Error retrieving file: %v", err))
		return
	}
	defer file.Close()

	_ = os.MkdirAll(uploadDir, os.ModePerm) // Ensure upload directory exists
	dstPath := filepath.Join(uploadDir, handler.Filename)
	dst, err := os.Create(dstPath)
	if err != nil {
		sendJSONResponse(w, false, fmt.Sprintf("Error creating file: %v", err))
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		sendJSONResponse(w, false, fmt.Sprintf("Error saving file: %v", err))
		return
	}

	log.Printf("File uploaded successfully: %s", handler.Filename)
	sendJSONResponse(w, true, "Upload successful!")
}

// sendJSONResponse sends a standardized JSON response.
func sendJSONResponse(w http.ResponseWriter, success bool, message string) {
	response := map[string]any{
		"success": success,
		"message": message,
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error sending JSON response: %v", err)
	}
}

// listFiles reads the upload directory and returns a list of filenames.
func listFiles(dir string) ([]string, error) {
	var files []string
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return files, nil // If dir doesn't exist, return empty list
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

// staticHandler serves files embedded in the binary.
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
		http.Error(w, "Could not get file info", http.StatusInternalServerError)
		return
	}

	// Ensure CSS files have the correct MIME type for proper browser rendering.
	if strings.HasSuffix(path, ".css") {
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
	}

	// Serve the content using http.ServeContent, which handles range requests and sets headers.
	// embed.File implements io.ReadSeeker, so type assertion is valid.
	http.ServeContent(w, r, stat.Name(), stat.ModTime(), file.(io.ReadSeeker))
}

func main() {
	// Ensure the upload directory exists on startup.
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		log.Fatalf("Could not create upload directory: %v", err)
	}

	// Register HTTP handlers.
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/static/", staticHandler)
	http.Handle("/download/", http.StripPrefix("/download/", http.FileServer(http.Dir(uploadDir))))

	localIPs := getLanIPs()

	// Print startup information to the console.
	fmt.Println("=====================================")
	fmt.Println("🚀 LAN File Transfer (Linear Style) 🚀")
	fmt.Printf("💻 Local access: http://127.0.0.1:%s\n", port)
	fmt.Println("🌐 Try these addresses on other LAN devices:")
	if len(localIPs) > 0 && localIPs[0] != "Unable to get IP" {
		for _, ip := range localIPs {
			fmt.Printf("   -> http://%s:%s\n", ip, port)
		}
	} else {
		fmt.Println("   ! Could not find suitable LAN IPs. Please check manually.")
	}
	fmt.Printf("📂 Uploaded files will be saved in the '%s' directory.\n", uploadDir)
	fmt.Println("💡 Press Ctrl+C to stop the server.")
	fmt.Println("=====================================")

	// Start the web server.
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
