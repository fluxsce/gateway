package channel

import (
	"fmt"

	"gateway/pkg/alert"
)

// CreateChannel 根据配置创建告警渠道（工厂方法）
// 配置结构:
//
//	{
//	    "type": "email",           // 必需：告警类型 (email/qq/wechat_work)
//	    "name": "email-prod",      // 必需：渠道名称
//	    "server": {                // 必需：服务器配置
//	        "smtp_host": "smtp.example.com",
//	        "smtp_port": 587,
//	        ...
//	    },
//	    "send": {                  // 必需：发送配置
//	        "to": ["admin@example.com"],
//	        ...
//	    }
//	}
//
// 参数:
//
//	config: 配置映射，必须包含 type, name, server, send 字段
//
// 返回:
//
//	alert.Channel: 创建的渠道实例
//	error: 错误信息
//
// 示例:
//
//	config := map[string]interface{}{
//	    "type": "email",
//	    "name": "email-prod",
//	    "server": map[string]interface{}{
//	        "smtp_host": "smtp.gmail.com",
//	        "smtp_port": 587,
//	        "username": "alert@example.com",
//	        "password": "password123",
//	        "from": "alert@example.com",
//	        "use_tls": true,
//	    },
//	    "send": map[string]interface{}{
//	        "to": []string{"admin@example.com"},
//	    },
//	}
//	channel, err := CreateChannel(config)
func CreateChannel(config map[string]interface{}) (alert.Channel, error) {
	// 1. 解析告警类型
	typeStr, ok := config["type"].(string)
	if !ok {
		return nil, fmt.Errorf("配置缺少必需字段 'type' 或类型错误")
	}

	// 2. 解析渠道名称
	name, ok := config["name"].(string)
	if !ok {
		return nil, fmt.Errorf("配置缺少必需字段 'name' 或类型错误")
	}

	// 3. 解析服务器配置
	serverConfig, ok := config["server"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("配置缺少必需字段 'server' 或类型错误")
	}

	// 4. 解析发送配置
	sendConfig, ok := config["send"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("配置缺少必需字段 'send' 或类型错误")
	}

	// 5. 根据类型创建对应的渠道
	switch alert.AlertType(typeStr) {
	case alert.AlertTypeEmail:
		return createEmailChannel(name, serverConfig, sendConfig)

	case alert.AlertTypeQQ:
		return createQQChannel(name, serverConfig, sendConfig)

	case alert.AlertTypeWeChatWork:
		return createWeChatWorkChannel(name, serverConfig, sendConfig)

	default:
		return nil, fmt.Errorf("不支持的告警类型: %s", typeStr)
	}
}

// createEmailChannel 创建邮件告警渠道
func createEmailChannel(name string, serverConfig, sendConfig map[string]interface{}) (alert.Channel, error) {
	// 解析服务器配置
	srvCfg := &EmailServerConfig{}

	// 必需字段
	if host, ok := serverConfig["smtp_host"].(string); ok {
		srvCfg.SMTPHost = host
	} else {
		return nil, fmt.Errorf("邮件渠道配置缺少必需字段: server.smtp_host")
	}

	if port, ok := serverConfig["smtp_port"].(int); ok {
		srvCfg.SMTPPort = port
	} else if port, ok := serverConfig["smtp_port"].(float64); ok {
		srvCfg.SMTPPort = int(port)
	} else {
		return nil, fmt.Errorf("邮件渠道配置缺少必需字段: server.smtp_port")
	}

	if username, ok := serverConfig["username"].(string); ok {
		srvCfg.Username = username
	} else {
		return nil, fmt.Errorf("邮件渠道配置缺少必需字段: server.username")
	}

	if password, ok := serverConfig["password"].(string); ok {
		srvCfg.Password = password
	} else {
		return nil, fmt.Errorf("邮件渠道配置缺少必需字段: server.password")
	}

	if from, ok := serverConfig["from"].(string); ok {
		srvCfg.From = from
	} else {
		return nil, fmt.Errorf("邮件渠道配置缺少必需字段: server.from")
	}

	// 可选字段
	if fromName, ok := serverConfig["from_name"].(string); ok {
		srvCfg.FromName = fromName
	}

	if useTLS, ok := serverConfig["use_tls"].(bool); ok {
		srvCfg.UseTLS = useTLS
	}

	if skipVerify, ok := serverConfig["skip_verify"].(bool); ok {
		srvCfg.SkipVerify = skipVerify
	}

	if timeout, ok := serverConfig["timeout"].(int); ok {
		srvCfg.Timeout = timeout
	} else if timeout, ok := serverConfig["timeout"].(float64); ok {
		srvCfg.Timeout = int(timeout)
	}

	// 解析发送配置
	sendCfg := &EmailSendConfig{}

	// 必需字段 - to
	if to, ok := sendConfig["to"].([]string); ok {
		sendCfg.To = to
	} else if to, ok := sendConfig["to"].([]interface{}); ok {
		sendCfg.To = make([]string, 0, len(to))
		for _, v := range to {
			if str, ok := v.(string); ok {
				sendCfg.To = append(sendCfg.To, str)
			}
		}
	} else {
		return nil, fmt.Errorf("邮件渠道配置缺少必需字段: send.to")
	}

	// 可选字段
	if cc, ok := sendConfig["cc"].([]string); ok {
		sendCfg.CC = cc
	} else if cc, ok := sendConfig["cc"].([]interface{}); ok {
		sendCfg.CC = make([]string, 0, len(cc))
		for _, v := range cc {
			if str, ok := v.(string); ok {
				sendCfg.CC = append(sendCfg.CC, str)
			}
		}
	}

	if bcc, ok := sendConfig["bcc"].([]string); ok {
		sendCfg.BCC = bcc
	} else if bcc, ok := sendConfig["bcc"].([]interface{}); ok {
		sendCfg.BCC = make([]string, 0, len(bcc))
		for _, v := range bcc {
			if str, ok := v.(string); ok {
				sendCfg.BCC = append(sendCfg.BCC, str)
			}
		}
	}

	return NewEmailChannel(name, srvCfg, sendCfg)
}

// createQQChannel 创建QQ告警渠道
func createQQChannel(name string, serverConfig, sendConfig map[string]interface{}) (alert.Channel, error) {
	// 解析服务器配置
	srvCfg := &QQServerConfig{}

	// 必需字段
	if webhookURL, ok := serverConfig["webhook_url"].(string); ok {
		srvCfg.WebhookURL = webhookURL
	} else {
		return nil, fmt.Errorf("QQ渠道配置缺少必需字段: server.webhook_url")
	}

	// 可选字段
	if secret, ok := serverConfig["secret"].(string); ok {
		srvCfg.Secret = secret
	}

	if timeout, ok := serverConfig["timeout"].(int); ok {
		srvCfg.Timeout = timeout
	} else if timeout, ok := serverConfig["timeout"].(float64); ok {
		srvCfg.Timeout = int(timeout)
	}

	// 解析发送配置
	sendCfg := &QQSendConfig{}

	// 可选字段
	if atAll, ok := sendConfig["at_all"].(bool); ok {
		sendCfg.AtAll = atAll
	}

	if atUsers, ok := sendConfig["at_users"].([]string); ok {
		sendCfg.AtUsers = atUsers
	} else if atUsers, ok := sendConfig["at_users"].([]interface{}); ok {
		sendCfg.AtUsers = make([]string, 0, len(atUsers))
		for _, v := range atUsers {
			if str, ok := v.(string); ok {
				sendCfg.AtUsers = append(sendCfg.AtUsers, str)
			}
		}
	}

	return NewQQChannel(name, srvCfg, sendCfg)
}

// createWeChatWorkChannel 创建企业微信告警渠道
func createWeChatWorkChannel(name string, serverConfig, sendConfig map[string]interface{}) (alert.Channel, error) {
	// 解析服务器配置
	srvCfg := &WeChatWorkServerConfig{}

	// 必需字段
	if webhookURL, ok := serverConfig["webhook_url"].(string); ok {
		srvCfg.WebhookURL = webhookURL
	} else {
		return nil, fmt.Errorf("企业微信渠道配置缺少必需字段: server.webhook_url")
	}

	// 可选字段
	if secret, ok := serverConfig["secret"].(string); ok {
		srvCfg.Secret = secret
	}

	if msgType, ok := serverConfig["message_type"].(string); ok {
		srvCfg.MessageType = msgType
	} else {
		srvCfg.MessageType = "markdown" // 默认使用markdown
	}

	if timeout, ok := serverConfig["timeout"].(int); ok {
		srvCfg.Timeout = timeout
	} else if timeout, ok := serverConfig["timeout"].(float64); ok {
		srvCfg.Timeout = int(timeout)
	}

	// 解析发送配置
	sendCfg := &WeChatWorkSendConfig{}

	// 可选字段
	if mentionedList, ok := sendConfig["mentioned_list"].([]string); ok {
		sendCfg.MentionedList = mentionedList
	} else if mentionedList, ok := sendConfig["mentioned_list"].([]interface{}); ok {
		sendCfg.MentionedList = make([]string, 0, len(mentionedList))
		for _, v := range mentionedList {
			if str, ok := v.(string); ok {
				sendCfg.MentionedList = append(sendCfg.MentionedList, str)
			}
		}
	}

	if mentionedMobileList, ok := sendConfig["mentioned_mobile_list"].([]string); ok {
		sendCfg.MentionedMobileList = mentionedMobileList
	} else if mentionedMobileList, ok := sendConfig["mentioned_mobile_list"].([]interface{}); ok {
		sendCfg.MentionedMobileList = make([]string, 0, len(mentionedMobileList))
		for _, v := range mentionedMobileList {
			if str, ok := v.(string); ok {
				sendCfg.MentionedMobileList = append(sendCfg.MentionedMobileList, str)
			}
		}
	}

	return NewWeChatWorkChannel(name, srvCfg, sendCfg)
}

// CreateChannels 批量创建告警渠道
// 参数:
//
//	configs: 配置列表，每个配置包含 type, name, server, send 字段
//
// 返回:
//
//	[]alert.Channel: 成功创建的渠道列表
//	[]error: 每个配置对应的错误（成功为nil）
//
// 示例:
//
//	configs := []map[string]interface{}{
//	    {
//	        "type": "email",
//	        "name": "email-prod",
//	        "server": {...},
//	        "send": {...},
//	    },
//	    {
//	        "type": "qq",
//	        "name": "qq-ops",
//	        "server": {...},
//	        "send": {...},
//	    },
//	}
//	channels, errs := CreateChannels(configs)
func CreateChannels(configs []map[string]interface{}) ([]alert.Channel, []error) {
	channels := make([]alert.Channel, 0, len(configs))
	errors := make([]error, len(configs))

	for i, config := range configs {
		channel, err := CreateChannel(config)
		if err != nil {
			errors[i] = err
			continue
		}
		channels = append(channels, channel)
		errors[i] = nil
	}

	return channels, errors
}

// GetSupportedTypes 获取所有支持的告警类型
func GetSupportedTypes() []string {
	return []string{
		string(alert.AlertTypeEmail),
		string(alert.AlertTypeQQ),
		string(alert.AlertTypeWeChatWork),
	}
}
