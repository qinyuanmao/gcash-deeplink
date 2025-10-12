# GCash Deep Link Generator - 使用指南

## 快速开始

### 1. 安装和运行

```bash
# 1. 进入项目目录
cd gcash-deeplink

# 2. 运行程序（启动 HTTP API 服务器）
go run main.go

# 或者先编译再运行
go build -o gcash-deeplink
./gcash-deeplink
```

服务器将在 `http://localhost:9000` 启动

### 2. 运行示例

```bash
# 查看内置示例
go run main.go examples
```

### 3. 使用前端界面

1. 启动后端服务：`go run main.go`
2. 在浏览器中打开 `example.html`
3. 输入 EMVCo QR Code 数据
4. 点击"生成 Deep Link"

## API 使用详解

### 端点 1: 解析 QR Code

**请求：**

```bash
curl -X POST http://localhost:9000/api/parse \
  -H "Content-Type: application/json" \
  -d '{
    "qrCode": "00020101021228530011ph.ppmi.p2m0111SRCPPHM2XXX0312MRCHNT-4H3TZ05030005204519953036085406100.005802PH5925SOCMED DIGITAL MARKETING 6010MakatiCity62650010ph.starpay0315SOCMED DIGITAL 0509OR#1Z1CSC0708TodayPay0803***88290012ph.ppmi.qrph0109OR#1Z1CSC63040275"
  }'
```

**响应：**

```json
{
  "success": true,
  "data": {
    "Version": "01",
    "Amount": "100.00",
    "Currency": "608",
    "CountryCode": "PH",
    "MerchantName": "SOCMED DIGITAL MARKETING",
    "MerchantCity": "MakatiCity",
    "MerchantCategoryCode": "5199",
    "ShopID": "MRCHNT-4H3TZ",
    "BankCode": "SRCPPHM2XXX",
    "AcqInfo": "OR#1Z1CSC",
    "CRC": "0275"
  }
}
```

### 端点 2: 生成 Deep Link

**基础请求：**

```bash
curl -X POST http://localhost:9000/api/generate \
  -H "Content-Type: application/json" \
  -d '{
    "qrCode": "00020101021228530011ph.ppmi.p2m..."
  }'
```

**完整请求（带所有参数）:**

```bash
curl -X POST http://localhost:9000/api/generate \
  -H "Content-Type: application/json" \
  -d '{
    "qrCode": "00020101021228530011ph.ppmi.p2m...",
    "orderId": "ORDER-12345",
    "merchantId": "217020000119199251998",
    "redirectUrl": "https://myshop.com/success",
    "notifyUrl": "https://myshop.com/webhook",
    "paymentType": "010"
  }'
```

**响应：**

```json
{
  "success": true,
  "deepLink": "gcash://com.mynt.gcash/app/006300000800?qrCode=...",
  "parsedData": {
    "Amount": "100.00",
    "MerchantName": "SOCMED DIGITAL MARKETING",
    "MerchantCity": "MakatiCity",
    "ShopID": "MRCHNT-4H3TZ"
  },
  "generatedAt": "2025-01-10T10:30:00Z"
}
```

### 端点 3: 验证 QR Code

**请求：**

```bash
curl -X POST http://localhost:9000/api/validate \
  -H "Content-Type: application/json" \
  -d '{
    "qrCode": "00020101021228530011ph.ppmi.p2m..."
  }'
```

**响应（有效）:**

```json
{
  "valid": true,
  "errors": []
}
```

**响应（无效）:**

```json
{
  "valid": false,
  "errors": [
    "QR Code 数据长度过短",
    "缺少必需字段 Tag 59: 商户名称"
  ]
}
```

## Go 代码使用示例

### 示例 1: 基础用法

```go
package main

import (
    "fmt"
    "gcash-deeplink/generator"
    "gcash-deeplink/parser"
)

func main() {
    qrCode := "00020101021228530011ph.ppmi.p2m..."
    
    // 解析
    p := parser.NewEMVCoParser()
    data, _ := p.Parse(qrCode)
    
    // 生成
    g := generator.NewDeepLinkGenerator()
    result, _ := g.Generate(data, nil)
    
    fmt.Println(result.DeepLink)
}
```

### 示例 2: 带验证

```go
package main

import (
    "fmt"
    "log"
    "gcash-deeplink/generator"
    "gcash-deeplink/parser"
)

func main() {
    qrCode := "00020101021228530011ph.ppmi.p2m..."
    
    // 验证
    p := parser.NewEMVCoParser()
    validation := p.Validate(qrCode)
    
    if !validation.Valid {
        log.Fatalf("QR Code 无效: %v", validation.Errors)
    }
    
    // 生成
    g := generator.NewDeepLinkGenerator()
    result, err := g.GenerateWithValidation(qrCode, nil)
    
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("✅ 生成成功！")
    fmt.Println("Deep Link:", result.DeepLink)
}
```

### 示例 3: 订单支付

```go
package main

import (
    "fmt"
    "time"
    "gcash-deeplink/generator"
    "gcash-deeplink/models"
)

func main() {
    qrCode := "00020101021228530011ph.ppmi.p2m..."
    
    options := &models.DeepLinkOptions{
        PaymentType: models.PaymentTypeDynamic,
        OrderID:     fmt.Sprintf("ORDER-%d", time.Now().Unix()),
        RedirectURL: "https://myshop.com/payment/success",
        NotifyURL:   "https://myshop.com/api/gcash/webhook",
        MerchantID:  "217020000119199251998",
    }
    
    g := generator.NewDeepLinkGenerator()
    result, _ := g.GenerateWithValidation(qrCode, options)
    
    fmt.Printf("订单号: %s\n", options.OrderID)
    fmt.Printf("Deep Link: %s\n", result.DeepLink)
}
```

### 示例 4: 多策略生成

```go
package main

import (
    "fmt"
    "gcash-deeplink/generator"
    "gcash-deeplink/parser"
)

func main() {
    qrCode := "00020101021228530011ph.ppmi.p2m..."
    
    p := parser.NewEMVCoParser()
    data, _ := p.Parse(qrCode)
    
    g := generator.NewDeepLinkGenerator()
    strategies := g.GenerateMultiple(data)
    
    for name, link := range strategies {
        fmt.Printf("%s: %s\n\n", name, link)
    }
}
```

## JavaScript 使用示例

### 示例 1: 基础调用

```javascript
async function generateDeepLink(qrCode) {
    try {
        const response = await fetch('http://localhost:9000/api/generate', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ qrCode })
        });
        
        const data = await response.json();
        
        if (data.success) {
            console.log('Deep Link:', data.deepLink);
            return data.deepLink;
        } else {
            console.error('生成失败：', data.error);
        }
    } catch (error) {
        console.error('请求失败：', error);
    }
}

// 使用
const qrCode = "00020101021228530011ph.ppmi.p2m...";
generateDeepLink(qrCode);
```

### 示例 2: 带参数

```javascript
async function generateDeepLinkWithOptions(qrCode, options) {
    const response = await fetch('http://localhost:9000/api/generate', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            qrCode,
            orderId: options.orderId,
            redirectUrl: options.redirectUrl,
            notifyUrl: options.notifyUrl,
            paymentType: options.paymentType
        })
    });
    
    return await response.json();
}

// 使用
const result = await generateDeepLinkWithOptions(qrCode, {
    orderId: 'ORDER-12345',
    redirectUrl: 'https://myshop.com/success',
    notifyUrl: 'https://myshop.com/webhook',
    paymentType: '010'
});
```

### 示例 3: React 组件

```jsx
import React, { useState } from 'react';

function GCashPayment() {
    const [qrCode, setQrCode] = useState('');
    const [deepLink, setDeepLink] = useState('');
    const [loading, setLoading] = useState(false);

    const generateDeepLink = async () => {
        setLoading(true);
        
        try {
            const response = await fetch('http://localhost:9000/api/generate', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ qrCode })
            });
            
            const data = await response.json();
            
            if (data.success) {
                setDeepLink(data.deepLink);
            }
        } catch (error) {
            console.error(error);
        } finally {
            setLoading(false);
        }
    };

    return (
        <div>
            <textarea 
                value={qrCode} 
                onChange={(e) => setQrCode(e.target.value)}
                placeholder="输入 QR Code"
            />
            <button onClick={generateDeepLink} disabled={loading}>
                {loading ? '生成中...' : '生成 Deep Link'}
            </button>
            {deepLink && (
                <div>
                    <a href={deepLink}>打开 GCash</a>
                </div>
            )}
        </div>
    );
}
```

## 常见使用场景

### 场景 1: 电商订单支付

```go
func CreateOrderPayment(orderID string, amount float64, qrCode string) (string, error) {
    options := &models.DeepLinkOptions{
        PaymentType: models.PaymentTypeDynamic,
        OrderID:     orderID,
        RedirectURL: fmt.Sprintf("https://shop.com/order/%s/success", orderID),
        NotifyURL:   "https://shop.com/api/payment/webhook",
    }
    
    g := generator.NewDeepLinkGenerator()
    result, err := g.GenerateWithValidation(qrCode, options)
    
    if err != nil {
        return "", err
    }
    
    // 保存到数据库
    SavePaymentToDatabase(orderID, result.DeepLink)
    
    return result.DeepLink, nil
}
```

### 场景 2: 店铺收款码

```go
func GenerateStoreQRCode(storeID string, qrCode string) (string, error) {
    options := &models.DeepLinkOptions{
        PaymentType: models.PaymentTypeStatic,
        MerchantID:  storeID,
    }
    
    g := generator.NewDeepLinkGenerator()
    result, err := g.GenerateWithValidation(qrCode, options)
    
    if err != nil {
        return "", err
    }
    
    return result.DeepLink, nil
}
```

### 场景 3: 批量生成

```go
func BatchGenerateDeepLinks(qrCodes []string) ([]string, error) {
    var deepLinks []string
    g := generator.NewDeepLinkGenerator()
    
    for _, qrCode := range qrCodes {
        result, err := g.GenerateWithValidation(qrCode, nil)
        if err != nil {
            return nil, err
        }
        deepLinks = append(deepLinks, result.DeepLink)
    }
    
    return deepLinks, nil
}
```

## 测试

### 运行单元测试

```bash
go test -v
```

### 运行特定测试

```bash
go test -v -run TestParseEMVCoQR
```

### 测试覆盖率

```bash
go test -cover
go test -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### 基准测试

```bash
go test -bench=. -benchmem
```

## 故障排除

### 问题 1: 后端服务无法启动

**症状**: `address already in use`

**解决方案**:
```bash
# 查找占用端口的进程
lsof -i :9000

# 杀死进程
kill -9 <PID>

# 或者使用不同的端口
PORT=8081 go run main.go
```

### 问题 2: CORS 错误

**症状**: 前端无法访问 API

**解决方案**: 代码中已包含 CORS 支持，确保：
1. 后端服务正在运行
2. 使用正确的 API 地址
3. 浏览器允许跨域请求

### 问题 3: Deep Link 不工作

**可能原因**:
1. QR Code 数据无效或不完整
2. GCash 应用未安装
3. 不在移动设备上测试
4. 网络连接问题

**调试步骤**:
1. 验证 QR Code: `POST /api/validate`
2. 检查生成的 Deep Link 格式
3. 在真实设备上测试
4. 查看 GCash 应用日志

## 性能优化

### 1. 使用连接池

```go
var (
    parserPool = sync.Pool{
        New: func() interface{} {
            return parser.NewEMVCoParser()
        },
    }
)

func ParseQRCode(qrCode string) (*models.EMVCoData, error) {
    p := parserPool.Get().(*parser.EMVCoParser)
    defer parserPool.Put(p)
    
    return p.Parse(qrCode)
}
```

### 2. 缓存结果

```go
var cache = make(map[string]*models.DeepLinkResult)
var cacheMutex sync.RWMutex

func GenerateWithCache(qrCode string) (*models.DeepLinkResult, error) {
    cacheMutex.RLock()
    if result, exists := cache[qrCode]; exists {
        cacheMutex.RUnlock()
        return result, nil
    }
    cacheMutex.RUnlock()
    
    // 生成新的
    g := generator.NewDeepLinkGenerator()
    result, err := g.GenerateWithValidation(qrCode, nil)
    if err != nil {
        return nil, err
    }
    
    // 缓存
    cacheMutex.Lock()
    cache[qrCode] = result
    cacheMutex.Unlock()
    
    return result, nil
}
```

## 生产环境部署

### 1. 编译

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o gcash-deeplink-linux

# Windows
GOOS=windows GOARCH=amd64 go build -o gcash-deeplink.exe

# macOS
GOOS=darwin GOARCH=amd64 go build -o gcash-deeplink-mac
```

### 2. 使用 systemd (Linux)

创建 `/etc/systemd/system/gcash-deeplink.service`:

```ini
[Unit]
Description=GCash Deep Link Generator
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/gcash-deeplink
ExecStart=/opt/gcash-deeplink/gcash-deeplink
Restart=always

[Install]
WantedBy=multi-user.target
```

启动服务：

```bash
sudo systemctl daemon-reload
sudo systemctl start gcash-deeplink
sudo systemctl enable gcash-deeplink
```

### 3. 使用 Docker

创建 `Dockerfile`:

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o gcash-deeplink

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/gcash-deeplink .
EXPOSE 9000
CMD ["./gcash-deeplink"]
```

构建和运行：

```bash
docker build -t gcash-deeplink .
docker run -p 9000:9000 gcash-deeplink
```

## 更多资源

- [EMVCo QR Code 规范](https://www.emvco.com/emv-technologies/qr-codes/)
- [GCash 开发者文档](https://www.gcash.com/developers) (如果可用)
- [Go 官方文档](https://golang.org/doc/)

## 支持

如有问题，请创建 GitHub Issue 或联系技术支持。
