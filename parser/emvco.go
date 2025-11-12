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
	data.MerchantCategoryCode = p.extractVariableLengthField(qrData, "52")

	// 解析货币代码 (Tag 53)
	data.Currency = p.extractVariableLengthField(qrData, "53")

	// 解析金额 (Tag 54)
	data.Amount = p.extractVariableLengthField(qrData, "54")

	// 解析国家代码 (Tag 58)
	data.CountryCode = p.extractVariableLengthField(qrData, "58")

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
	// 按照 EMVCo TLV (Tag-Length-Value) 结构顺序解析
	// 这样可以避免匹配到值内部的数字模式
	i := 0
	for i <= len(qrData)-4 {
		// 读取当前位置的 tag (2字节)
		currentTag := qrData[i : i+2]

		// 读取长度字段 (2字节)
		lengthStr := qrData[i+2 : i+4]
		if !isDigit(lengthStr[0]) || !isDigit(lengthStr[1]) {
			i++
			continue
		}

		length, err := strconv.Atoi(lengthStr)
		if err != nil {
			i++
			continue
		}

		// 检查是否有足够的数据
		if i+4+length > len(qrData) {
			i++
			continue
		}

		// 如果是我们要找的 tag,返回值
		if currentTag == tag {
			value := qrData[i+4 : i+4+length]
			return strings.TrimSpace(value)
		}

		// 跳过当前 TLV (Tag + Length + Value)
		i += 4 + length
	}
	return ""
}

// parseMerchantAccountInfo 解析商户账户信息
func (p *EMVCoParser) parseMerchantAccountInfo(qrData string, data *models.EMVCoData) {
	// Tag 26, 27 或 28 - 商户账户信息模板
	// 使用逐字符解析避免正则表达式匹配错误
	i := 0
	for i < len(qrData)-4 {
		tag := qrData[i : i+2]
		if tag == "26" || tag == "27" || tag == "28" {
			// 验证这是一个 Tag（前面应该是另一个 Tag 的结束或开始）
			lengthStr := qrData[i+2 : i+4]
			if len(lengthStr) == 2 && isDigit(lengthStr[0]) && isDigit(lengthStr[1]) {
				length, _ := strconv.Atoi(lengthStr)
				if i+4+length <= len(qrData) {
					merchantInfo := qrData[i+4 : i+4+length]

					// 提取银行代码 (tfrbnkcode) - 子标签 01
					subIdx := 0
					for subIdx < len(merchantInfo)-4 {
						subTag := merchantInfo[subIdx : subIdx+2]
						subLengthStr := merchantInfo[subIdx+2 : subIdx+4]
						if len(subLengthStr) == 2 && isDigit(subLengthStr[0]) && isDigit(subLengthStr[1]) {
							subLength, _ := strconv.Atoi(subLengthStr)
							if subIdx+4+subLength <= len(merchantInfo) {
								subValue := merchantInfo[subIdx+4 : subIdx+4+subLength]

								if subTag == "01" {
									data.BankCode = subValue
								} else if subTag == "03" {
									data.ShopID = subValue
								}

								subIdx += 4 + subLength
							} else {
								break
							}
						} else {
							break
						}
					}

					return // 找到并处理完 Tag 26/27/28，退出
				}
			}
		}
		i++
	}
}

// isDigit 检查字符是否为数字
func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
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

			// 子标签 03 - Store Label (商店标签/获取方信息)
			if subMatch := regexp.MustCompile(`03(\d{2})(.+)`).FindStringSubmatch(additionalData); len(subMatch) > 2 {
				subLength, _ := strconv.Atoi(subMatch[1])
				if len(subMatch[2]) >= subLength {
					data.AcqInfo03 = subMatch[2][:subLength]
				}
			}

			// 子标签 05 - Reference Label (参考标签/获取方信息)
			if subMatch := regexp.MustCompile(`05(\d{2})(.+)`).FindStringSubmatch(additionalData); len(subMatch) > 2 {
				subLength, _ := strconv.Atoi(subMatch[1])
				if len(subMatch[2]) >= subLength {
					data.AcqInfo05 = subMatch[2][:subLength]
				}
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
