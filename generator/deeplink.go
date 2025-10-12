package generator

import (
	"fmt"
	"net/url"
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
	deepLink := fmt.Sprintf("%s?%s", GCashBaseURL, values.Encode())

	return &models.DeepLinkResult{
		Success:     true,
		DeepLink:    deepLink,
		ParsedData:  data,
		Options:     options,
		GeneratedAt: time.Now(),
	}, nil
}

// GenerateWithValidation 生成并验证 Deep Link
func (g *DeepLinkGenerator) GenerateWithValidation(qrData string, options *models.DeepLinkOptions) (*models.DeepLinkResult, error) {
	// 1. 解析 QR Code
	p := parser.NewEMVCoParser()
	data, err := p.Parse(qrData)
	if err != nil {
		return g.errorResult(fmt.Sprintf("解析失败: %v", err))
	}

	// 2. 验证 QR Code
	validation := p.Validate(qrData)
	if !validation.Valid {
		return g.errorResult(fmt.Sprintf("验证失败: %v", validation.Errors))
	}

	// 3. 生成 Deep Link
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
		options.ClientID = fmt.Sprintf("%d", time.Now().UnixNano()/1e6)
	}

	// 支付类型
	if options.PaymentType == "" {
		if options.OrderID != "" {
			options.PaymentType = models.PaymentTypeDynamic // 有订单号，使用动态支付
		} else {
			options.PaymentType = models.PaymentTypeStandard // 标准支付
		}
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
	values.Add("orderAmount", options.OrderAmount)
	values.Add("merchantName", data.MerchantName)
	values.Add("qrCodeFormat", "EMVCO")
	values.Add("sub", "p2mpay")
	values.Add("lucky", fmt.Sprintf("%t", options.EnableLucky))
	values.Add("bizNo", options.BizNo)
	values.Add("clientId", options.ClientID)

	// 可选参数 - 只在有值时添加
	g.addIfNotEmpty(values, "merchantId", options.MerchantID)
	g.addIfNotEmpty(values, "orderId", options.OrderID)
	g.addIfNotEmpty(values, "tfrbnkcode", data.BankCode)
	g.addIfNotEmpty(values, "shopId", data.ShopID)
	g.addIfNotEmpty(values, "tfrAcctNo", data.ShopID)
	g.addIfNotEmpty(values, "acqInfo", data.AcqInfo)
	g.addIfNotEmpty(values, "merchantCity", data.MerchantCity)
	g.addIfNotEmpty(values, "merchantCategoryCode", data.MerchantCategoryCode)

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

// buildParam3 构建 param3 参数
func (g *DeepLinkGenerator) buildParam3(options *models.DeepLinkOptions) string {
	if options.CustomParam3 != "" {
		return options.CustomParam3
	}
	return fmt.Sprintf("99960005~ph.ppmi.p2m~~~%s", options.PaymentType)
}

// buildParam5 构建 param5 参数
func (g *DeepLinkGenerator) buildParam5(data *models.EMVCoData, options *models.DeepLinkOptions) string {
	if options.CustomParam5 != "" {
		return options.CustomParam5
	}

	if data.ShopID != "" && data.MerchantName != "" && data.AcqInfo != "" {
		return fmt.Sprintf("%s~%s~AAAAA   ~%s",
			data.ShopID,
			data.MerchantName,
			data.AcqInfo,
		)
	}

	return ""
}

// addIfNotEmpty 只在值非空时添加参数
func (g *DeepLinkGenerator) addIfNotEmpty(values url.Values, key, value string) {
	if value != "" && value != "null" {
		values.Add(key, value)
	}
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
