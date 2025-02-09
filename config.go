package main

type Language int32

type FileFormat int32

const (
	PNG  FileFormat = iota
	JPG             = iota
	TIFF            = iota
	BMP             = iota
)

func (f FileFormat) String() string {
	return [...]string{"png", "jpg", "tiff", "bmp"}[int32(f)]
}

type Theme int32

const (
	ThemeLight Theme = iota
	ThemeDark        = iota
)

type Font int32

const (
	FontDefault     Font = iota
	FontBerkleyMono      = iota
	FontArial            = iota
	FontComicSans        = iota
	FontZapfino          = iota
	FontSpleen           = iota
	FontCount            = iota
)

type Config struct {
	Language          Language
	FileFormat        FileFormat
	ActiveFormatIndex int32
	CurrentTheme      Theme
	CurrentFont       Font
	FontSize          int64
}

func (c *Config) GetActiveFileFormat() FileFormat {
	return FileFormat(c.ActiveFormatIndex)
}

func NewConfig() Config {
	return Config{Language: English, FileFormat: TIFF, CurrentTheme: ThemeLight, CurrentFont: FontZapfino, FontSize: 18}
}
