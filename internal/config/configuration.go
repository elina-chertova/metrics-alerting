package config

type Settings struct {
	IsCompress      bool
	IsSendBatch     bool
	FlagContentType string
}

func NewSettings() *Settings {
	return &Settings{
		IsCompress:      true,
		IsSendBatch:     false,
		FlagContentType: "application/json",
	}
}
