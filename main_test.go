package main

import (
	"net/url"
	"testing"

	"github.com/qinyuanmao/gcash-deeplink/generator"
	"github.com/qinyuanmao/gcash-deeplink/models"
	"github.com/qinyuanmao/gcash-deeplink/parser"
)

func TestParseEMVCoQR(t *testing.T) {
	qrCode := "00020101021228530011ph.ppmi.p2m0111SRCPPHM2XXX0312MRCHNT-4H3TZ05030005204519953036085406100.005802PH5925SOCMED DIGITAL MARKETING 6010MakatiCity62650010ph.starpay0315SOCMED DIGITAL 0509OR#1Z1CSC0708TodayPay0803***88290012ph.ppmi.qrph0109OR#1Z1CSC63040275"

	p := parser.NewEMVCoParser()
	data, err := p.Parse(qrCode)

	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}

	if data.Amount != "100.00" {
		t.Errorf("金额错误: got %s, want 100.00", data.Amount)
	}

	if data.MerchantName != "SOCMED DIGITAL MARKETING" {
		t.Errorf("商户名称错误: got %s", data.MerchantName)
	}

	if data.MerchantCity != "MakatiCity" {
		t.Errorf("城市错误: got %s", data.MerchantCity)
	}

	if data.ShopID != "MRCHNT-4H3TZ" {
		t.Errorf("ShopID 错误: got %s", data.ShopID)
	}
}

func TestGenerateDeepLink(t *testing.T) {
	qrCode := "00020101021228530011ph.ppmi.p2m0111SRCPPHM2XXX0312MRCHNT-4H3TZ05030005204519953036085406100.005802PH5925SOCMED DIGITAL MARKETING 6010MakatiCity62650010ph.starpay0315SOCMED DIGITAL 0509OR#1Z1CSC0708TodayPay0803***88290012ph.ppmi.qrph0109OR#1Z1CSC63040275"

	g := generator.NewDeepLinkGenerator()
	options := &models.DeepLinkOptions{
		PaymentType: models.PaymentTypeStandard,
	}

	result, err := g.GenerateWithValidation(qrCode, options)

	if err != nil {
		t.Fatalf("生成失败: %v", err)
	}

	if !result.Success {
		t.Error("生成结果应该成功")
	}

	if result.DeepLink == "" {
		t.Error("Deep Link 不应为空")
	}

	if result.ParsedData == nil {
		t.Error("解析数据不应为空")
	}
}

func TestValidateQRCode(t *testing.T) {
	tests := []struct {
		name    string
		qrCode  string
		wantErr bool
	}{
		{
			name:    "有效的 QR Code",
			qrCode:  "00020101021228530011ph.ppmi.p2m0111SRCPPHM2XXX0312MRCHNT-4H3TZ05030005204519953036085406100.005802PH5925SOCMED DIGITAL MARKETING 6010MakatiCity62650010ph.starpay0315SOCMED DIGITAL 0509OR#1Z1CSC0708TodayPay0803***88290012ph.ppmi.qrph0109OR#1Z1CSC63040275",
			wantErr: false,
		},
		{
			name:    "空 QR Code",
			qrCode:  "",
			wantErr: true,
		},
		{
			name:    "过短的 QR Code",
			qrCode:  "0002010102",
			wantErr: true,
		},
	}

	p := parser.NewEMVCoParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := p.Validate(tt.qrCode)
			if (len(result.Errors) > 0) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", result.Errors, tt.wantErr)
			}
		})
	}
}

func TestGenerateMultipleStrategies(t *testing.T) {
	qrCode := "00020101021228530011ph.ppmi.p2m0111SRCPPHM2XXX0312MRCHNT-4H3TZ05030005204519953036085406100.005802PH5925SOCMED DIGITAL MARKETING 6010MakatiCity62650010ph.starpay0315SOCMED DIGITAL 0509OR#1Z1CSC0708TodayPay0803***88290012ph.ppmi.qrph0109OR#1Z1CSC63040275"

	p := parser.NewEMVCoParser()
	data, err := p.Parse(qrCode)
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}

	g := generator.NewDeepLinkGenerator()
	strategies := g.GenerateMultiple(data)

	expectedStrategies := []string{"minimal", "dynamic", "with_callback"}
	for _, strategy := range expectedStrategies {
		if link, exists := strategies[strategy]; !exists || link == "" {
			t.Errorf("策略 %s 应该存在且不为空", strategy)
		}
	}
}

// TestAcqInfoFallbackToShopID 测试 Tag 62-05 为空时 acqInfo 回退到 Tag 28-03
func TestAcqInfoFallbackToShopID(t *testing.T) {
	// 构造 QR Code: Tag 28-03 有值 (2163386327968797571)，Tag 62 无 subtag 05
	// Tag 28: 00=ph.ppmi.p2m(11), 01=GXCHPHM2XXX(11), 03=2163386327968797571(19) → 总长53
	// Tag 62: 00=ph.ppmi.qrph(12), 03=TESTORDER1(10), 07=TERM0001(8) → 总长42
	qrCode := "00020101021228530011ph.ppmi.p2m0111GXCHPHM2XXX031921633863279687975715204519953036085406100.005802PH5910TESTMERCH16006Manila62420012ph.ppmi.qrph0310TESTORDER10708TERM000163041234"

	p := parser.NewEMVCoParser()
	data, err := p.Parse(qrCode)
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}

	// ShopID 应该来自 Tag 28-03
	if data.ShopID != "2163386327968797571" {
		t.Errorf("ShopID 错误: got %q, want %q", data.ShopID, "2163386327968797571")
	}

	// ReferenceLabel (Tag 62-05) 应该为空
	if data.ReferenceLabel != "" {
		t.Errorf("ReferenceLabel 应为空: got %q", data.ReferenceLabel)
	}

	// AcqInfo 应该回退到 ShopID (Tag 28-03)
	if data.AcqInfo != "2163386327968797571" {
		t.Errorf("AcqInfo 回退错误: got %q, want %q", data.AcqInfo, "2163386327968797571")
	}

	// TerminalLabel 应该来自 Tag 62-07
	if data.TerminalLabel != "TERM0001" {
		t.Errorf("TerminalLabel 错误: got %q, want %q", data.TerminalLabel, "TERM0001")
	}
}

// TestAcqInfoPriority6205 测试 Tag 62-05 优先于 28-03 (符合 Luca 的模板)
func TestAcqInfoPriority6205(t *testing.T) {
	// acqInfo 默认使用 Tag 62-05 (Reference Label)
	qrCode := "00020101021228790011ph.ppmi.p2m0111PAEYPHM2XXX0324VkHUE2Fz8Ee2YxnTVPX34TZs041003030028860503010520473995303608540520.005802PH5916NEXA ONLINE SHOP6013General Trias62430012ph.ppmi.qrph0306lsFK7X05062110000803***88440012ph.ppmi.qrph0124VkHUE2Fz8Ee2YxnTVPX34TZs63042A5A"

	p := parser.NewEMVCoParser()
	data, err := p.Parse(qrCode)
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}

	// ReferenceLabel 应该来自 Tag 62-05
	if data.ReferenceLabel != "211000" {
		t.Errorf("ReferenceLabel 错误: got %q, want %q", data.ReferenceLabel, "211000")
	}

	// AcqInfo 应优先使用 62-05 (Reference Label)
	if data.AcqInfo != "211000" {
		t.Errorf("AcqInfo 错误: got %q, want %q", data.AcqInfo, "211000")
	}
}

// TestDeepLinkWithTerminalLabel 测试带 Terminal Label 的 param5 格式
func TestDeepLinkWithTerminalLabel(t *testing.T) {
	g := generator.NewDeepLinkGenerator()

	data := &models.EMVCoData{
		ShopID:         "SHOP123",
		OrderID:        "ORDER456",
		ReferenceLabel: "REF789",
		AcqInfo:        "REF789",
		TerminalLabel:  "TERM001",
		Amount:         "100.00",
		MerchantName:   "TEST",
		RawData:        "test",
	}

	result, err := g.Generate(data, &models.DeepLinkOptions{
		PaymentType: models.PaymentTypeStandard,
	})
	if err != nil {
		t.Fatalf("生成失败: %v", err)
	}

	// 检查 param5 包含 TerminalLabel
	// 格式: ShopID~OrderID~TerminalLabel~~AcqInfo
	expected := "SHOP123~ORDER456~TERM001~~REF789"
	if !containsParam(result.DeepLink, "param5", expected) {
		t.Errorf("param5 格式错误，期望包含 %q，实际 DeepLink: %s", expected, result.DeepLink)
	}

	// 检查 acqInfo 参数
	if !containsParam(result.DeepLink, "acqInfo", "REF789") {
		t.Errorf("acqInfo 参数错误，期望 REF789")
	}
}

func containsParam(deepLink, key, value string) bool {
	parsed, err := url.Parse(deepLink)
	if err != nil || parsed == nil {
		return false
	}
	return parsed.Query().Get(key) == value
}

// TestCoinsOldFormatDefaultAcqInfo 测试旧格式 Coins QR 无 KnownUID 时 acqInfo 取 62-05
func TestCoinsOldFormatDefaultAcqInfo(t *testing.T) {
	// 旧格式: Tag 28-03 = Coins Reference Number, Tag 62-05 = UID
	// 无 KnownUID 时，默认 62-05 优先 → acqInfo = UID (需要 KnownUID 来修正)
	qrCode := "00020101021228600011ph.ppmi.p2m0111DCPHPHM1XXX03192163953825260794775050301152044816530360854031005802PH5909PoLhevWiN6011Baguio city62380011ph.ppmi.p2m051920828990834787223046304178C"

	p := parser.NewEMVCoParser()
	data, err := p.Parse(qrCode)
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}

	// 默认 62-05 优先 → acqInfo = UID
	if data.AcqInfo != "2082899083478722304" {
		t.Errorf("AcqInfo 错误: got %q, want %q", data.AcqInfo, "2082899083478722304")
	}
}

// TestKnownUIDAsAcqInfo 测试 KnownUID 直接作为 acqInfo (商户号)
func TestKnownUIDAsAcqInfo(t *testing.T) {
	// KnownUID = 商户号，直接作为 acqInfo
	// shopId/tfrAcctNo = Tag 28-03 (订单号)
	// param5 = 订单号~GCash名称~~商户号
	qrCode := "00020101021228600011ph.ppmi.p2m0111DCPHPHM1XXX03192165045737094936496050301152044816530360854031555802PH5909BuzhuYazi6011Baguio city62380011ph.ppmi.p2m051920828990834787223046304A0F1"
	knownUID := "2082899083478722304"

	p := parser.NewEMVCoParser()
	data, err := p.Parse(qrCode)
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}

	g := generator.NewDeepLinkGenerator()
	result, err := g.Generate(data, &models.DeepLinkOptions{
		PaymentType: models.PaymentTypeDynamic,
		KnownUID:    knownUID,
	})
	if err != nil {
		t.Fatalf("生成失败: %v", err)
	}

	parsed, _ := url.Parse(result.DeepLink)
	params := parsed.Query()

	// acqInfo = KnownUID (商户号)
	if params.Get("acqInfo") != knownUID {
		t.Errorf("acqInfo 错误: got %q, want %q", params.Get("acqInfo"), knownUID)
	}
	// shopId = Tag 28-03 (订单号)
	if params.Get("shopId") != "2165045737094936496" {
		t.Errorf("shopId 错误: got %q, want %q", params.Get("shopId"), "2165045737094936496")
	}
	// tfrAcctNo = Tag 28-03 (订单号)
	if params.Get("tfrAcctNo") != "2165045737094936496" {
		t.Errorf("tfrAcctNo 错误: got %q, want %q", params.Get("tfrAcctNo"), "2165045737094936496")
	}
	// param5 = 订单号~GCash名称~~商户号 (无 Tag 62-03 时用 merchantName)
	expectedParam5 := "2165045737094936496~BuzhuYazi~~2082899083478722304"
	if params.Get("param5") != expectedParam5 {
		t.Errorf("param5 错误: got %q, want %q", params.Get("param5"), expectedParam5)
	}
}

// TestStandardMerchantNoKnownUID 测试标准商户 (PAEYPHM2XXX) 无 KnownUID
func TestStandardMerchantNoKnownUID(t *testing.T) {
	// 标准商户: shopId=28-03, acqInfo=62-05, param5=shopId~orderID(62-03)~~~acqInfo
	qrCode := "00020101021228790011ph.ppmi.p2m0111PAEYPHM2XXX0324cPkF7TJFkyii3Ri9nUJ6qmGQ0410030300288605030105204504653036085406200.005802PH5909BuzhuYazi6011Taguig City62430012ph.ppmi.qrph0306jqdpjj05062110000803***88440012ph.ppmi.qrph0124cPkF7TJFkyii3Ri9nUJ6qmGQ63044FB3"

	p := parser.NewEMVCoParser()
	data, err := p.Parse(qrCode)
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}

	g := generator.NewDeepLinkGenerator()
	result, err := g.Generate(data, &models.DeepLinkOptions{PaymentType: models.PaymentTypeDynamic})
	if err != nil {
		t.Fatalf("生成失败: %v", err)
	}

	parsed, _ := url.Parse(result.DeepLink)
	params := parsed.Query()

	if params.Get("shopId") != "cPkF7TJFkyii3Ri9nUJ6qmGQ" {
		t.Errorf("shopId 错误: got %q", params.Get("shopId"))
	}
	if params.Get("acqInfo") != "211000" {
		t.Errorf("acqInfo 错误: got %q, want %q", params.Get("acqInfo"), "211000")
	}
	// param5: 有 orderID(62-03) 时用 orderID~terminalLabel 格式
	expectedParam5 := "cPkF7TJFkyii3Ri9nUJ6qmGQ~jqdpjj~~~211000"
	if params.Get("param5") != expectedParam5 {
		t.Errorf("param5 错误: got %q, want %q", params.Get("param5"), expectedParam5)
	}
}

func BenchmarkParseQRCode(b *testing.B) {
	qrCode := "00020101021228530011ph.ppmi.p2m0111SRCPPHM2XXX0312MRCHNT-4H3TZ05030005204519953036085406100.005802PH5925SOCMED DIGITAL MARKETING 6010MakatiCity62650010ph.starpay0315SOCMED DIGITAL 0509OR#1Z1CSC0708TodayPay0803***88290012ph.ppmi.qrph0109OR#1Z1CSC63040275"
	p := parser.NewEMVCoParser()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = p.Parse(qrCode)
	}
}

func BenchmarkGenerateDeepLink(b *testing.B) {
	qrCode := "00020101021228530011ph.ppmi.p2m0111SRCPPHM2XXX0312MRCHNT-4H3TZ05030005204519953036085406100.005802PH5925SOCMED DIGITAL MARKETING 6010MakatiCity62650010ph.starpay0315SOCMED DIGITAL 0509OR#1Z1CSC0708TodayPay0803***88290012ph.ppmi.qrph0109OR#1Z1CSC63040275"
	g := generator.NewDeepLinkGenerator()
	options := &models.DeepLinkOptions{
		PaymentType: models.PaymentTypeStandard,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = g.GenerateWithValidation(qrCode, options)
	}
}
