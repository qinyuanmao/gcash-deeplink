package models

import "time"

// EMVCoData EMVCo QR Code 解析后的数据
type EMVCoData struct {
	// 基础字段
	Version     string // Tag 00 - 版本号
	InitMethod  string // Tag 01 - 初始化方法
	Amount      string // Tag 54 - 交易金额
	Currency    string // Tag 53 - 货币代码 (608 = PHP)
	CountryCode string // Tag 58 - 国家代码 (PH)

	// 商户信息
	MerchantName         string // Tag 59 - 商户名称
	MerchantCity         string // Tag 60 - 商户城市
	MerchantCategoryCode string // Tag 52 - 商户分类码 (MCC)

	// 账户信息
	ShopID   string // 店铺 ID
	BankCode string // 银行代码

	// 附加数据
	AcqInfo        string // 获取方信息
	OrderReference string // 订单参考号
	CRC            string // Tag 63 - CRC 校验码

	// 原始数据
	RawData string // 原始 QR Code 数据
}

// PaymentType 支付类型
type PaymentType string

const (
	PaymentTypeStandard    PaymentType = "000" // 标准 P2M 支付
	PaymentTypeDynamic     PaymentType = "010" // 动态 QR 支付
	PaymentTypeStatic      PaymentType = "001" // 静态 QR 支付
	PaymentTypeInstallment PaymentType = "020" // 分期付款
	PaymentTypePreAuth     PaymentType = "030" // 预授权
)

// DeepLinkOptions GCash Deep Link 生成选项
type DeepLinkOptions struct {
	// 必需参数
	QRCode               string // EMVCo QR Code 数据
	OrderAmount          string // 订单金额
	AcqInfo              string // 获取方信息
	MerchantCategoryCode string // 商户分类码 (MCC)

	// 可选参数
	MerchantID   string      // 商户 ID (可选)
	MerchantName string      // 商户名称 (可选)
	OrderID      string      // 订单 ID
	PaymentType  PaymentType // 支付类型
	RedirectURL  string      // 支付完成后跳转 URL
	NotifyURL    string      // 服务器回调通知 URL
	ClientID     string      // 客户端 ID (自动生成)
	ShopID       string      // 店铺 ID

	// 高级选项
	EnableLucky  bool   // 是否启用抽奖
	BizNo        string // 业务单号
	CustomParam3 string // 自定义 param3
	CustomParam5 string // 自定义 param5
}

// DeepLinkResult Deep Link 生成结果
type DeepLinkResult struct {
	Success     bool             `json:"success"`
	DeepLink    string           `json:"deepLink,omitempty"`
	ParsedData  *EMVCoData       `json:"parsedData,omitempty"`
	Options     *DeepLinkOptions `json:"options,omitempty"`
	Error       string           `json:"error,omitempty"`
	GeneratedAt time.Time        `json:"generatedAt"`
}

// ValidationResult 验证结果
type ValidationResult struct {
	Valid  bool     `json:"valid"`
	Errors []string `json:"errors,omitempty"`
}
