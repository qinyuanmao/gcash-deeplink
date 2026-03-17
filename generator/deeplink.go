package generator

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/qinyuanmao/gcash-deeplink/models"
	"github.com/qinyuanmao/gcash-deeplink/parser"
)

const (
	// GCash Deep Link 基础 URL
	GCashBaseURL = "gcash://com.mynt.gcash/app/006300000800"
)

// DeepLinkGenerator GCash Deep Link 生成器
type DeepLinkGenerator struct{}

// NewDeepLinkGenerator 创建生成器实例
func NewDeepLinkGenerator() *DeepLinkGenerator {
	return &DeepLinkGenerator{}
}

// Generate 生成 GCash Deep Link
func (g *DeepLinkGenerator) Generate(data *models.EMVCoData, options *models.DeepLinkOptions) (*models.DeepLinkResult, error) {
	// 验证输入
	if data == nil {
		return g.errorResult("解析数据不能为空")
	}
	if options == nil {
		options = &models.DeepLinkOptions{}
	}

	// 填充默认值
	g.fillDefaults(data, options)

	// 构建参数
	values := g.buildParameters(data, options)

	// 生成 Deep Link
	// 使用 %20 替换 + 编码空格，确保 Android Uri.getQueryParameter() 正确解码
	query := strings.ReplaceAll(values.Encode(), "+", "%20")
	deepLink := fmt.Sprintf("%s?%s", GCashBaseURL, query)

	return &models.DeepLinkResult{
		Success:     true,
		DeepLink:    deepLink,
		ParsedData:  data,
		Options:     options,
		GeneratedAt: time.Now(),
	}, nil
}

// GenerateWithValidation 解析 QR Code 并生成 Deep Link
// Parse() 内部优先使用严格解析（含 CRC），失败时自动回退宽松解析
// GCash 后端会自行校验 QR 码，因此此处不再额外调用 Validate() 拦截
func (g *DeepLinkGenerator) GenerateWithValidation(qrData string, options *models.DeepLinkOptions) (*models.DeepLinkResult, error) {
	p := parser.NewEMVCoParser()
	data, err := p.Parse(qrData)
	if err != nil {
		return g.errorResult(fmt.Sprintf("解析失败: %v", err))
	}
	return g.Generate(data, options)
}

// fillDefaults 填充默认值
func (g *DeepLinkGenerator) fillDefaults(data *models.EMVCoData, options *models.DeepLinkOptions) {
	// QR Code 数据
	if options.QRCode == "" {
		options.QRCode = data.RawData
	}

	// 订单金额
	if options.OrderAmount == "" {
		options.OrderAmount = data.Amount
	}

	// 客户端 ID
	if options.ClientID == "" {
		options.ClientID = "2023062916065505394208" // 默认的客户端 ID
	}

	// 支付类型
	if options.PaymentType == "" {
		if options.OrderID != "" {
			options.PaymentType = models.PaymentTypeDynamic // 有订单号，使用动态支付
		} else {
			options.PaymentType = models.PaymentTypeStandard // 标准支付
		}
	}

	// 默认 merchantId
	if options.MerchantID == "" {
		options.MerchantID = "217020000119199251998"
	}

	// 新版 QR 格式: 28-03=UID, 62-05=订单号
	// 交换 shopId 和 acqInfo，使 shopId=订单号, acqInfo=UID
	// 仅在两个值都非空时才交换，防止 ShopID 被清空导致 param5 丢失
	if options.NewQRFormat && data.AcqInfo != "" && data.ShopID != "" {
		data.ShopID, data.AcqInfo = data.AcqInfo, data.ShopID
	}

	if options.ShopID == "" {
		options.ShopID = data.ShopID
	}

	if options.MerchantName == "" {
		options.MerchantName = data.MerchantName
	}

	// 业务单号
	if options.BizNo == "" {
		options.BizNo = "null"
	}
}

// buildParameters 构建 URL 参数
func (g *DeepLinkGenerator) buildParameters(data *models.EMVCoData, options *models.DeepLinkOptions) url.Values {
	values := url.Values{}

	// 必需参数
	values.Add("qrCode", options.QRCode)
	values.Add("bizNo", options.BizNo)
	values.Add("orderAmount", options.OrderAmount)
	values.Add("qrCodeFormat", "EMVCO")
	values.Add("sub", "p2mpay")
	values.Add("clientId", options.ClientID)
	values.Add("merchantName", options.MerchantName)

	// 可选参数 - 只在有值时添加
	g.addIfNotEmpty(values, "merchantId", options.MerchantID)
	g.addIfNotEmpty(values, "orderId", options.OrderID)
	g.addIfNotEmpty(values, "tfrbnkcode", data.BankCode)
	g.addIfNotEmpty(values, "shopId", options.ShopID)
	g.addIfNotEmpty(values, "tfrAcctNo", options.ShopID)
	g.addIfNotEmpty(values, "acqInfo", data.AcqInfo)

	// 回调 URL
	g.addIfNotEmpty(values, "redirectUrl", options.RedirectURL)
	g.addIfNotEmpty(values, "returnUrl", options.RedirectURL)
	g.addIfNotEmpty(values, "notifyUrl", options.NotifyURL)
	g.addIfNotEmpty(values, "callbackUrl", options.NotifyURL)

	// param3 和 param5
	param3 := g.buildParam3(options)
	param5 := g.buildParam5(data, options)
	g.addIfNotEmpty(values, "param3", param3)
	g.addIfNotEmpty(values, "param5", param5)

	return values
}

// addIfNotEmpty 只在值非空时添加参数
func (g *DeepLinkGenerator) addIfNotEmpty(values url.Values, key, value string) {
	if value != "" && value != "null" {
		values.Add(key, value)
	}
}

// buildParam3 构建 param3 参数
func (g *DeepLinkGenerator) buildParam3(options *models.DeepLinkOptions) string {
	return fmt.Sprintf("99960005~ph.ppmi.p2m~~~%s", options.PaymentType)
}

// buildParam5 构建 param5 参数
// 格式: ShopID~MerchantName~TerminalLabel~AcqInfo (4段3波浪，对齐 Luca 模板)
func (g *DeepLinkGenerator) buildParam5(data *models.EMVCoData, options *models.DeepLinkOptions) string {
	if options.ShopID == "" {
		return ""
	}
	return fmt.Sprintf("%s~%s~%s~%s", options.ShopID, options.MerchantName, data.TerminalLabel, data.AcqInfo)
}

// errorResult 创建错误结果
func (g *DeepLinkGenerator) errorResult(errMsg string) (*models.DeepLinkResult, error) {
	return &models.DeepLinkResult{
		Success:     false,
		Error:       errMsg,
		GeneratedAt: time.Now(),
	}, fmt.Errorf(errMsg)
}

// GenerateMultiple 生成多种策略的 Deep Link
func (g *DeepLinkGenerator) GenerateMultiple(data *models.EMVCoData) map[string]string {
	strategies := make(map[string]string)

	// 策略 1: 最简化
	result1, _ := g.Generate(data, &models.DeepLinkOptions{
		PaymentType: models.PaymentTypeStandard,
	})
	if result1.Success {
		strategies["minimal"] = result1.DeepLink
	}

	// 策略 2: 动态支付
	result2, _ := g.Generate(data, &models.DeepLinkOptions{
		PaymentType: models.PaymentTypeDynamic,
		OrderID:     fmt.Sprintf("ORDER-%d", time.Now().Unix()),
	})
	if result2.Success {
		strategies["dynamic"] = result2.DeepLink
	}

	// 策略 3: 带回调
	result3, _ := g.Generate(data, &models.DeepLinkOptions{
		PaymentType: models.PaymentTypeDynamic,
		OrderID:     fmt.Sprintf("ORDER-%d", time.Now().Unix()),
		RedirectURL: "https://yourdomain.com/payment/success",
		NotifyURL:   "https://yourdomain.com/api/gcash/notify",
	})
	if result3.Success {
		strategies["with_callback"] = result3.DeepLink
	}

	return strategies
}
