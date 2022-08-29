package plugins

import (
	"github.com/niuhuan/mirai-framework"
	"qq-bot/plugins/games/farm"
	"qq-bot/plugins/query/ark"
	"qq-bot/plugins/query/imglab"
	"qq-bot/plugins/sys/ignore"
	"qq-bot/plugins/sys/log"
	"qq-bot/plugins/sys/menu"
	"qq-bot/plugins/tools/gm"
)

func Register(c *mirai.Client) {
	// 事件监听器
	actionsListeners := []*mirai.ActionListener{
		log.NewListenerInstance(),
	}
	c.SetActionListeners(actionsListeners)
	ark.InitDB()
	// 自定义组件
	cPlugins := []*mirai.Plugin{
		gm.NewPluginInstance(),
		//circle.NewPluginInstance(),
		ark.NewPluginInstance(),
		imglab.NewPluginInstance(),
		farm.NewPluginInstance(),
		// 最后可以添加拦截所有私聊并做出回复的插件, 做一个连天系统
	}
	// 系统组件
	sPlugins := []*mirai.Plugin{
		log.NewPluginInstance(),
		// 忽略指定用户的发言 所以放在菜单的钱main
		ignore.NewPluginInstance(),
		menu.NewPluginInstance(cPlugins),
	}
	c.SetPlugins(
		append(sPlugins, cPlugins...),
	)
	// 插件过滤器 true为阻止该插件
	c.SetPluginBlocker(func(plugin *mirai.Plugin, contactType int, contactNumber int64) bool {
		return false
	})
}
