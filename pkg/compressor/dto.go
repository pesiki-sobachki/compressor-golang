package compressor

// OptionsDTO represents compression options exposed to package consumers.
type OptionsDTO struct {
	Format    string // "jpeg", "png", "webp"
	Quality   int    // 1–100
	MaxWidth  int    // optional, 0 = no limit
	MaxHeight int    // optional, 0 = no limit
}

// ResultDTO describes basic info about the compressed image.
type ResultDTO struct {
	MimeType string
	Size     int64
}
