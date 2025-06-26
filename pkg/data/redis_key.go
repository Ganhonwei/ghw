package data

const (
	// 大富翁中奖牌
	SCRATCHTICKET_POKER = "scratchticket_poker"
	// 大富翁上次重置时间
	SCRATCHTICKET_RESET   = "scratchticket_reset"
	SCRATCHTICKET_VERSION = "scratchticket_version"

	CUSTOM_PHOTO = "custom_photo"
)

func GetCustomPhotoKey(userid string) string {
	return CUSTOM_PHOTO + "_" + userid
}
