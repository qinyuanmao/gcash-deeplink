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

	// 构建参数 (按固定顺序)
	queryString := g.buildQueryString(data, options)

	// 生成 Deep Link
	deepLink := fmt.Sprintf("%s?%s", GCashBaseURL, queryString)

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

// buildQueryString 按固定顺序构建 URL 参数
func (g *DeepLinkGenerator) buildQueryString(data *models.EMVCoData, options *models.DeepLinkOptions) string {
	var params []string

	// 按照指定顺序添加参数:
	// qrCode, merchantId, bizNo, orderAmount, merchantName, shopId,
	// qrCodeFormat, tfrbnkcode, clientId, param3, param5, tfrAcctNo, acqInfo, sub, lucky

	// 1. qrCode (必需)
	params = append(params, "qrCode="+url.QueryEscape(options.QRCode))

	// 2. merchantId (可选)
	if options.MerchantID != "" {
		params = append(params, "merchantId="+url.QueryEscape(options.MerchantID))
	}

	// 3. bizNo (必需)
	params = append(params, "bizNo="+url.QueryEscape(options.BizNo))

	// 4. orderAmount (必需)
	params = append(params, "orderAmount="+url.QueryEscape(options.OrderAmount))

	// 5. merchantName (必需)
	params = append(params, "merchantName="+url.QueryEscape(data.MerchantName))

	// 6. shopId (可选)
	if options.ShopID != "" {
		params = append(params, "shopId="+url.QueryEscape(options.ShopID))
	}

	// 7. qrCodeFormat (必需)
	params = append(params, "qrCodeFormat=EMVCO")

	// 8. tfrbnkcode (可选)
	if data.BankCode != "" {
		params = append(params, "tfrbnkcode="+url.QueryEscape(data.BankCode))
	}

	// 9. clientId (必需)
	params = append(params, "clientId="+url.QueryEscape(options.ClientID))

	// 10. param3 (必需)
	param3 := g.buildParam3(options)
	params = append(params, "param3="+url.QueryEscape(param3))

	// 11. param5 (可选)
	param5 := g.buildParam5(data, options)
	if param5 != "" {
		params = append(params, "param5="+url.QueryEscape(param5))
	}

	// 12. tfrAcctNo (可选)
	if options.ShopID != "" {
		params = append(params, "tfrAcctNo="+url.QueryEscape(options.ShopID))
	}

	// 13. acqInfo (可选)
	if data.AcqInfo != "" {
		params = append(params, "acqInfo="+url.QueryEscape(data.AcqInfo))
	}

	// 14. sub (必需)
	params = append(params, "sub=p2mpay")

	// 15. lucky (可选,仅在显式设置时添加)
	if options.EnableLucky != nil {
		params = append(params, fmt.Sprintf("lucky=%t", *options.EnableLucky))
	}

	// 其他可选参数 (不在指定顺序中,追加到最后)
	if options.MerchantCity != "" {
		params = append(params, "merchantCity="+url.QueryEscape(options.MerchantCity))
	}
	if options.MerchantCategoryCode != "" {
		params = append(params, "merchantCategoryCode="+url.QueryEscape(options.MerchantCategoryCode))
	}
	if options.RedirectURL != "" {
		params = append(params, "redirectUrl="+url.QueryEscape(options.RedirectURL))
		params = append(params, "returnUrl="+url.QueryEscape(options.RedirectURL))
	}
	if options.NotifyURL != "" {
		params = append(params, "notifyUrl="+url.QueryEscape(options.NotifyURL))
		params = append(params, "callbackUrl="+url.QueryEscape(options.NotifyURL))
	}
	if options.OrderID != "" {
		params = append(params, "orderId="+url.QueryEscape(options.OrderID))
	}

	return strings.Join(params, "&")
}

// buildParam3 构建 param3 参数
func (g *DeepLinkGenerator) buildParam3(options *models.DeepLinkOptions) string {
	return fmt.Sprintf("99960005~ph.ppmi.p2m~~~%s", options.PaymentType)
}

// buildParam5 构建 param5 参数
func (g *DeepLinkGenerator) buildParam5(data *models.EMVCoData, options *models.DeepLinkOptions) string {
	// param5 格式：ShopID~OrderID~~~AcqInfo
	// 只要有 ShopID 就生成 param5
	if options.ShopID != "" {
		// 有 AcqInfo，使用格式：ShopID~OrderID~~~AcqInfo
		return fmt.Sprintf("%s~%s~~~%s",
			options.ShopID,
			data.OrderID,
			data.AcqInfo,
		)
	}

	return ""
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
