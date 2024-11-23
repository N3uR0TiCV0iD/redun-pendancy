package models

type LoadStatus int

const (
	LoadStatus_None    = 0
	LoadStatus_Loaded  = 1
	LoadStatus_Skipped = -1
)
