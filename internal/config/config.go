package config

type HTTP struct {
	Address         string `yaml:"address"`
	ReadTimeout     string `yaml:"read_timeout"`
	WriteTimeout    string `yaml:"write_timeout"`
	IdleTimeout     string `yaml:"idle_timeout"`
	MaxUploadSizeMB int64  `yaml:"max_upload_size_mb"`
}

func (h HTTP) MaxUploadSizeBytes() int64 {
	return h.MaxUploadSizeMB * 1024 * 1024
}

type Storage struct {
	Path             string `yaml:"path"`
	CompressedSubdir string `yaml:"compressed_subdir"`
	TmpSubdir        string `yaml:"tmp_subdir"`
}

type Log struct {
	Level      string `yaml:"level"`
	Format     string `yaml:"format"`
	WithCaller bool   `yaml:"with_caller"`
}

type Image struct {
	DefaultFormat  string   `yaml:"default_format"`
	DefaultQuality int      `yaml:"default_quality"`
	MaxWidth       int      `yaml:"max_width"`
	MaxHeight      int      `yaml:"max_height"`
	AllowFormats   []string `yaml:"allow_formats"`
}

type Config struct {
	HTTP    HTTP    `yaml:"http"`
	Storage Storage `yaml:"storage"`
	Log     Log     `yaml:"log"`
	Image   Image   `yaml:"image"`
}
