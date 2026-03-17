package parser

import (
	"fmt"
	"strconv"
	"strings"

	"go.mercari.io/go-emv-code/mpm"
	"go.mercari.io/go-emv-code/tlv"

	"github.com/qinyuanmao/gcash-deeplink/models"
)

// EMVCoParser EMVCo QR Code 解析器
type EMVCoParser struct{}

// NewEMVCoParser 创建解析器实例
func NewEMVCoParser() *EMVCoParser {
	return &EMVCoParser{}
}

// merchantAccountSub Tag 26-51 子标签结构
type merchantAccountSub struct {
	GlobalUID string `emv:"00"`
	BankCode  string `emv:"01"`
	ShopID    string `emv:"03"`
}

// additionalDataSub Tag 62 子标签结构
type additionalDataSub struct {
	OrderID       string `emv:"03"`
	AcqInfo       string `emv:"05"`
	TerminalLabel string `emv:"07"`
}

// Parse 解析 EMVCo QR Code
// 优先使用 mercari mpm.Decode（含 CRC 校验），失败时回退到宽松 TLV 解析（跳过 CRC）
// GCash 后端会自行校验 QR 码，CRC 错误不应阻断 deeplink 生成
func (p *EMVCoParser) Parse(qrData string) (*models.EMVCoData, error) {
	if qrData == "" {
		return nil, fmt.Errorf("QR Code 数据不能为空")
	}

	code, err := mpm.Decode([]byte(qrData))
	if err != nil {
		// 严格模式失败（CRC 错误等），回退到宽松 TLV 解析
		return parseFallback(qrData)
	}

	data := &models.EMVCoData{
		RawData:              qrData,
		Version:              code.PayloadFormatIndicator,
		InitMethod:           string(code.PointOfInitiationMethod),
		MerchantCategoryCode: strings.TrimSpace(code.MerchantCategoryCode),
		Currency:             strings.TrimSpace(code.TransactionCurrency),
		CountryCode:          strings.TrimSpace(code.CountryCode),
		MerchantName:         strings.TrimSpace(code.MerchantName),
		MerchantCity:         strings.TrimSpace(code.MerchantCity),
	}

	// Amount (NullString)
	if code.TransactionAmount.Valid {
		data.Amount = code.TransactionAmount.String
	}

	// Tag 02-51 Merchant Account Info — 解析子标签
	parseMerchantSubTags(code, data)

	// Tag 62 Additional Data — 解析子标签
	parseAdditionalSubTags(code.AdditionalDataFieldTemplate, data)

	return data, nil
}

// parseFallback 宽松 TLV 解析 — 跳过 CRC 校验，直接提取字段
func parseFallback(qrData string) (*models.EMVCoData, error) {
	data := &models.EMVCoData{RawData: qrData}
	i := 0
	for i+4 <= len(qrData) {
		tag := qrData[i : i+2]
		length, err := strconv.Atoi(qrData[i+2 : i+4])
		if err != nil || length < 0 {
			return nil, fmt.Errorf("TLV 解析失败: 位置 %d 长度无效", i+2)
		}
		if i+4+length > len(qrData) {
			break
		}
		value := qrData[i+4 : i+4+length]

		switch tag {
		case "00":
			data.Version = value
		case "01":
			data.InitMethod = value
		case "52":
			data.MerchantCategoryCode = strings.TrimSpace(value)
		case "53":
			data.Currency = strings.TrimSpace(value)
		case "54":
			data.Amount = value
		case "58":
			data.CountryCode = strings.TrimSpace(value)
		case "59":
			data.MerchantName = strings.TrimSpace(value)
		case "60":
			data.MerchantCity = strings.TrimSpace(value)
		case "62":
			parseAdditionalSubTags(value, data)
		case "63":
			data.CRC = value
		default:
			// Tag 02-51: Merchant Account Information
			tagNum, _ := strconv.Atoi(tag)
			if tagNum >= 2 && tagNum <= 51 {
				parseMerchantSubTagsFallback(value, data)
			}
		}
		i += 4 + length
	}
	return data, nil
}

// parseMerchantSubTagsFallback 从原始 TLV value 中解析 merchant 子标签
func parseMerchantSubTagsFallback(value string, data *models.EMVCoData) {
	if data.BankCode != "" {
		return // 已找到 ph.ppmi.p2m 的 merchant account
	}
	var sub merchantAccountSub
	_ = tlv.NewDecoder(strings.NewReader(value), "emv", 512, 2, 2, nil).Decode(&sub)
	if strings.Contains(sub.GlobalUID, "ph.ppmi.p2m") {
		data.BankCode = sub.BankCode
		data.ShopID = sub.ShopID
	}
}

// Validate 验证 EMVCo QR Code（mpm.Decode 自带 CRC 校验和格式验证）
func (p *EMVCoParser) Validate(qrData string) *models.ValidationResult {
	if qrData == "" {
		return &models.ValidationResult{
			Valid:  false,
			Errors: []string{"QR Code 数据不能为空"},
		}
	}

	_, err := mpm.Decode([]byte(qrData))
	if err != nil {
		return &models.ValidationResult{
			Valid:  false,
			Errors: []string{err.Error()},
		}
	}
	return &models.ValidationResult{Valid: true}
}

// parseMerchantSubTags 从 MerchantAccountInformation 中解析子标签
func parseMerchantSubTags(code *mpm.Code, data *models.EMVCoData) {
	for _, t := range code.MerchantAccountInformation {
		var sub merchantAccountSub
		_ = tlv.NewDecoder(strings.NewReader(t.Value), "emv", 512, 2, 2, nil).Decode(&sub)
		// 只取包含 ph.ppmi.p2m 的 merchant account
		if strings.Contains(sub.GlobalUID, "ph.ppmi.p2m") {
			data.BankCode = sub.BankCode
			data.ShopID = sub.ShopID
			return
		}
	}
}

// parseAdditionalSubTags 从 AdditionalDataFieldTemplate 中解析子标签
func parseAdditionalSubTags(template string, data *models.EMVCoData) {
	if template == "" {
		return
	}
	var sub additionalDataSub
	_ = tlv.NewDecoder(strings.NewReader(template), "emv", 512, 2, 2, nil).Decode(&sub)
	data.OrderID = sub.OrderID
	data.AcqInfo = sub.AcqInfo
	data.TerminalLabel = sub.TerminalLabel
}

// GetSummary 获取 QR Code 摘要信息
func (p *EMVCoParser) GetSummary(data *models.EMVCoData) string {
	return fmt.Sprintf(`EMVCo QR Code 信息:
商户: %s
城市: %s
金额: ₱%s
店铺ID: %s
银行代码: %s
商户分类: %s
订单号: %s`,
		data.MerchantName,
		data.MerchantCity,
		data.Amount,
		data.ShopID,
		data.BankCode,
		data.MerchantCategoryCode,
		data.OrderID,
	)
}
