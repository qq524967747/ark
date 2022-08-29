package ark

import (
	"encoding/json"
)

// 玩家
type ArkShopPlayer struct {
	Id           int             `gorm:"column:Id"`
	SteamId      int64           `gorm:"column:SteamId"`
	Kits         json.RawMessage `gorm:"column:Kits"`
	Points       int             `gorm:"column:Points"`
	TotalSpent   int             `gorm:"column:TotalSpent"`
	QQ           int64           `gorm:"column:QQ"`
	HasNewKits   bool            `gorm:"column:HasNewKits"`
	LastSignTime string          `gorm:"column:LastSignTime"`
	Permission   string          `gorm:"column:Permission"`
}

func (receiver ArkShopPlayer) TableName() string {
	return "ArkShopPlayers"
}

var RewardList = map[int]Reward{}

//红包
type Reward struct {
	Count       int     //个数
	Money       int     //总金额(分)
	RemainCount int     //剩余个数
	RemainMoney int     //剩余金额(分)
	UserList    []int64 //拆分列表
}
