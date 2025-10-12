# GCash Deep Link 手动编辑测试指南

## 功能说明

网页界面已增强，支持：
1. ✅ 自动生成 GCash Deep Link
2. ✅ **在 textarea 中手动编辑 Deep Link**
3. ✅ 使用手动编辑后的链接打开 GCash
4. ✅ 复制手动编辑后的链接

## 测试步骤

### 1. 启动服务器

```bash
# 方式 1: 直接运行
go run main.go

# 方式 2: 使用 Makefile
make run

# 方式 3: 使用构建的可执行文件
./build/gcash-deeplink
```

服务器会自动打开浏览器访问 http://localhost:9000

### 2. 生成 Deep Link

在网页中输入测试 QR Code：

```
00020101021228530011ph.ppmi.p2m0111SRCPPHM2XXX0312MRCHNT-4H3TZ05030005204519953036085406100.005802PH5925SOCMED DIGITAL MARKETING 6010MakatiCity62650010ph.starpay0315SOCMED DIGITAL 0509OR#1Z1CSC0708TodayPay0803***88290012ph.ppmi.qrph0109OR#1Z1CSC63040275
```

点击"生成 Deep Link"按钮。

### 3. 查看生成结果

生成成功后，您会看到：
- 商户信息（商户名称、金额、城市、店铺 ID）
- **可编辑的 Deep Link textarea**

生成的 Deep Link 示例：
```
gcash://com.mynt.gcash/app/006300000800?acqInfo=OR%231Z1CSC&bizNo=null&clientId=1760244872419&lucky=false&merchantCategoryCode=5199&merchantName=SOCMED+DIGITAL+MARKETING&orderAmount=100.005802&param3=99960005~ph.ppmi.p2m~~~000&qrCode=...&qrCodeFormat=EMVCO&sub=p2mpay
```

### 4. 手动编辑 Deep Link

现在您可以在 textarea 中直接编辑链接。常见的编辑场景：

#### 场景 1: 修改金额
将 `orderAmount=100.005802` 改为 `orderAmount=50.00`

```
gcash://com.mynt.gcash/app/006300000800?acqInfo=OR%231Z1CSC&bizNo=null&clientId=1760244872419&lucky=false&merchantCategoryCode=5199&merchantName=SOCMED+DIGITAL+MARKETING&orderAmount=50.00&param3=99960005~ph.ppmi.p2m~~~000&qrCode=...&qrCodeFormat=EMVCO&sub=p2mpay
```

#### 场景 2: 修改商户名称
将 `merchantName=SOCMED+DIGITAL+MARKETING` 改为 `merchantName=TEST+SHOP`

#### 场景 3: 添加订单 ID
在链接末尾添加 `&orderId=TEST-ORDER-123`

#### 场景 4: 启用幸运抽奖
将 `lucky=false` 改为 `lucky=true`

#### 场景 5: 添加回调 URL
在链接末尾添加：
```
&redirectUrl=https://yoursite.com/success&notifyUrl=https://yoursite.com/webhook
```

### 5. 测试修改后的链接

#### 方式 1: 点击"打开 GCash"按钮
- 点击绿色的"🚀 打开 GCash"按钮
- 如果在移动设备上且安装了 GCash，会自动打开应用
- 在桌面浏览器上，会尝试打开但可能失败（这是正常的）

#### 方式 2: 复制链接
- 点击"📋 复制链接"按钮
- 链接会被复制到剪贴板
- 可以：
  - 通过 AirDrop/微信等发送到手机
  - 在移动浏览器中粘贴访问
  - 在终端中测试：`open "gcash://..."`（macOS）

### 6. 在移动设备上测试

#### iOS / Android:
1. 确保已安装 GCash 应用
2. 将链接发送到手机（通过邮件、消息等）
3. 在移动浏览器中打开链接
4. 或直接点击"打开 GCash"按钮（如果在手机浏览器中打开页面）

#### 预期行为：
- ✅ GCash 应用会自动打开
- ✅ 显示支付页面，金额为您修改后的值
- ✅ 商户名称显示为您修改后的值

## 功能特性

### ✨ 新增特性
- 📝 **可编辑 textarea**: Deep Link 显示在可编辑的文本区域
- ✏️ **实时编辑**: 所有修改即时生效
- 🎯 **高亮聚焦**: 点击 textarea 时会高亮边框
- 📋 **智能复制**: 复制功能会获取编辑后的最新内容
- 🚀 **打开测试**: 打开 GCash 使用编辑后的最新链接

### 🎨 UI 改进
- 清晰的标签："生成的 Deep Link (可手动编辑)"
- 等宽字体显示，便于阅读和编辑
- 自动换行，支持长链接
- 可调整高度（支持拖动调整）
- 聚焦时背景色变化，提供视觉反馈

## 测试用例

### 测试用例 1: 基础生成和编辑
1. 输入 QR Code
2. 点击生成
3. 在 textarea 中修改金额
4. 点击"打开 GCash"
5. ✅ 应使用修改后的金额

### 测试用例 2: 复制功能
1. 生成 Deep Link
2. 手动修改参数
3. 点击"复制链接"
4. 粘贴到文本编辑器
5. ✅ 应显示修改后的链接

### 测试用例 3: 多次编辑
1. 生成链接
2. 修改金额
3. 点击打开（或复制）
4. 再次修改商户名称
5. 再次点击打开（或复制）
6. ✅ 每次都应使用最新的编辑内容

### 测试用例 4: 完全自定义链接
1. 不生成链接
2. 直接在结果区域的 textarea 中粘贴一个完整的 GCash Deep Link
3. 点击"打开 GCash"
4. ✅ 应能正常打开

## API 测试

### 使用 curl 测试生成 API

```bash
curl -X POST http://localhost:9000/api/generate \
  -H "Content-Type: application/json" \
  -d '{
    "qrCode": "00020101021228530011ph.ppmi.p2m0111SRCPPHM2XXX0312MRCHNT-4H3TZ05030005204519953036085406100.005802PH5925SOCMED DIGITAL MARKETING 6010MakatiCity62650010ph.starpay0315SOCMED DIGITAL 0509OR#1Z1CSC0708TodayPay0803***88290012ph.ppmi.qrph0109OR#1Z1CSC63040275",
    "orderId": "TEST-12345",
    "merchantId": "217020000119199251998"
  }' | python3 -m json.tool
```

### 健康检查

```bash
curl http://localhost:9000/health
```

## 常见问题

### Q: 为什么点击"打开 GCash"没反应？
A: 在桌面浏览器上这是正常的，因为 GCash 只能在移动设备上打开。请使用"复制链接"功能，将链接发送到手机测试。

### Q: 可以修改哪些参数？
A: 您可以修改任何 URL 参数，包括：
- `orderAmount` - 订单金额
- `merchantName` - 商户名称
- `orderId` - 订单 ID
- `redirectUrl` - 回调 URL
- `lucky` - 是否启用幸运抽奖
- 以及其他任何参数

### Q: 修改后的链接安全吗？
A: 手动修改链接时请注意：
- ⚠️ 修改金额等敏感参数可能导致支付失败
- ⚠️ GCash 服务器会验证参数的有效性
- ✅ 此功能主要用于测试和调试
- ✅ 生产环境应使用 API 生成标准链接

### Q: 如何在手机上测试？
A: 最简单的方式：
1. 在电脑上打开页面
2. 生成并编辑链接
3. 点击"复制链接"
4. 通过消息/邮件发送到手机
5. 在手机上点击链接

或者直接在手机浏览器中打开 http://localhost:9000（需要在同一网络）

## 技术实现

### 前端变更
- 将只读的 `<div>` 改为可编辑的 `<textarea>`
- 添加专门的 CSS 样式 `.deeplink-edit`
- 修改 JavaScript 函数从 textarea 读取值：
  - `openGCash()` - 使用 `deepLinkEdit.value`
  - `copyLink()` - 使用 `deepLinkEdit.value`
  - `displayResult()` - 写入 `deepLinkEdit.value`

### 代码位置
- HTML: `/Users/qinyuanmao/Downloads/gcash-deeplink/public/index.html`
- 服务器: `/Users/qinyuanmao/Downloads/gcash-deeplink/main.go`

## 下一步

现在您可以：
1. ✅ 在浏览器中测试手动编辑功能
2. ✅ 将链接发送到手机测试实际打开 GCash
3. ✅ 尝试不同的参数组合
4. ✅ 验证修改后的参数是否生效

祝测试愉快！🎉
