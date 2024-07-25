package model

type ItemID string

func (i ItemID) String() string {
	return string(i)
}

type DfItemID string

func (i DfItemID) String() string {
	return string(i)
}

type BannerID string

func (i BannerID) String() string {
	return string(i)
}
