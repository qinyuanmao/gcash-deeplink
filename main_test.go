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

// TestAcqInfoFromReferenceLabel 测试 Tag 62-05 有值时优先使用
func TestAcqInfoFromReferenceLabel(t *testing.T) {
	// Tag 62 有 subtag 05 时，acqInfo 应使用 62-05 的值
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

	// AcqInfo 应该使用 62-05 (Reference Label)
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
