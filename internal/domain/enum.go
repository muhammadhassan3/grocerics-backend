package domain

type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

func (r Role) IsValid() bool {
	switch r {
	case RoleAdmin, RoleUser:
		return true
	}
	return false
}

type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusDisabled UserStatus = "disabled"
)

func (s UserStatus) IsValid() bool {
	switch s {
	case UserStatusActive, UserStatusDisabled:
		return true
	}
	return false
}

type Status string

const (
	StatusActive   Status = "active"
	StatusDisabled Status = "disabled"
)

func (s Status) IsValid() bool {
	switch s {
	case StatusActive, StatusDisabled:
		return true
	}
	return false
}

type OTPPurpose string

const (
	OTPPurposeLogin             OTPPurpose = "login"
	OTPPurposeRegister          OTPPurpose = "register"
	OTPPurposePhoneVerification OTPPurpose = "phone_verification"
)

func (p OTPPurpose) IsValid() bool {
	switch p {
	case OTPPurposeLogin, OTPPurposeRegister, OTPPurposePhoneVerification:
		return true
	}
	return false
}

// (e.g. "₹72/100gm") is computed from the value + unit.
type VolumeUnit string

const (
	VolumeUnitKg  VolumeUnit = "kg"
	VolumeUnitGm  VolumeUnit = "gm"
	VolumeUnitLtr VolumeUnit = "ltr"
	VolumeUnitMl  VolumeUnit = "ml"
	VolumeUnitPcs VolumeUnit = "pcs"
)

func (u VolumeUnit) IsValid() bool {
	switch u {
	case VolumeUnitKg, VolumeUnitGm, VolumeUnitLtr, VolumeUnitMl, VolumeUnitPcs:
		return true
	}
	return false
}

type PriceSource string

const (
	PriceSourceAPI    PriceSource = "api"
	PriceSourceManual PriceSource = "manual"
)

func (s PriceSource) IsValid() bool {
	switch s {
	case PriceSourceAPI, PriceSourceManual:
		return true
	}
	return false
}

type BannerTargetType string

const (
	BannerTargetNone     BannerTargetType = "none"
	BannerTargetCategory BannerTargetType = "category"
	BannerTargetProduct  BannerTargetType = "product"
	BannerTargetURL      BannerTargetType = "url"
)

func (t BannerTargetType) IsValid() bool {
	switch t {
	case BannerTargetNone, BannerTargetCategory, BannerTargetProduct, BannerTargetURL:
		return true
	}
	return false
}

type DevicePlatform string

const (
	DevicePlatformAndroid DevicePlatform = "android"
	DevicePlatformIOS     DevicePlatform = "ios"
)

func (p DevicePlatform) IsValid() bool {
	switch p {
	case DevicePlatformAndroid, DevicePlatformIOS:
		return true
	}
	return false
}
