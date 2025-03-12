package pkg

type S3Base struct {
	FilesName []string
	BktName   string
	GpGpass   string
}

type UpdateIndex struct {
	Index []string
	Map   map[string]interface{}
}

var S3Index = "index"
