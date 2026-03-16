//go:build ignore
// +build ignore

package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/qinyuanmao/gcash-deeplink/generator"
	"github.com/qinyuanmao/gcash-deeplink/models"
	"github.com/qinyuanmao/gcash-deeplink/parser"
)

// DeepLinkDebugger Deep Link 调试工具
type DeepLinkDebugger struct{}

// NewDeepLinkDebugger 创建调试器
func NewDeepLinkDebugger() *DeepLinkDebugger {
	return &DeepLinkDebugger{}
}

// DebugDeepLink 调试 Deep Link
func (d *DeepLinkDebugger) DebugDeepLink(deepLink string) {
	fmt.Println("=== Deep Link 调试信息 ===\n")

	// 1. 解析 URL
	parsedURL, err := url.Parse(deepLink)
	if err != nil {
		fmt.Printf("❌ URL 解析失败: %v\n", err)
		return
	}

	fmt.Printf("✅ URL Scheme: %s\n", parsedURL.Scheme)
	fmt.Printf("✅ URL Host: %s\n", parsedURL.Host)
	fmt.Printf("✅ URL Path: %s\n\n", parsedURL.Path)

	// 2. 解析查询参数
	params := parsedURL.Query()

	fmt.Println("📋 查询参数：")
	fmt.Println(strings.Repeat("-", 80))

	// 关键参数检查
	criticalParams := []string{
		"qrCode",
		"orderAmount",
		"merchantName",
		"qrCodeFormat",
		"sub",
		"clientId",
	}

	for _, key := range criticalParams {
		value := params.Get(key)
		if value != "" {
			fmt.Printf("✅ %-20s: %s\n", key, d.truncate(value, 60))
		} else {
			fmt.Printf("❌ %-20s: [缺失]\n", key)
		}
	}

	fmt.Println()

	// 3. 可选参数
	optionalParams := []string{
		"merchantId",
		"orderId",
		"tfrbnkcode",
		"shopId",
		"tfrAcctNo",
		"acqInfo",
		"redirectUrl",
		"returnUrl",
		"notifyUrl",
		"callbackUrl",
		"param3",
		"param5",
		"bizNo",
	}

	fmt.Println("📋 可选参数：")
	fmt.Println(strings.Repeat("-", 80))

	for _, key := range optionalParams {
		value := params.Get(key)
		if value != "" {
			fmt.Printf("✅ %-20s: %s\n", key, d.truncate(value, 60))
		}
	}

	fmt.Println()

	// 4. 验证 QR Code（如果存在）
	if qrCode := params.Get("qrCode"); qrCode != "" {
		fmt.Println("🔍 QR Code 验证：")
		fmt.Println(strings.Repeat("-", 80))

		p := parser.NewEMVCoParser()
		validation := p.Validate(qrCode)

		if validation.Valid {
			fmt.Println("✅ QR Code 有效")

			// 解析详细信息
			data, err := p.Parse(qrCode)
			if err == nil {
				fmt.Printf("   商户: %s\n", data.MerchantName)
				fmt.Printf("   城市: %s\n", data.MerchantCity)
				fmt.Printf("   金额: ₱%s\n", data.Amount)
				fmt.Printf("   店铺ID: %s\n", data.ShopID)
				fmt.Printf("   银行代码: %s\n", data.BankCode)
			}
		} else {
			fmt.Println("❌ QR Code 无效：")
			for _, err := range validation.Errors {
				fmt.Printf("   - %s\n", err)
			}
		}
	}

	fmt.Println()

	// 5. 安全检查
	fmt.Println("🔒 安全检查：")
	fmt.Println(strings.Repeat("-", 80))

	d.checkSecurity(params)

	fmt.Println()

	// 6. 兼容性检查
	fmt.Println("📱 兼容性检查：")
	fmt.Println(strings.Repeat("-", 80))

	d.checkCompatibility(parsedURL, params)
}

// DebugQRCode 调试 QR Code
func (d *DeepLinkDebugger) DebugQRCode(qrCode string) {
	fmt.Println("=== QR Code 调试信息 ===\n")

	p := parser.NewEMVCoParser()

	// 1. 基本验证
	fmt.Println("📋 基本验证：")
	fmt.Println(strings.Repeat("-", 80))

	validation := p.Validate(qrCode)
	if validation.Valid {
		fmt.Println("✅ QR Code 格式有效")
	} else {
		fmt.Println("❌ QR Code 格式无效：")
		for _, err := range validation.Errors {
			fmt.Printf("   - %s\n", err)
		}
		return
	}

	fmt.Println()

	// 2. 详细解析
	fmt.Println("📋 解析结果：")
	fmt.Println(strings.Repeat("-", 80))

	data, err := p.Parse(qrCode)
	if err != nil {
		fmt.Printf("❌ 解析失败: %v\n", err)
		return
	}

	// 输出 JSON
	jsonData, _ := json.MarshalIndent(data, "", "  ")
	fmt.Println(string(jsonData))

	fmt.Println()

	// 3. 关键字段检查
	fmt.Println("🔍 关键字段检查：")
	fmt.Println(strings.Repeat("-", 80))

	checks := []struct {
		name  string
		value string
		valid bool
	}{
		{"版本号", data.Version, data.Version == "01"},
		{"商户名称", data.MerchantName, data.MerchantName != ""},
		{"商户城市", data.MerchantCity, data.MerchantCity != ""},
		{"金额", data.Amount, data.Amount != ""},
		{"货币代码", data.Currency, data.Currency == "608"},
		{"国家代码", data.CountryCode, data.CountryCode == "PH"},
		{"店铺 ID", data.ShopID, data.ShopID != ""},
		{"银行代码", data.BankCode, data.BankCode != ""},
	}

	for _, check := range checks {
		if check.valid {
			fmt.Printf("✅ %-15s: %s\n", check.name, check.value)
		} else {
			fmt.Printf("⚠️  %-15s: %s (可能有问题)\n", check.name, check.value)
		}
	}
}

// truncate 截断长字符串
func (d *DeepLinkDebugger) truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// checkSecurity 安全检查
func (d *DeepLinkDebugger) checkSecurity(params url.Values) {
	// 检查是否有 SQL 注入特征
	dangerousChars := []string{"'", "\"", ";", "--", "/*", "*/", "DROP", "DELETE", "INSERT"}

	for key, values := range params {
		for _, value := range values {
			valueUpper := strings.ToUpper(value)
			for _, dangerous := range dangerousChars {
				if strings.Contains(valueUpper, strings.ToUpper(dangerous)) {
					fmt.Printf("⚠️  参数 '%s' 包含潜在危险字符: %s\n", key, dangerous)
				}
			}
		}
	}

	// 检查 URL 长度
	totalLen := 0
	for key, values := range params {
		for _, value := range values {
			totalLen += len(key) + len(value) + 2 // key=value&
		}
	}

	if totalLen > 2048 {
		fmt.Printf("⚠️  URL 总长度过长 (%d 字符)，可能导致兼容性问题\n", totalLen)
	} else {
		fmt.Printf("✅ URL 长度正常 (%d 字符)\n", totalLen)
	}
}

// checkCompatibility 兼容性检查
func (d *DeepLinkDebugger) checkCompatibility(parsedURL *url.URL, params url.Values) {
	// 检查 URL Scheme
	if parsedURL.Scheme == "gcash" {
		fmt.Println("✅ URL Scheme 正确 (gcash://)")
	} else {
		fmt.Printf("❌ URL Scheme 不正确: %s (应该是 gcash://)\n", parsedURL.Scheme)
	}

	// 检查必需参数
	requiredParams := []string{"qrCode", "orderAmount", "merchantName"}
	missing := []string{}

	for _, param := range requiredParams {
		if params.Get(param) == "" {
			missing = append(missing, param)
		}
	}

	if len(missing) > 0 {
		fmt.Printf("⚠️  缺少必需参数: %v\n", missing)
	} else {
		fmt.Println("✅ 所有必需参数都存在")
	}

	// 检查 param3 格式
	if param3 := params.Get("param3"); param3 != "" {
		if strings.Contains(param3, "99960005~ph.ppmi.p2m") {
			fmt.Println("✅ param3 格式正确")
		} else {
			fmt.Println("⚠️  param3 格式可能不正确")
		}
	}
}

// CompareDeepLinks 比较两个 Deep Link
func (d *DeepLinkDebugger) CompareDeepLinks(link1, link2 string) {
	fmt.Println("=== Deep Link 对比 ===\n")

	url1, _ := url.Parse(link1)
	url2, _ := url.Parse(link2)

	params1 := url1.Query()
	params2 := url2.Query()

	// 找出所有参数
	allKeys := make(map[string]bool)
	for key := range params1 {
		allKeys[key] = true
	}
	for key := range params2 {
		allKeys[key] = true
	}

	fmt.Println("参数对比：")
	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("%-20s | %-25s | %-25s\n", "参数", "Link 1", "Link 2")
	fmt.Println(strings.Repeat("-", 80))

	for key := range allKeys {
		val1 := params1.Get(key)
		val2 := params2.Get(key)

		status := "="
		if val1 != val2 {
			status = "≠"
		}

		fmt.Printf("%-20s | %-25s | %-25s [%s]\n",
			key,
			d.truncate(val1, 25),
			d.truncate(val2, 25),
			status,
		)
	}
}

// GenerateTestDeepLink 生成测试用 Deep Link
func (d *DeepLinkDebugger) GenerateTestDeepLink(qrCode string) {
	fmt.Println("=== 生成测试 Deep Link ===\n")

	g := generator.NewDeepLinkGenerator()

	// 策略 1: 最简化（推荐用于排查问题）
	fmt.Println("策略 1: 最简化配置")
	fmt.Println(strings.Repeat("-", 80))

	result1, err := g.GenerateWithValidation(qrCode, &models.DeepLinkOptions{
		PaymentType: models.PaymentTypeStandard,
	})

	if err != nil {
		fmt.Printf("❌ 生成失败: %v\n\n", err)
	} else {
		fmt.Printf("✅ 生成成功\n%s\n\n", result1.DeepLink)
	}

	// 策略 2: 完整配置
	fmt.Println("策略 2: 完整配置")
	fmt.Println(strings.Repeat("-", 80))

	result2, err := g.GenerateWithValidation(qrCode, &models.DeepLinkOptions{
		PaymentType: models.PaymentTypeDynamic,
		OrderID:     "TEST-" + fmt.Sprintf("%d", time.Now().Unix()),
		RedirectURL: "https://test.com/success",
		NotifyURL:   "https://test.com/webhook",
	})

	if err != nil {
		fmt.Printf("❌ 生成失败: %v\n\n", err)
	} else {
		fmt.Printf("✅ 生成成功\n%s\n\n", result2.DeepLink)
	}
}

func main() {
	debugger := NewDeepLinkDebugger()

	// 示例 1: 调试 Deep Link
	fmt.Println("示例 1: 调试现有 Deep Link")
	fmt.Println(strings.Repeat("=", 80))
	deepLink := "gcash://com.mynt.gcash/app/006300000800?qrCode=00020101021228790011ph.ppmi.p2m0111PAEYPHM2XXX0324VkHUE2Fz8Ee2YxnTVPX34TZs041003030028860503010520473995303608540520.005802PH5916NEXA ONLINE SHOP6013General Trias62430012ph.ppmi.qrph0306WDXYZ005062110000803***88440012ph.ppmi.qrph0124VkHUE2Fz8Ee2YxnTVPX34TZs63047458&merchantId=217020000119199251998&bizNo=None&orderAmount=20.00&merchantName=NEXA%20ONLINE%20SHOP&shopId=VkHUE2Fz8Ee2YxnTVPX34TZs&qrCodeFormat=EMVCO&tfrbnkcode=PAEYPHM2XXX&clientId=2023062916065505394208&param3=99960005%7Eph.ppmi.p2m%7E%7E%7E301&param5=VkHUE2Fz8Ee2YxnTVPX34TZs%7WDXYZ0%7E%7E%7E211000&tfrAcctNo=VkHUE2Fz8Ee2YxnTVPX34TZs&acqInfo=211000&sub=p2mpay&lucky=false"
	debugger.DebugDeepLink(deepLink)

	fmt.Println("\n" + strings.Repeat("=", 80) + "\n")

	// 示例 2: 调试 QR Code
	fmt.Println("示例 2: 调试 QR Code")
	fmt.Println(strings.Repeat("=", 80))
	qrCode := "00020101021228790011ph.ppmi.p2m0111PAEYPHM2XXX0324VkHUE2Fz8Ee2YxnTVPX34TZs041003030028860503010520473995303608540520.005802PH5916NEXA ONLINE SHOP6013General Trias62430012ph.ppmi.qrph0306WDXYZ005062110000803***88440012ph.ppmi.qrph0124VkHUE2Fz8Ee2YxnTVPX34TZs63047458"
	debugger.DebugQRCode(qrCode)
}
