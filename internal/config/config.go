package config

type Config struct {
	ReleaseDir   string       `json:"release_dir"`
	UploadDir    string       `json:"upload_dir"`
	UnpackScript string       `json:"unpack"`
	ReloadScript string       `json:"reload"`
	HealthScript string       `json:"health_check"`
	EmailConfig  *EmailConfig `json:"email_conf"`
}
