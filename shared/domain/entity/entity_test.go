package entity_test

import (
	"testing"

	"github.com/dealance/shared/domain/entity"
)

func TestSignupStageCanAdvanceTo(t *testing.T) {
	tests := []struct {
		from     entity.SignupStage
		to       entity.SignupStage
		expected bool
	}{
		{entity.SignupStageEmailVerify, entity.SignupStageAuthSetup, true},
		{entity.SignupStageAuthSetup, entity.SignupStageCountry, true},
		{entity.SignupStageCountry, entity.SignupStageRole, true},
		{entity.SignupStageRole, entity.SignupStageKYC, true},
		{entity.SignupStageKYC, entity.SignupStageComplete, true},
		// Invalid transitions
		{entity.SignupStageEmailVerify, entity.SignupStageCountry, false},
		{entity.SignupStageEmailVerify, entity.SignupStageComplete, false},
		{entity.SignupStageComplete, entity.SignupStageEmailVerify, false},
		{entity.SignupStageRole, entity.SignupStageAuthSetup, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.from)+"->"+string(tt.to), func(t *testing.T) {
			result := tt.from.CanAdvanceTo(tt.to)
			if result != tt.expected {
				t.Errorf("CanAdvanceTo(%q, %q) = %v, want %v", tt.from, tt.to, result, tt.expected)
			}
		})
	}
}

func TestConstants(t *testing.T) {
	if entity.KYCMaxAttempts != 3 {
		t.Errorf("KYCMaxAttempts = %d, want 3", entity.KYCMaxAttempts)
	}
	if entity.AccessTokenTTLMinutes != 15 {
		t.Errorf("AccessTokenTTLMinutes = %d, want 15", entity.AccessTokenTTLMinutes)
	}
	if entity.RefreshTokenTTLDays != 30 {
		t.Errorf("RefreshTokenTTLDays = %d, want 30", entity.RefreshTokenTTLDays)
	}
	// SEBI thresholds
	if entity.SEBINetWorthThresholdPaise != 2_00_00_000_00 {
		t.Errorf("SEBI NetWorth = %d", entity.SEBINetWorthThresholdPaise)
	}
}
