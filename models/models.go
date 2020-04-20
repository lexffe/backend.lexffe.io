package models

type ObjectType string
type AssetType string


const (
	TypeGenericPage ObjectType = "generic_page"
	TypePostPage    ObjectType = "post"
	TypeCVPage      ObjectType = "cv"
)

const (
	TypeGenericRef   ObjectType = "generic_reference"
	TypeProjectRef   ObjectType = "project"
	TypeHighlightRef ObjectType = "highlight"
	TypeAsset        ObjectType = "asset"
)

const (
	AssetImage AssetType = "image"
	AssetAudio AssetType = "audio"
	AssetVideo AssetType = "video"
)
