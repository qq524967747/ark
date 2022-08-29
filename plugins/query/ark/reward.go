package ark

import (
	"math/rand"
	"time"
)

func GrabReward(key int, uin int64) int {
	reward := RewardList[key]
	if reward.RemainCount <= 0 {
		panic("RemainCount <= 0")
	}
	//最后一个
	if reward.RemainCount-1 == 0 {
		money := reward.RemainMoney
		reward.RemainCount = 0
		reward.RemainMoney = 0
		return money
	}
	//是否可以直接0.01
	if (reward.RemainMoney / reward.RemainCount) == 1 {
		money := 1
		reward.RemainMoney -= money
		reward.RemainCount--
		return money
	}

	//红包算法参考 https://www.zhihu.com/question/22625187
	//最大可领金额 = 剩余金额的平均值x2 = (剩余金额 / 剩余数量) * 2
	//领取金额范围 = 0.01 ~ 最大可领金额
	maxMoney := int(reward.RemainMoney/reward.RemainCount) * 2
	rand.Seed(time.Now().UnixNano())
	money := rand.Intn(maxMoney)
	for money == 0 {
		//防止零
		money = rand.Intn(maxMoney)
	}
	reward.RemainMoney -= money
	//防止剩余金额负数
	if reward.RemainMoney < 0 {
		money += reward.RemainMoney
		reward.RemainMoney = 0
		reward.RemainCount = 0
	} else {
		reward.RemainCount--
	}
	reward.UserList = append(reward.UserList, uin)
	RewardList[key] = reward
	return money
}

// 发红包
func GenerateReward(count, money int) int {

	reward := Reward{Count: count, Money: money,
		RemainCount: count, RemainMoney: money}

	var max int
	for k, _ := range RewardList {
		for k1, _ := range RewardList {
			if k1 > k {
				max = k1
			} else {
				max = k
			}
		}
	}
	key := max + 1
	RewardList[key] = reward
	return key
}

// 抢红包
func RobReward(id int, uin int64) (point int, reward Reward) {
	reward = RewardList[id]
	if reward.RemainCount > 0 {
		point = GrabReward(id, uin)
		return
	}
	return 0, reward
}
