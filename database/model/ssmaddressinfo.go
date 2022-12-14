package model

type SsmFingerprint String
type SsmChain String
type SsmHDPath String

type SsmAddressInfo struct {
	SsmAddressID SsmAddressID   `gorm:"unique_index;not null"`  // [FK] Reference to SsmAddress table
	Chain        SsmChain       `gorm:"index;not null;size:16"` // Ssm chain, non mutable
	Fingerprint  SsmFingerprint `gorm:"index;not null;size:8"`  // Ssm fingerprint, non mutable
	HDPath       SsmHDPath      `gorm:"index;not null;size:24"` // Ssm HDPath, non mutable
}

func (p *SsmAddressInfo) IsValid() bool {
	return p.SsmAddressID > 0 && len(p.Chain) > 0 && len(p.Fingerprint) > 0 && len(p.HDPath) > 0
}
