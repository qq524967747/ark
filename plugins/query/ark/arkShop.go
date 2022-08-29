package ark

import (
	"fmt"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/niuhuan/mirai-framework"
	"log"
	"math/rand"
	"qq-bot/config"
	"qq-bot/database/db"
	"qq-bot/database/redis"
	"strconv"
	"strings"
	"time"
)

const id = "ARK_SHOP"
const name = "ark商店"

func NewPluginInstance() *mirai.Plugin {
	return &mirai.Plugin{
		Id: func() string {
			return id
		},
		Name: func() string {
			return name
		},
		OnMessage: func(client *mirai.Client, messageInterface interface{}) bool {
			content := client.MessageContent(messageInterface)
			if content == name {
				client.ReplyText(messageInterface,
					"在群中发送以下内容 \n\n bdsteam \n 新手礼包 \n 签到 \n 积分\\我的金币\\金币\\我的积分 \n 发红包 \n 抢红包 \n 打劫@一个人 \n 下线 \n")
				return true
			}
			return false
		},
		OnGroupMessage: func(client *mirai.Client, groupMessage *message.GroupMessage) bool {
			content := client.MessageContent(groupMessage)
			content = strings.TrimSpace(content)
			if "新手礼包" == content {
				kits(client, groupMessage)
				return true
			}
			if "签到" == content {
				sign(client, groupMessage)
				return true
			}

			if ("宁龟王" == content || strings.Contains(content, "龟")) && groupMessage.Sender.Uin != 1039426060 {
				client.ReplyText(groupMessage, "你才是龟王")
				return true
			}
			if "积分" == content || "我的金币" == content || "金币" == content || "我的积分" == content {
				points(client, groupMessage)
				return true
			}
			if strings.HasPrefix(content, "发红包") {
				generateReward(client, content, groupMessage)

				return true
			}

			if strings.HasPrefix(content, "抢红包") {
				robReward(client, content, groupMessage)

				return true
			}

			if strings.HasPrefix(content, "修改权限") {
				robReward(client, content, groupMessage)

				return true
			}

			if "下线" == content {
				// todo
				return true
			}

			if strings.HasPrefix(content, "bdsteam") {
				bdsteam(client, content, groupMessage)
				return true
			}

			if strings.HasPrefix(content, "打劫") {
				rob(client, groupMessage)
				return true
			}
			return false
		},
	}
}

// 获取积分
func points(client *mirai.Client, groupMessage *message.GroupMessage) {
	points := loadPoint(groupMessage.GroupCode, groupMessage.Sender.Uin)
	client.ReplyText(groupMessage, fmt.Sprintf("积分合计 : %v", points.Points))
}

// 获取新手礼包
func kits(client *mirai.Client, groupMessage *message.GroupMessage) {
	var user ArkShopPlayer
	err := db.Client().Where("qq = ?", groupMessage.Sender.Uin).Find(&user).Error
	if err != nil {
		log.Println(err.Error())
		return
	}
	if user.Id == 0 {
		client.ReplyText(groupMessage, "您还未绑定QQ，输入: bdsteam steamid \n 或者游戏中输入 /bdqq QQ号")
		return
	}
	if !user.HasNewKits {
		client.ReplyText(groupMessage, "您已经领过新手礼包了")
		return
	}
	lock, err := redis.TryLock(fmt.Sprintf("ARK::LOCK::%v", groupMessage.GroupCode), time.Second*5, time.Second*15)
	if err != nil {
		return
	}
	defer lock.Unlock()
	err = db.Client().Where("qq = ?", groupMessage.Sender.Uin).Update("points", user.Points+config.Config.Ark.KitsPoint).Error
	if err != nil {
		log.Println(err.Error())
		return
	}
	client.ReplyText(
		groupMessage,
		fmt.Sprintf(
			"领取成功 : \n"+
				" 获得积分 : %v \n"+
				" 积分合计 : %v \n\n",
			config.Config.Ark.KitsPoint, user.Points+config.Config.Ark.KitsPoint,
		),
	)
}

// 签到
func sign(client *mirai.Client, groupMessage *message.GroupMessage) {
	day := time.Now()
	dayStr := day.Format("2006-01-02")
	lock, err := redis.TryLock(fmt.Sprintf("ARK::LOCK::%v", groupMessage.Sender.Uin), time.Second*5, time.Second*15)
	if err != nil {
		return
	}
	defer lock.Unlock()
	user := loadPoint(groupMessage.GroupCode, groupMessage.Sender.Uin)
	if user.Id == 0 {
		client.ReplyText(groupMessage, "您还未绑定steamid，清先绑定")
		return
	}
	if user.Id != 0 && dayStr == user.LastSignTime {
		client.ReplyText(groupMessage, "您今天已经签到过")
		return
	}
	up := rand.Int()%15 + 15 // 基础积分15, 随机积分15
	saveLastSignTime(groupMessage.GroupCode, groupMessage.Sender.Uin, dayStr)
	incPoint(groupMessage.GroupCode, groupMessage.Sender.Uin, user.Points+up)
	client.ReplyText(
		groupMessage,
		fmt.Sprintf(
			"签到成功 : \n"+
				" 获得积分 : %v \n"+
				" 积分合计 : %v \n\n",
			up, user.Points+up,
		),
	)
}

// 打劫
func rob(client *mirai.Client, groupMessage *message.GroupMessage) {
	at := client.MessageFirstAt(groupMessage)
	if at == 0 {
		client.ReplyText(groupMessage, "您需要发送 打劫并@一个人 才能打劫他人积分")
		return
	}
	lock, err := redis.TryLock(fmt.Sprintf("ARK::LOCK::%v", groupMessage.Sender.Uin), time.Second*5, time.Second*15)
	if err != nil {
		return
	}
	defer lock.Unlock()
	timeKey := fmt.Sprintf("ARK::ROB::%v", groupMessage.Sender.Uin)
	_, err = redis.GetStringError(timeKey)
	if err == nil {
		client.ReplyText(groupMessage, "五分钟内只能打劫一次")
		return
	}
	if err == redis.Nil {
		srcPoints := loadPoint(groupMessage.GroupCode, groupMessage.Sender.Uin)
		dstPoints := loadPoint(groupMessage.GroupCode, at)
		if 1000 >= dstPoints.Points {
			client.ReplyText(groupMessage, "他已经没有钱可以被打劫了")
			return
		}
		if rand.Int()%100 < 10 {
			if redis.SetString(timeKey, "1", time.Second) {
				if srcPoints.Points < 1000 {
					srcPoints.Points = 1000
				}
				incPoint(groupMessage.GroupCode, groupMessage.Sender.Uin, srcPoints.Points-1000)
				client.ReplyText(
					groupMessage,
					fmt.Sprintf("打劫时被狗咬, 丢失 %v 积分, 积分合计 : %v", 100, srcPoints.Points-1000),
				)
			}
			return
		}
		if redis.SetString(timeKey, "1", time.Second) {
			inc := rand.Int() % 2000
			incPoint(groupMessage.GroupCode, at, dstPoints.Points-inc)
			incPoint(groupMessage.GroupCode, groupMessage.Sender.Uin, srcPoints.Points+inc)
			client.ReplyText(
				groupMessage,
				fmt.Sprintf("打劫到 %v 积分, 积分合计 : %v", inc, srcPoints.Points+inc),
			)
		}
	}
}

// 发红包
func generateReward(client *mirai.Client, content string, groupMessage *message.GroupMessage) {
	// todo  权限判断
	contentList := strings.Split(content, " ")
	if len(contentList) != 3 {
		client.ReplyText(groupMessage, "格式不对，注意空格，示例：发红包 10 100000")
		return
	}
	count, err := strconv.Atoi(contentList[1])
	if err != nil {
		client.ReplyText(groupMessage, "格式不对，注意空格，示例：发红包 10 100000")
		return
	}
	point, err := strconv.Atoi(contentList[2])
	if err != nil {
		client.ReplyText(groupMessage, "格式不对，注意空格，示例：发红包 10 100000")
		return
	}
	key := GenerateReward(count, point)
	client.ReplyText(groupMessage, fmt.Sprintf("红包数目：%d \n 红包数目：%d \n 剩余: %d \n 输入：抢红包 %d", point, count, count, key))
}

// 抢红包
func robReward(client *mirai.Client, content string, groupMessage *message.GroupMessage) {
	user := loadPoint(groupMessage.GroupCode, groupMessage.Sender.Uin)
	if user.Id == 0 {
		client.ReplyText(groupMessage, "您还未绑定steamid，清先绑定")
		return
	}
	contentList := strings.Split(content, " ")
	if len(contentList) != 2 {
		client.ReplyText(groupMessage, "格式不对，注意空格，示例：抢红包 10")
		return
	}
	key, err := strconv.Atoi(contentList[1])
	if err != nil {
		client.ReplyText(groupMessage, "格式不对，注意空格，示例：发红包 10 100000")
		return
	}
	if _, ok := RewardList[key]; !ok {
		client.ReplyText(groupMessage, "红包不存在")
		return
	}
	for _, v := range RewardList[key].UserList {
		if v == groupMessage.Sender.Uin {
			client.ReplyText(groupMessage, "您已经领取过该红包")
			return
		}
	}
	lock, err := redis.TryLock(fmt.Sprintf("ARK::ROBREWARD::%v", groupMessage.Sender.Uin), time.Second*5, time.Second*15)
	if err != nil {
		return
	}
	defer lock.Unlock()
	point, reward := RobReward(key, groupMessage.Sender.Uin)
	allPoint := user.Points + point
	incPoint(groupMessage.GroupCode, groupMessage.Sender.Uin, allPoint)
	client.ReplyText(groupMessage, fmt.Sprintf("恭喜，成功抢到：%d 积分 \n 积分总计：%d \n\n\n 红包总计：%d \n 红包剩余数目：%d \n 剩余积分: %d \n 输入：抢红包 %d", point, allPoint, reward.Money, RewardList[key].RemainCount, RewardList[key].RemainMoney, key))
}

// 绑定steamid
func bdsteam(client *mirai.Client, content string, groupMessage *message.GroupMessage) {
	contentList := strings.Split(content, " ")
	switch len(contentList) {
	case 1:
		client.ReplyText(groupMessage, "没有可用信息")
		return
	case 2:
		if contentList[1] == " " {
			client.ReplyText(groupMessage, "格式不对，注意空格，示例：只有steamid：bdsteam dsakj231kj43 \n 或者，有qq和steamid：bdsteam 12345678 dsakj231kj43")
			return
		}
		user := searchUserBySteam(groupMessage.GroupCode, contentList[1])
		if user.Id == 0 {
			client.ReplyText(groupMessage, "steamid 没有找到")
			return
		}
		if user.QQ != 0 {
			client.ReplyText(groupMessage, "您已经绑定QQ，无需再次绑定，如需更改QQ，请联系管理员")
			return
		}
		err := setQQ(groupMessage.GroupCode, groupMessage.Sender.Uin, contentList[1])
		if err != nil {
			client.ReplyText(groupMessage, "机器人出错了")
			return
		}
		client.ReplyText(groupMessage, "设置steamId成功")
		return
	case 3:
		if contentList[1] == " " || contentList[2] == " " {
			client.ReplyText(groupMessage, "格式不对，注意空格，示例：只有steamid：bdsteam dsakj231kj43 \n 或者，有qq和steamid：bdsteam 12345678 dsakj231kj43")
			return
		}
		uint, err := strconv.Atoi(contentList[1])
		if err != nil {
			client.ReplyText(groupMessage, "QQ号不正确")
			return
		}
		user := searchUserBySteam(groupMessage.GroupCode, contentList[2])
		if user.Id == 0 {
			client.ReplyText(groupMessage, "steamid 没有找到")
			return
		}
		if user.QQ != 0 {
			client.ReplyText(groupMessage, "您已经绑定QQ，无需再次绑定，如需更改QQ，请联系管理员")
			return
		}
		err = setQQ(groupMessage.GroupCode, int64(uint), contentList[2])
		if err != nil {
			client.ReplyText(groupMessage, "机器人出错了")
			return
		}
		client.ReplyText(groupMessage, "设置steamId成功")
	default:
		client.ReplyText(groupMessage, "格式不对，注意空格，示例：只有steamid：bdsteam dsakj231kj43 \n 或者，有qq和steamid：bdsteam 12345678 dsakj231kj43")
		return
	}
}

// 修改权限
func changeVIP(client *mirai.Client, content string, groupMessage *message.GroupMessage) {

}

func loadPoint(groupCode int64, uin int64) ArkShopPlayer {
	var user ArkShopPlayer
	loadDb := db.Client()
	/*if groupCode != 0 {
		loadDb = loadDb.Where("GroupCode = ?", groupCode)
	}*/
	if uin == 0 {
		return user
	}
	loadDb = loadDb.Where("QQ = ?", uin)
	err := loadDb.Find(&user).Error
	if err != nil {
		log.Println(err.Error())
	}
	return user
}

func searchUserBySteam(groupCode int64, steamId string) ArkShopPlayer {
	var user ArkShopPlayer
	loadDb := db.Client()
	/*if groupCode != 0 {
		loadDb = loadDb.Where("GroupCode = ?", groupCode)
	}*/
	loadDb = loadDb.Where("SteamId = ?", steamId)
	err := loadDb.Find(&user).Error
	if err != nil {
		log.Println(err.Error())
	}
	return user
}

func incPoint(groupCode int64, uin int64, inc int) {
	updateDb := db.Client()
	/*	if groupCode != 0 {
		updateDb = updateDb.Where("GroupCode = ?", groupCode)
	}*/
	if uin == 0 {
		log.Println("uin is nil")
		return
	}
	updateDb = updateDb.Model(&ArkShopPlayer{})
	updateDb = updateDb.Where("QQ = ?", uin)
	err := updateDb.Update("Points", inc).Error
	if err != nil {
		log.Println(err.Error())
		return
	}
}

func setQQ(groupCode int64, uin int64, steamId string) error {
	updateDb := db.Client()
	/*	if groupCode != 0 {
		updateDb = updateDb.Where("GroupCode = ?", groupCode)
	}*/
	updateDb.Debug()
	if uin == 0 {
		log.Println("uin is nil")
		return fmt.Errorf("uin is nil")
	}
	updateDb = updateDb.Model(&ArkShopPlayer{})
	updateDb = updateDb.Where("SteamId = ?", steamId)
	err := updateDb.Update("QQ", uin).Error
	if err != nil {
		log.Println(err.Error())
		return err
	}
	return nil
}

func saveLastSignTime(groupCode int64, uin int64, t string) error {
	updateDb := db.Client()
	/*if groupCode != 0 {
		updateDb = updateDb.Where("GroupCode = ?", groupCode)
	}*/
	updateDb.Debug()
	if uin == 0 {
		log.Println("uin is nil")
		return fmt.Errorf("uin is nil")
	}
	var user ArkShopPlayer
	user.QQ = uin
	user.LastSignTime = t
	updateDb = updateDb.Where("QQ = ?", uin)
	err := updateDb.Updates(&user).Error
	if err != nil {
		log.Println(err.Error())
		return err
	}

	return err
}
