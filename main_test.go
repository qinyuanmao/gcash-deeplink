package main

import (
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
