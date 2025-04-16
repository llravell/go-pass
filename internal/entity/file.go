package entity

type FileUploadStatus string

const (
	FileUploadStatusPending FileUploadStatus = "pending"
	FileUploadStatusDone    FileUploadStatus = "done"
)

type File struct {
	Name        string
	Size        int64
	MinioBucket string
	Meta        string
}
