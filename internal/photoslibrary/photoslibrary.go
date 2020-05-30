package photoslibrary

type AlbumsList struct {
	Albums        []Album
	NextPageToken string
}

type SharedAlbumsList struct {
	SharedAlbums  []Album
	NextPageToken string
}

type MediaItemsList struct {
	MediaItems    []MediaItem
	NextPageToken string
}

type Album struct {
	ID                    string
	Title                 string
	ProductURL            string
	IsWriteable           bool
	ShareInfo             ShareInfo
	MediaItemsCount       string
	CoverPhotoBaseURL     string
	CoverPhotoMediaItemID string
}

type MediaItem struct {
	ID            string
	Description   string
	ProductURL    string
	BaseURL       string
	MimeType      string
	MediaMetadata MediaMetadata
	Filename      string
}

type MediaMetadata struct {
	CreationTime string
	Width        string
	Height       string
	Photo        Photo
	Video        Video
}

type Photo struct {
	CameraMake      string
	CameraModel     string
	FocalLength     float32
	ApertureFNumber float32
	IsoEquivalent   int
	ExposureTime    string
}

type Video struct {
	CameraMake  string
	CameraModel string
	Fps         int
	Status      string
}

type SharedAlbumOptions struct {
	IsCollaborative bool
	IsCommentable   bool
}

type ShareInfo struct {
	SharedAlbumOptions SharedAlbumOptions
	ShareableURL       string
	ShareToken         string
	IsJoined           bool
	IsOwned            bool
}
