# GCash Deep Link Generator

将 EMVCo QR Code 解析并生成 GCash Deep Link 的 Golang 工具。

## 功能特性

- ✅ 完整的 EMVCo QR Code 解析
- ✅ GCash Deep Link 生成
- ✅ 多种支付策略支持
- ✅ HTTP API 接口
- ✅ 输入验证
- ✅ 完整的错误处理
- ✅ 单元测试和基准测试

## 快速开始

### 安装

```bash
# 克隆或下载项目
cd gcash-deeplink

# 安装依赖（本项目仅使用标准库，无需额外依赖）
go mod tidy
```

### 运行

```bash
# 启动 HTTP API 服务器
go run main.go

# 运行示例
go run main.go examples

# 运行测试
go test -v

# 运行基准测试
go test -bench=.
```

## 使用示例

### 1. 作为 Go 包使用

```go
package main

import (
    "fmt"
    "gcash-deeplink/generator"
    "gcash-deeplink/models"
    "gcash-deeplink/parser"
)

func main() {
    // EMVCo QR Code 数据
    qrCode := "00020101021228530011ph.ppmi.p2m..."

    // 解析 QR Code
    p := parser.NewEMVCoParser()
    data, err := p.Parse(qrCode)
    if err != nil {
        panic(err)
    }

    // 生成 Deep Link
    g := generator.NewDeepLinkGenerator()
    options := &models.DeepLinkOptions{
        PaymentType: models.PaymentTypeDynamic,
        OrderID:     "ORDER-12345",
        RedirectURL: "https://myshop.com/success",
        NotifyURL:   "https://myshop.com/webhook",
    }

    result, err := g.Generate(data, options)
    if err != nil {
        panic(err)
    }

    fmt.Println("Deep Link:", result.DeepLink)
}
```

### 2. HTTP API 使用

启动服务器：

```bash
go run main.go
```

#### API 端点

**POST /api/parse** - 解析 QR Code

```bash
curl -X POST http://localhost:9000/api/parse \
  -H "Content-Type: application/json" \
  -d '{
    "qrCode": "00020101021228530011ph.ppmi.p2m..."
  }'
```

**POST /api/generate** - 生成 Deep Link

```bash
curl -X POST http://localhost:9000/api/generate \
  -H "Content-Type: application/json" \
  -d '{
    "qrCode": "00020101021228530011ph.ppmi.p2m...",
    "orderId": "ORDER-12345",
    "redirectUrl": "https://myshop.com/success",
    "notifyUrl": "https://myshop.com/webhook",
    "paymentType": "010"
  }'
```

**POST /api/validate** - 验证 QR Code

```bash
curl -X POST http://localhost:9000/api/validate \
  -H "Content-Type: application/json" \
  -d '{
    "qrCode": "00020101021228530011ph.ppmi.p2m..."
  }'
```

**GET /health** - 健康检查

```bash
curl http://localhost:9000/health
```

## 支付类型

| 类型代码 | 说明         | 使用场景                |
| -------- | ------------ | ----------------------- |
| `000`    | 标准支付     | 普通商户收款            |
| `010`    | 动态 QR 支付 | 订单支付，每次生成新 QR |
| `001`    | 静态 QR 支付 | 店铺固定收款码          |
| `020`    | 分期付款     | 支持分期的支付          |
| `030`    | 预授权       | 预授权后扣款            |

## 项目结构

```
gcash-deeplink/
├── go.mod              # Go 模块文件
├── main.go             # 主程序和 HTTP API
├── main_test.go        # 测试文件
├── models/             # 数据模型
│   └── types.go
├── parser/             # EMVCo QR Code 解析器
│   └── emvco.go
└── generator/          # GCash Deep Link 生成器
    └── deeplink.go
```

## API 响应示例

### 成功响应

```json
{
  "success": true,
  "deepLink": "gcash://com.mynt.gcash/app/006300000800?qrCode=...",
  "parsedData": {
    "Amount": "100.00",
    "MerchantName": "SOCMED DIGITAL MARKETING",
    "MerchantCity": "MakatiCity",
    "ShopID": "MRCHNT-4H3TZ",
    "BankCode": "SRCPPHM2XXX",
    "MerchantCategoryCode": "5199"
  },
  "generatedAt": "2025-01-10T10:30:00Z"
}
```

### 错误响应

```json
{
  "success": false,
  "error": "QR Code 数据不能为空",
  "generatedAt": "2025-01-10T10:30:00Z"
}
```

## 前端集成示例

参考项目中的 `example.html` 文件，提供了完整的前端调用示例。

## 配置选项

### DeepLinkOptions

```go
type DeepLinkOptions struct {
    QRCode      string      // EMVCo QR Code 数据
    OrderAmount string      // 订单金额
    MerchantID  string      // 商户 ID (可选)
    OrderID     string      // 订单 ID
    PaymentType PaymentType // 支付类型
    RedirectURL string      // 支付完成后跳转 URL
    NotifyURL   string      // 服务器回调通知 URL
    ClientID    string      // 客户端 ID (自动生成)
    EnableLucky bool        // 是否启用抽奖
    BizNo       string      // 业务单号
}
```

## 测试

```bash
# 运行所有测试
go test -v

# 运行特定测试
go test -v -run TestParseEMVCoQR

# 运行基准测试
go test -bench=. -benchmem

# 测试覆盖率
go test -cover
```

## 性能

基准测试结果（参考）：

```
BenchmarkParseQRCode-8       50000    30000 ns/op
BenchmarkGenerateDeepLink-8  30000    45000 ns/op
```

## 注意事项

1. **merchantId 是可选的** - 如果您的系统中没有 merchantId，可以不传递此参数
2. **param3 格式** - 最后的 3 位数字（000/010/001）表示支付类型
3. **URL 编码** - 使用 `url.Values` 自动处理 URL 编码
4. **\u0026 vs &** - 这只是 JSON 编码的差异，实际使用时都是 `&` 符号

## 常见问题

### Q: 如何处理支付成功回调？

A: 设置 `redirectUrl` 和 `notifyUrl`：

```go
options := &models.DeepLinkOptions{
    RedirectURL: "https://yoursite.com/success",  // 用户重定向
    NotifyURL:   "https://yoursite.com/webhook",  // 服务器回调
}
```

### Q: 为什么生成的 Deep Link 不工作？

A: 检查以下几点：
1. QR Code 数据是否完整有效
2. GCash 应用是否已安装
3. 是否在移动设备上测试
4. 网络连接是否正常

### Q: 如何自定义 param3 和 param5？

A: 使用 `CustomParam3` 和 `CustomParam5` 选项：

```go
options := &models.DeepLinkOptions{
    CustomParam3: "your_custom_param3",
    CustomParam5: "your_custom_param5",
}
```

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！

## 联系方式

如有问题，请创建 GitHub Issue。

## 更新日志

### v1.0.0 (2025-01-10)
- 初始版本发布
- 支持 EMVCo QR Code 解析
- 支持 GCash Deep Link 生成
- 提供 HTTP API 接口
- 完整的测试覆盖

---

**注意**：此工具仅用于生成 Deep Link，实际支付流程需要有效的 GCash 商户账号和相关配置。
