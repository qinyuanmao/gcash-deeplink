package parser

import (
	"fmt"
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

// Parse 解析 EMVCo QR Code（使用 mercari mpm.Decode）
func (p *EMVCoParser) Parse(qrData string) (*models.EMVCoData, error) {
	if qrData == "" {
		return nil, fmt.Errorf("QR Code 数据不能为空")
	}

	code, err := mpm.Decode([]byte(qrData))
	if err != nil {
		return nil, fmt.Errorf("EMVCo 解析失败: %w", err)
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
