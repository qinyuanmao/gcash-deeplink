package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/qinyuanmao/gcash-deeplink/generator"
	"github.com/qinyuanmao/gcash-deeplink/models"
	"github.com/qinyuanmao/gcash-deeplink/parser"
)

func main() {
	// 显示欢迎信息
	printBanner()

	// 运行示例
	if len(os.Args) > 1 && os.Args[1] == "examples" {
		runExamples()
		return
	}

	// 启动 HTTP API 服务器
	startHTTPServer()
}

func printBanner() {
	banner := `
╔═══════════════════════════════════════════════════════════╗
║                                                           ║
║      GCash Deep Link Generator                            ║
║      EMVCo QR Code → GCash Deep Link                      ║
║                                                           ║
╚═══════════════════════════════════════════════════════════╝
`
	fmt.Println(banner)
}

func runExamples() {
	fmt.Println("=== 运行示例 ===\n")

	// 示例 1: 基础用法
	example1()

	fmt.Println("\n" + strings.Repeat("=", 80) + "\n")

	// 示例 2: 带回调的支付
	example2()

	fmt.Println("\n" + strings.Repeat("=", 80) + "\n")

	// 示例 3: 多种策略
	example3()
}

func example1() {
	fmt.Println("【示例 1: 基础用法】\n")

	qrCode := "00020101021228530011ph.ppmi.p2m0111SRCPPHM2XXX0312MRCHNT-4H3TZ05030005204519953036085406100.005802PH5925SOCMED DIGITAL MARKETING 6010MakatiCity62650010ph.starpay0315SOCMED DIGITAL 0509OR#1Z1CSC0708TodayPay0803***88290012ph.ppmi.qrph0109OR#1Z1CSC63040275"

	// 解析 QR Code
	p := parser.NewEMVCoParser()
	data, err := p.Parse(qrCode)
	if err != nil {
		log.Fatal(err)
	}

	// 打印解析结果
	fmt.Println(p.GetSummary(data))
	fmt.Println()

	// 生成 Deep Link
	g := generator.NewDeepLinkGenerator()
	result, err := g.Generate(data, nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("生成的 Deep Link:")
	fmt.Println(result.DeepLink)
}

func example2() {
	fmt.Println("【示例 2: 带回调的订单支付】\n")

	qrCode := "00020101021228790011ph.ppmi.p2m0111PAEYPHM2XXX0324VkHUE2Fz8Ee2YxnTVPX34TZs0410030300288605030105204739953036085406100.005802PH5916NEXA ONLINE SHOP6013General Trias62430012ph.ppmi.qrph0306wWMBdH05062110000803***88440012ph.ppmi.qrph0124VkHUE2Fz8Ee2YxnTVPX34TZs63041C3C"

	g := generator.NewDeepLinkGenerator()

	options := &models.DeepLinkOptions{
		PaymentType: models.PaymentTypeDynamic,
		OrderID:     fmt.Sprintf("ORDER-%d", time.Now().Unix()),
		RedirectURL: "https://myshop.com/payment/success",
		NotifyURL:   "https://myshop.com/api/gcash/webhook",
		MerchantID:  "217020000119199251998",
	}

	result, err := g.GenerateWithValidation(qrCode, options)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("订单号: %s\n", options.OrderID)
	fmt.Printf("商户: %s\n", result.ParsedData.MerchantName)
	fmt.Printf("金额: ₱%s\n", result.ParsedData.Amount)
	fmt.Println()
	fmt.Println("生成的 Deep Link:")
	fmt.Println(result.DeepLink)
}

func example3() {
	fmt.Println("【示例 3: 生成多种策略的 Deep Link】\n")

	qrCode := "00020101021228530011ph.ppmi.p2m0111SRCPPHM2XXX0312MRCHNT-4H3TZ05030005204519953036085406100.005802PH5925SOCMED DIGITAL MARKETING 6010MakatiCity62650010ph.starpay0315SOCMED DIGITAL 0509OR#1Z1CSC0708TodayPay0803***88290012ph.ppmi.qrph0109OR#1Z1CSC63040275"

	p := parser.NewEMVCoParser()
	data, _ := p.Parse(qrCode)

	g := generator.NewDeepLinkGenerator()
	strategies := g.GenerateMultiple(data)

	for name, link := range strategies {
		fmt.Printf("策略: %s\n", name)
		fmt.Printf("链接: %s\n\n", link)
	}
}

// HTTP API 服务器
func startHTTPServer() {
	// 静态文件服务器
	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/", fs)

	// API 端点
	http.HandleFunc("/api/parse", handleParse)
	http.HandleFunc("/api/generate", handleGenerate)
	http.HandleFunc("/api/validate", handleValidate)
	http.HandleFunc("/health", handleHealth)

	serverURL := "http://localhost:9000"

	fmt.Println("🚀 HTTP API 服务启动")
	fmt.Println("📍 地址：" + serverURL)
	fmt.Println("🌐 Web 界面：" + serverURL)
	fmt.Println("\n可用端点：")
	fmt.Println("  GET    /               - Web 界面")
	fmt.Println("  POST   /api/parse      - 解析 EMVCo QR Code")
	fmt.Println("  POST   /api/generate   - 生成 GCash Deep Link")
	fmt.Println("  POST   /api/validate   - 验证 QR Code")
	fmt.Println("  GET    /health         - 健康检查")
	fmt.Println()

	// 自动打开浏览器
	go openBrowser(serverURL)

	log.Fatal(http.ListenAndServe(":9000", enableCORS(http.DefaultServeMux)))
}

// API 处理函数
func handleParse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "只支持 POST 请求", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		QRCode string `json:"qrCode"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "无效的 JSON",
		})
		return
	}

	p := parser.NewEMVCoParser()
	data, err := p.Parse(req.QRCode)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    data,
	})
}

func handleGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "只支持 POST 请求", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		QRCode      string `json:"qrCode"`
		OrderID     string `json:"orderId,omitempty"`
		MerchantID  string `json:"merchantId,omitempty"`
		RedirectURL string `json:"redirectUrl,omitempty"`
		NotifyURL   string `json:"notifyUrl,omitempty"`
		PaymentType string `json:"paymentType,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "无效的 JSON",
		})
		return
	}

	if req.QRCode == "" {
		respondJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "qrCode 不能为空",
		})
		return
	}

	options := &models.DeepLinkOptions{
		OrderID:     req.OrderID,
		MerchantID:  req.MerchantID,
		RedirectURL: req.RedirectURL,
		NotifyURL:   req.NotifyURL,
	}

	if req.PaymentType != "" {
		options.PaymentType = models.PaymentType(req.PaymentType)
	}

	g := generator.NewDeepLinkGenerator()
	result, err := g.GenerateWithValidation(req.QRCode, options)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	respondJSON(w, http.StatusOK, result)
}

func handleValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "只支持 POST 请求", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		QRCode string `json:"qrCode"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "无效的 JSON",
		})
		return
	}

	p := parser.NewEMVCoParser()
	validation := p.Validate(req.QRCode)

	respondJSON(w, http.StatusOK, validation)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "healthy",
		"service": "GCash Deep Link Generator",
		"time":    time.Now().Format(time.RFC3339),
	})
}

// 辅助函数
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func enableCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		h.ServeHTTP(w, r)
	})
}

// 自动打开浏览器
func openBrowser(url string) {
	// 等待服务器启动
	time.Sleep(500 * time.Millisecond)

	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}

	if err != nil {
		fmt.Printf("⚠️  无法自动打开浏览器: %v\n", err)
		fmt.Printf("请手动访问: %s\n", url)
	} else {
		fmt.Printf("✅ 已自动打开浏览器\n")
	}
}
