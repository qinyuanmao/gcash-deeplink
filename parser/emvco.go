package parser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	
	"github.com/qinyuanmao/gcash-deeplink/models"
)

// EMVCoParser EMVCo QR Code 解析器
type EMVCoParser struct{}

// NewEMVCoParser 创建解析器实例
func NewEMVCoParser() *EMVCoParser {
	return &EMVCoParser{}
}

// Parse 解析 EMVCo QR Code
func (p *EMVCoParser) Parse(qrData string) (*models.EMVCoData, error) {
	if qrData == "" {
		return nil, fmt.Errorf("QR Code 数据不能为空")
	}

	data := &models.EMVCoData{
		RawData: qrData,
	}

	// 解析版本 (Tag 00)
	if match := regexp.MustCompile(`^00(\d{2})(\d+)`).FindStringSubmatch(qrData); len(match) > 2 {
		length, _ := strconv.Atoi(match[1])
		if len(match[2]) >= length {
			data.Version = match[2][:length]
		}
	}

	// 解析初始化方法 (Tag 01)
	if match := regexp.MustCompile(`01(\d{2})(\d+)`).FindStringSubmatch(qrData); len(match) > 2 {
		length, _ := strconv.Atoi(match[1])
		if len(match[2]) >= length {
			data.InitMethod = match[2][:length]
		}
	}

	// 解析商户分类代码 (Tag 52)
	if match := regexp.MustCompile(`5204(\d{4})`).FindStringSubmatch(qrData); len(match) > 1 {
		data.MerchantCategoryCode = match[1]
	}

	// 解析货币代码 (Tag 53)
	if match := regexp.MustCompile(`5303(\d{3})`).FindStringSubmatch(qrData); len(match) > 1 {
		data.Currency = match[1]
	}

	// 解析金额 (Tag 54)
	if match := regexp.MustCompile(`5406([\d.]+)`).FindStringSubmatch(qrData); len(match) > 1 {
		data.Amount = match[1]
	}

	// 解析国家代码 (Tag 58)
	if match := regexp.MustCompile(`5802([A-Z]{2})`).FindStringSubmatch(qrData); len(match) > 1 {
		data.CountryCode = match[1]
	}

	// 解析商户名称 (Tag 59)
	data.MerchantName = p.extractVariableLengthField(qrData, "59")

	// 解析商户城市 (Tag 60)
	data.MerchantCity = p.extractVariableLengthField(qrData, "60")

	// 解析商户账户信息 (Tag 26 或 27)
	p.parseMerchantAccountInfo(qrData, data)

	// 解析附加数据 (Tag 62)
	p.parseAdditionalData(qrData, data)

	// 解析 CRC (Tag 63)
	if match := regexp.MustCompile(`6304([A-F0-9]{4})$`).FindStringSubmatch(qrData); len(match) > 1 {
		data.CRC = match[1]
	}

	return data, nil
}

// extractVariableLengthField 提取可变长度字段
func (p *EMVCoParser) extractVariableLengthField(qrData, tag string) string {
	// 匹配格式: TAG + LENGTH(2位) + VALUE
	pattern := fmt.Sprintf(`%s(\d{2})(.+?)(?:\d{2}\d{2}|$)`, tag)
	if match := regexp.MustCompile(pattern).FindStringSubmatch(qrData); len(match) > 2 {
		length, _ := strconv.Atoi(match[1])
		if len(match[2]) >= length {
			return strings.TrimSpace(match[2][:length])
		}
	}
	return ""
}

// parseMerchantAccountInfo 解析商户账户信息
func (p *EMVCoParser) parseMerchantAccountInfo(qrData string, data *models.EMVCoData) {
	// Tag 26 或 27 - 商户账户信息模板
	for _, tag := range []string{"26", "27", "28"} {
		pattern := fmt.Sprintf(`%s(\d{2})(.+?)(?:(?:2[6-9]|[3-9]\d)\d{2}|$)`, tag)
		if match := regexp.MustCompile(pattern).FindStringSubmatch(qrData); len(match) > 2 {
			length, _ := strconv.Atoi(match[1])
			if len(match[2]) >= length {
				merchantInfo := match[2][:length]
				
				// 提取银行代码 (子标签 00)
				if subMatch := regexp.MustCompile(`0011([A-Z0-9]+)`).FindStringSubmatch(merchantInfo); len(subMatch) > 1 {
					data.BankCode = subMatch[1]
				}
				
				// 提取 ShopID (子标签 01, 03, 或 04)
				for _, subTag := range []string{"01", "03", "04"} {
					if subMatch := regexp.MustCompile(fmt.Sprintf(`%s(\d{2})([A-Z0-9\-]+)`, subTag)).FindStringSubmatch(merchantInfo); len(subMatch) > 2 {
						subLength, _ := strconv.Atoi(subMatch[1])
						if len(subMatch[2]) >= subLength {
							data.ShopID = subMatch[2][:subLength]
							break
						}
					}
				}
				
				break
			}
		}
	}
}

// parseAdditionalData 解析附加数据
func (p *EMVCoParser) parseAdditionalData(qrData string, data *models.EMVCoData) {
	// Tag 62 - 附加数据模板
	pattern := `62(\d{2})(.+?)(?:63\d{2}|$)`
	if match := regexp.MustCompile(pattern).FindStringSubmatch(qrData); len(match) > 2 {
		length, _ := strconv.Atoi(match[1])
		if len(match[2]) >= length {
			additionalData := match[2][:length]
			
			// 子标签 01 - 账单号/订单参考号
			if subMatch := regexp.MustCompile(`01(\d{2})(.+)`).FindStringSubmatch(additionalData); len(subMatch) > 2 {
				subLength, _ := strconv.Atoi(subMatch[1])
				if len(subMatch[2]) >= subLength {
					data.OrderReference = subMatch[2][:subLength]
				}
			}

			// 子标签 03 - 获取方信息 (某些 QR Code 中 AcqInfo 在这里)
			var acqInfo03 string
			if subMatch := regexp.MustCompile(`03(\d{2})(.+)`).FindStringSubmatch(additionalData); len(subMatch) > 2 {
				subLength, _ := strconv.Atoi(subMatch[1])
				if len(subMatch[2]) >= subLength {
					acqInfo03 = subMatch[2][:subLength]
				}
			}

			// 子标签 05 - 获取方信息 (某些 QR Code 中 AcqInfo 在这里)
			var acqInfo05 string
			if subMatch := regexp.MustCompile(`05(\d{2})(.+)`).FindStringSubmatch(additionalData); len(subMatch) > 2 {
				subLength, _ := strconv.Atoi(subMatch[1])
				if len(subMatch[2]) >= subLength {
					acqInfo05 = subMatch[2][:subLength]
				}
			}

			// 选择合适的 AcqInfo
			// 策略：优先使用包含非数字字符的值
			// 1. 如果 acqInfo05 包含非数字字符，使用 acqInfo05
			// 2. 否则，如果 acqInfo03 包含非数字字符，使用 acqInfo03
			// 3. 否则，使用 acqInfo05（如果有）
			// 4. 最后使用 acqInfo03
			isDigitOnly03 := acqInfo03 != "" && regexp.MustCompile(`^\d+$`).MatchString(acqInfo03)
			isDigitOnly05 := acqInfo05 != "" && regexp.MustCompile(`^\d+$`).MatchString(acqInfo05)

			if acqInfo05 != "" && !isDigitOnly05 {
				// acqInfo05 有非数字字符，优先使用
				data.AcqInfo = acqInfo05
			} else if acqInfo03 != "" && !isDigitOnly03 {
				// acqInfo03 有非数字字符，使用它
				data.AcqInfo = acqInfo03
			} else if acqInfo05 != "" {
				// 两者都是纯数字或为空，优先使用 acqInfo05
				data.AcqInfo = acqInfo05
			} else {
				// 最后使用 acqInfo03
				data.AcqInfo = acqInfo03
			}
		}
	}
}

// Validate 验证 EMVCo QR Code
func (p *EMVCoParser) Validate(qrData string) *models.ValidationResult {
	result := &models.ValidationResult{
		Valid:  true,
		Errors: []string{},
	}

	// 1. 检查是否为空
	if qrData == "" {
		result.Valid = false
		result.Errors = append(result.Errors, "QR Code 数据不能为空")
		return result
	}

	// 2. 检查最小长度
	if len(qrData) < 50 {
		result.Valid = false
		result.Errors = append(result.Errors, "QR Code 数据长度过短")
	}

	// 3. 检查版本号 (Tag 00)
	if !regexp.MustCompile(`^0002`).MatchString(qrData) {
		result.Valid = false
		result.Errors = append(result.Errors, "QR Code 应以 0002 开头 (版本号)")
	}

	// 4. 检查必需的标签
	requiredTags := map[string]string{
		"52": "商户分类码 (MCC)",
		"53": "货币代码",
		"58": "国家代码",
		"59": "商户名称",
		"60": "商户城市",
	}

	for tag, name := range requiredTags {
		pattern := fmt.Sprintf(`%s\d{2}`, tag)
		if !regexp.MustCompile(pattern).MatchString(qrData) {
			result.Errors = append(result.Errors, fmt.Sprintf("缺少必需字段 Tag %s: %s", tag, name))
			result.Valid = false
		}
	}

	// 5. 检查 CRC (Tag 63)
	if !regexp.MustCompile(`6304[A-F0-9]{4}$`).MatchString(qrData) {
		result.Errors = append(result.Errors, "CRC 校验码格式不正确或缺失")
		result.Valid = false
	}

	return result
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
		data.OrderReference,
	)
}
