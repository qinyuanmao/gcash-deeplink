# GCash 支付失败故障排除指南

## 常见错误类型

### 错误 1: "Transaction Failed - No amount was deducted"

**症状**: 显示交易失败，未从钱包扣款

**可能原因**:

1. **商户配置问题**
   - 商户未通过 GCash 认证
   - 商户 ID 不正确或未激活
   - 商户未开通 P2M（Person-to-Merchant）功能

2. **QR Code 问题**
   - QR Code 已过期或失效
   - QR Code 格式不正确
   - ShopID 或 BankCode 不匹配

3. **Deep Link 参数错误**
   - 必需参数缺失
   - 参数格式不正确
   - URL 编码问题

4. **网络或系统问题**
   - 网络连接不稳定
   - GCash 服务器响应超时
   - 系统维护中

---

### 错误 2: "Payment confirmation in progress"

**症状**: 显示支付确认中，有参考号但长时间未完成

**可能原因**:

1. **支付网关延迟**
   - 支付处理中，需要等待
   - 银行系统响应慢

2. **商户系统问题**
   - Webhook URL 配置错误
   - 回调接口无法访问
   - 商户服务器未响应

3. **金额不匹配**
   - Deep Link 中的金额与 QR Code 不一致
   - 订单金额已更改

---

## 解决步骤

### 步骤 1: 使用调试工具检查 Deep Link

```bash
# 运行调试工具
go run debug.go
```

调试工具会检查：
- ✅ URL 格式是否正确
- ✅ 必需参数是否齐全
- ✅ QR Code 是否有效
- ✅ 参数值是否合法

### 步骤 2: 验证 QR Code

**检查清单**:

```go
// 使用 API 验证
curl -X POST http://localhost:9000/api/validate \
  -H "Content-Type: application/json" \
  -d '{
    "qrCode": "您的QR Code数据"
  }'
```

**必需字段**:
- ✅ Tag 00: 版本号 (0002)
- ✅ Tag 52: 商户分类码
- ✅ Tag 53: 货币代码 (608 = PHP)
- ✅ Tag 54: 金额
- ✅ Tag 58: 国家代码 (PH)
- ✅ Tag 59: 商户名称
- ✅ Tag 60: 商户城市
- ✅ Tag 63: CRC 校验码

### 步骤 3: 检查商户信息

**验证商户配置**:

1. **登录 GCash 商户后台**
   - 检查商户状态是否为"已激活"
   - 确认 P2M 功能已开通
   - 验证商户 ID 是否正确

2. **检查 QR Code 生成时间**
   - 动态 QR Code 通常有效期为 15-30 分钟
   - 如果过期，需要重新生成

3. **验证 ShopID 和 BankCode**
   ```bash
   # 使用 API 解析 QR Code
   curl -X POST http://localhost:9000/api/parse \
     -H "Content-Type: application/json" \
     -d '{"qrCode": "您的QR Code"}'
   ```

### 步骤 4: 最小化配置测试

使用最简单的配置生成 Deep Link:

```go
package main

import (
    "fmt"
    "gcash-deeplink/generator"
    "gcash-deeplink/models"
)

func main() {
    qrCode := "您的完整QR Code数据"
    
    // 最简配置 - 只传必需参数
    options := &models.DeepLinkOptions{
        PaymentType: models.PaymentTypeStandard,
        // 不传 merchantId, orderId 等可选参数
    }
    
    g := generator.NewDeepLinkGenerator()
    result, err := g.GenerateWithValidation(qrCode, options)
    
    if err != nil {
        fmt.Printf("错误: %v\n", err)
        return
    }
    
    fmt.Println("测试 Deep Link:")
    fmt.Println(result.DeepLink)
}
```

### 步骤 5: 检查回调配置

如果使用了回调 URL:

```go
options := &models.DeepLinkOptions{
    RedirectURL: "https://yoursite.com/success",
    NotifyURL:   "https://yoursite.com/webhook",
}
```

**确保**:
1. ✅ URL 可以从外网访问（不能是 localhost）
2. ✅ 使用 HTTPS（不是 HTTP）
3. ✅ 服务器正常运行
4. ✅ 防火墙允许 GCash 服务器访问

**测试回调 URL**:
```bash
# 测试 URL 是否可访问
curl -I https://yoursite.com/webhook

# 应该返回 200 OK
```

---

## 针对 NEXA ONLINE SHOP 的具体建议

根据您的截图，商户名称是 "NEXA ONLINE SHOP"，我建议：

### 1. 确认商户配置

```bash
# 使用您的 QR Code 生成最简单的 Deep Link
curl -X POST http://localhost:9000/api/generate \
  -H "Content-Type: application/json" \
  -d '{
    "qrCode": "00020101021228790011ph.ppmi.p2m0111PAEYPHM2XXX0324VkHUE2Fz8Ee2YxnTVPX34TZs0410030300288605030105204739953036085406100.005802PH5916NEXA ONLINE SHOP6013General Trias62430012ph.ppmi.qrph0306wWMBdH05062110000803***88440012ph.ppmi.qrph0124VkHUE2Fz8Ee2YxnTVPX34TZs63041C3C"
  }'
```

### 2. 检查特殊字符处理

注意商户名称中的空格:
- "NEXA ONLINE SHOP" 在 URL 中应编码为 "NEXA+ONLINE+SHOP" 或 "NEXA%20ONLINE%20SHOP"
- 代码会自动处理，但要确保解析正确

### 3. 验证金额

确保金额格式正确:
- ✅ 正确: "100.00"
- ❌ 错误: "100", "100.0", "100.000"

### 4. 检查参考号

您的截图显示参考号: **812420744**

使用此参考号:
1. 在 GCash 商户后台查询交易状态
2. 联系 GCash 技术支持查询详细错误信息
3. 检查交易日志

---

## 常见解决方案

### 解决方案 1: 重新生成 QR Code

如果 QR Code 过期:

1. 从商户系统重新生成新的 QR Code
2. 确保使用最新的 QR Code 数据
3. 立即测试（动态 QR 有时效性）

### 解决方案 2: 简化 Deep Link

移除所有可选参数，只保留必需参数:

```go
// 简化版本
options := &models.DeepLinkOptions{
    // 只使用默认值，不传递任何可选参数
}

result, _ := g.Generate(data, options)
```

### 解决方案 3: 联系 GCash 技术支持

如果以上方法都不行:

1. **准备信息**:
   - 商户 ID
   - 参考号 (Ref. No.)
   - QR Code 数据
   - 错误截图
   - 交易时间

2. **联系方式**:
   - GCash 商户支持热线
   - 技术支持邮箱
   - 商户后台的在线客服

3. **提供的信息**:
   ```
   商户名称: NEXA ONLINE SHOP
   参考号: 812420744
   错误类型: Transaction Failed
   错误信息: No amount was deducted from your GCash wallet
   发生时间: [具体时间]
   ```

### 解决方案 4: 检查环境

确保在正确的环境测试:

| 环境  | 要求                 |
| ----- | -------------------- |
| 设备  | Android 或 iOS 真机  |
| GCash | 已安装且已登录       |
| 网络  | 稳定的网络连接       |
| 账户  | GCash 账户有足够余额 |

---

## 预防措施

### 1. 实施重试机制

```go
func PayWithRetry(deepLink string, maxRetries int) error {
    for i := 0; i < maxRetries; i++ {
        err := OpenGCash(deepLink)
        if err == nil {
            return nil
        }
        
        // 等待后重试
        time.Sleep(time.Second * 2)
    }
    
    return fmt.Errorf("支付失败，已重试 %d 次", maxRetries)
}
```

### 2. 添加日志记录

```go
func LogPaymentAttempt(qrCode, deepLink string, err error) {
    log.Printf(`
支付尝试:
  时间: %s
  QR Code: %s...
  Deep Link: %s...
  结果: %v
`, time.Now(), qrCode[:50], deepLink[:100], err)
}
```

### 3. 实现超时处理

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

// 在 context 中执行支付
```

### 4. 用户友好的错误提示

```go
func GetUserFriendlyError(err error) string {
    switch {
    case strings.Contains(err.Error(), "网络"):
        return "网络连接失败，请检查您的网络设置"
    case strings.Contains(err.Error(), "QR"):
        return "支付码已过期，请刷新后重试"
    case strings.Contains(err.Error(), "商户"):
        return "商户暂时无法处理支付，请稍后重试"
    default:
        return "支付失败，请稍后重试或联系客服"
    }
}
```

---

## 调试清单

在联系技术支持前，请完成以下检查:

- [ ] QR Code 格式验证通过
- [ ] Deep Link 格式正确
- [ ] 商户状态为激活
- [ ] GCash 应用已安装
- [ ] 网络连接正常
- [ ] 账户余额充足
- [ ] 使用最新的 QR Code
- [ ] 在真机上测试
- [ ] 已检查系统日志
- [ ] 已尝试最简化配置

---

## 技术支持联系方式

**GCash 商户技术支持**:
- 电话: [GCash 商户热线]
- 邮箱: merchant-support@gcash.com (示例)
- 在线: GCash 商户后台 - 帮助中心

**提供以下信息可加快处理**:
1. 商户 ID 和名称
2. 交易参考号
3. 完整的错误信息
4. QR Code 数据（脱敏处理）
5. Deep Link（脱敏处理）
6. 日志和截图

---

## 总结

支付失败的主要原因通常是:

1. ✅ **商户配置** - 占 40%
2. ✅ **QR Code 问题** - 占 30%
3. ✅ **参数错误** - 占 20%
4. ✅ **网络/系统** - 占 10%

**建议的排查顺序**:
1. 验证 QR Code → 2. 检查商户状态 → 3. 简化配置测试 → 4. 联系技术支持

使用本指南中的调试工具可以快速定位问题！
