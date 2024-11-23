package models

type PackageType int

const (
	PackageType_Package    = 0x01
	PackageType_Project    = 0x02
	PackageType_ExeProject = PackageType_Project | 0x04
	PackageType_Tool       = PackageType_Package | 0x08
	PackageType_Any        = PackageType_ExeProject | PackageType_Tool
)
