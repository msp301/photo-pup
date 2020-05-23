package photoslibrary

type AlbumsList struct {
	Albums        []Album
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
