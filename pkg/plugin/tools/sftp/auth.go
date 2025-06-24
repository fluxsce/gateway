package sftp

import (
	"fmt"
	"os"
	
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// createSSHConfig 创建SSH客户端配置
// 根据SFTP配置创建SSH连接所需的ClientConfig
// 包含认证方法、主机密钥验证、超时设置等
// 返回:
//   *ssh.ClientConfig: SSH客户端配置
//   error: 配置创建失败时的错误
func (c *sftpClient) createSSHConfig() (*ssh.ClientConfig, error) {
	// 创建基础SSH配置
	config := &ssh.ClientConfig{
		User:            c.config.Username,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 默认不验证主机密钥
		Timeout:         c.config.ConnectTimeout,
	}
	
	// 配置主机密钥验证
	if err := c.configureHostKeyVerification(config); err != nil {
		return nil, fmt.Errorf("配置主机密钥验证失败: %w", err)
	}
	
	// 配置认证方法
	authMethods, err := c.createAuthMethods()
	if err != nil {
		return nil, fmt.Errorf("创建认证方法失败: %w", err)
	}
	
	if len(authMethods) == 0 {
		return nil, fmt.Errorf("未配置任何有效的认证方法")
	}
	
	config.Auth = authMethods
	return config, nil
}

// configureHostKeyVerification 配置主机密钥验证
// 根据配置设置主机密钥验证策略
// 参数:
//   config: SSH客户端配置
// 返回:
//   error: 配置失败时的错误
func (c *sftpClient) configureHostKeyVerification(config *ssh.ClientConfig) error {
	// 如果未启用主机密钥验证，使用默认的不安全回调
	if c.config.HostKeyVerification == nil || !c.config.HostKeyVerification.Enabled {
		return nil
	}
	
	// 如果不要求严格的主机密钥检查，使用默认回调
	if !c.config.HostKeyVerification.StrictHostKeyChecking {
		return nil
	}
	
	// 优先使用已知主机文件
	if c.config.HostKeyVerification.KnownHostsFile != "" {
		hostKeyCallback, err := knownhosts.New(c.config.HostKeyVerification.KnownHostsFile)
		if err != nil {
			return fmt.Errorf("加载已知主机文件失败: %w", err)
		}
		config.HostKeyCallback = hostKeyCallback
		return nil
	}
	
	// 使用受信任的主机密钥列表
	if len(c.config.HostKeyVerification.TrustedHostKeys) > 0 {
		trustedKeys, err := c.parseTrustedHostKeys()
		if err != nil {
			return fmt.Errorf("解析受信任主机密钥失败: %w", err)
		}
		
		// 使用第一个受信任的密钥（可以扩展为支持多个）
		config.HostKeyCallback = ssh.FixedHostKey(trustedKeys[0])
		return nil
	}
	
	return fmt.Errorf("启用了严格主机密钥检查，但未提供已知主机文件或受信任密钥")
}

// parseTrustedHostKeys 解析受信任的主机密钥
// 将配置中的主机密钥数据解析为SSH公钥对象
// 返回:
//   []ssh.PublicKey: 解析后的公钥列表
//   error: 解析失败时的错误
func (c *sftpClient) parseTrustedHostKeys() ([]ssh.PublicKey, error) {
	var trustedKeys []ssh.PublicKey
	
	for i, keyData := range c.config.HostKeyVerification.TrustedHostKeys {
		key, err := ssh.ParsePublicKey(keyData)
		if err != nil {
			return nil, fmt.Errorf("解析第%d个受信任主机密钥失败: %w", i+1, err)
		}
		trustedKeys = append(trustedKeys, key)
	}
	
	return trustedKeys, nil
}

// createAuthMethods 创建SSH认证方法列表
// 根据配置创建支持的认证方法，支持密码认证和公钥认证
// 返回:
//   []ssh.AuthMethod: 认证方法列表
//   error: 创建失败时的错误
func (c *sftpClient) createAuthMethods() ([]ssh.AuthMethod, error) {
	var authMethods []ssh.AuthMethod
	
	// 添加公钥认证
	if c.config.PublicKeyAuth != nil {
		publicKeyAuth, err := c.createPublicKeyAuth()
		if err != nil {
			return nil, fmt.Errorf("创建公钥认证失败: %w", err)
		}
		if publicKeyAuth != nil {
			authMethods = append(authMethods, publicKeyAuth)
		}
	}
	
	// 添加密码认证
	if c.config.PasswordAuth != nil {
		passwordAuth := c.createPasswordAuth()
		if passwordAuth != nil {
			authMethods = append(authMethods, passwordAuth)
		}
	}
	
	// 添加键盘交互认证（如果支持）
	if c.config.KeyboardInteractiveAuth != nil {
		keyboardAuth := c.createKeyboardInteractiveAuth()
		if keyboardAuth != nil {
			authMethods = append(authMethods, keyboardAuth)
		}
	}
	
	return authMethods, nil
}

// createPublicKeyAuth 创建公钥认证方法
// 从配置中加载私钥并创建公钥认证
// 返回:
//   ssh.AuthMethod: 公钥认证方法
//   error: 创建失败时的错误
func (c *sftpClient) createPublicKeyAuth() (ssh.AuthMethod, error) {
	var keyData []byte
	var err error
	
	// 优先使用提供的私钥数据
	if len(c.config.PublicKeyAuth.PrivateKeyData) > 0 {
		keyData = c.config.PublicKeyAuth.PrivateKeyData
	} else if c.config.PublicKeyAuth.PrivateKeyPath != "" {
		// 从文件加载私钥
		keyData, err = os.ReadFile(c.config.PublicKeyAuth.PrivateKeyPath)
		if err != nil {
			return nil, fmt.Errorf("读取私钥文件失败: %w", err)
		}
	} else {
		// 没有提供私钥数据
		return nil, nil
	}
	
	// 解析私钥
	var signer ssh.Signer
	if c.config.PublicKeyAuth.Passphrase != "" {
		// 带密码的私钥
		signer, err = ssh.ParsePrivateKeyWithPassphrase(keyData, []byte(c.config.PublicKeyAuth.Passphrase))
		if err != nil {
			return nil, fmt.Errorf("解析带密码的私钥失败: %w", err)
		}
	} else {
		// 无密码的私钥
		signer, err = ssh.ParsePrivateKey(keyData)
		if err != nil {
			return nil, fmt.Errorf("解析私钥失败: %w", err)
		}
	}
	
	return ssh.PublicKeys(signer), nil
}

// createPasswordAuth 创建密码认证方法
// 根据配置创建密码认证
// 返回:
//   ssh.AuthMethod: 密码认证方法，如果未配置密码则返回nil
func (c *sftpClient) createPasswordAuth() ssh.AuthMethod {
	// 检查是否配置了密码或允许空密码
	if c.config.PasswordAuth.Password != "" || c.config.PasswordAuth.AllowEmptyPassword {
		return ssh.Password(c.config.PasswordAuth.Password)
	}
	return nil
}

// createKeyboardInteractiveAuth 创建键盘交互认证方法
// 用于支持一些需要交互式认证的场景
// 返回:
//   ssh.AuthMethod: 键盘交互认证方法
func (c *sftpClient) createKeyboardInteractiveAuth() ssh.AuthMethod {
	// 如果配置了键盘交互认证
	if c.config.KeyboardInteractiveAuth.Enabled {
		return ssh.KeyboardInteractive(func(user, instruction string, questions []string, echos []bool) ([]string, error) {
			// 简单实现：使用密码回答所有问题
			answers := make([]string, len(questions))
			for i := range questions {
				// 如果有预设的回答，使用预设回答
				if i < len(c.config.KeyboardInteractiveAuth.Answers) {
					answers[i] = c.config.KeyboardInteractiveAuth.Answers[i]
				} else if c.config.PasswordAuth != nil {
					// 否则尝试使用密码
					answers[i] = c.config.PasswordAuth.Password
				}
			}
			return answers, nil
		})
	}
	return nil
} 