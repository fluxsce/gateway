package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"
	"syscall"

	"gateway/pkg/config"
	"gateway/pkg/security"

	"golang.org/x/term"
)

const (
	version = "1.0.0"
	banner  = `
╔═══════════════════════════════════════════════════════════╗
║          Gateway 密码加密工具 (Password Encryptor)        ║
║                      Version %s                           ║
╚═══════════════════════════════════════════════════════════╝
`
)

func main() {
	var (
		password     = flag.String("p", "", "待加密的密码（如果不提供，将交互式输入）")
		key          = flag.String("k", "", "加密密钥（如果不提供，使用配置文件中的默认密钥）")
		useRandomKey = flag.Bool("r", false, "使用随机生成的密钥加密（会同时输出密钥和密文）")
		decrypt      = flag.Bool("d", false, "解密模式（默认为加密模式）")
		ciphertext   = flag.String("c", "", "待解密的密文（解密模式必填）")
		showHelp     = flag.Bool("h", false, "显示帮助信息")
		showVersion  = flag.Bool("v", false, "显示版本信息")
		generateKey  = flag.Bool("g", false, "生成新的随机密钥")
		configDir    = flag.String("config", "./configs", "配置文件目录")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, banner, version)
		fmt.Fprintf(os.Stderr, "\n用法: %s [选项]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "选项:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\n示例:\n")
		fmt.Fprintf(os.Stderr, "  # 交互式加密密码（使用默认密钥）\n")
		fmt.Fprintf(os.Stderr, "  %s\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # 使用命令行参数加密密码\n")
		fmt.Fprintf(os.Stderr, "  %s -p \"my-password\"\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # 使用指定密钥加密\n")
		fmt.Fprintf(os.Stderr, "  %s -p \"my-password\" -k \"my-secret-key\"\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # 使用随机生成的密钥加密（会输出密钥和密文）\n")
		fmt.Fprintf(os.Stderr, "  %s -p \"my-password\" -r\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # 解密密码\n")
		fmt.Fprintf(os.Stderr, "  %s -d -c \"ENCY_AQAM...\"\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # 生成新的随机密钥\n")
		fmt.Fprintf(os.Stderr, "  %s -g\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # 从环境变量读取密码（Linux/Mac）\n")
		fmt.Fprintf(os.Stderr, "  echo \"my-password\" | %s\n\n", os.Args[0])
	}

	flag.Parse()

	// 显示版本信息
	if *showVersion {
		fmt.Printf("Gateway 密码加密工具 v%s\n", version)
		os.Exit(0)
	}

	// 显示帮助信息
	if *showHelp {
		flag.Usage()
		os.Exit(0)
	}

	// 生成新密钥
	if *generateKey {
		generateNewKey()
		waitBeforeExit()
		os.Exit(0)
	}

	// 加载配置（如果需要使用默认密钥）
	if err := config.LoadConfig(*configDir); err != nil {
		fmt.Fprintf(os.Stderr, "警告: 加载配置文件失败: %v\n", err)
		fmt.Fprintf(os.Stderr, "将使用硬编码的默认密钥\n")
	}

	// 解密模式
	if *decrypt {
		decryptPassword(*ciphertext, *key)
		waitBeforeExit()
		return
	}

	// 如果没有任何参数（除了程序名），显示交互式菜单
	if len(flag.Args()) == 0 && *password == "" && !*useRandomKey {
		interactiveMenu()
		waitBeforeExit()
		return
	}

	// 加密模式
	var plaintext string
	if *password != "" {
		plaintext = *password
	} else {
		// 交互式输入
		var err error
		plaintext, err = readPassword("请输入密码: ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: 读取密码失败: %v\n", err)
			os.Exit(1)
		}
		if plaintext == "" {
			fmt.Fprintf(os.Stderr, "错误: 密码不能为空\n")
			os.Exit(1)
		}
	}

	// 如果使用随机密钥，先生成密钥
	if *useRandomKey {
		randomKey, err := security.GenerateSecretKey()
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: 生成随机密钥失败: %v\n", err)
			waitBeforeExit()
			os.Exit(1)
		}
		encryptPasswordWithRandomKey(plaintext, randomKey)
	} else {
		encryptPassword(plaintext, *key)
	}
	// 命令行模式不需要等待，交互式模式已在 interactiveMenu 中处理
}

// encryptPassword 加密密码
func encryptPassword(plaintext, secretKey string) {
	var ciphertext string
	var err error

	if secretKey != "" {
		// 使用指定的密钥
		fmt.Printf("使用指定的密钥进行加密...\n")
		ciphertext, err = security.AESEncryptToString(secretKey, plaintext)
	} else {
		// 使用默认密钥
		fmt.Printf("使用默认密钥进行加密（从配置文件读取）...\n")
		ciphertext, err = security.EncryptWithDefaultKey(plaintext)
		fmt.Printf("提示: 默认密钥已从配置文件读取\n")
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 加密失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("加密成功！")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Printf("\n密文: %s\n\n", ciphertext)
	fmt.Println("使用说明:")
	fmt.Println("  1. 将上述密文复制到配置文件或数据库中")
	fmt.Println("  2. 使用 -d -c 参数可以解密验证")
	fmt.Println("  3. 示例解密命令:")
	fmt.Printf("     %s -d -c \"%s\"\n", os.Args[0], ciphertext)
	fmt.Println()
}

// encryptPasswordWithRandomKey 使用随机密钥加密密码
func encryptPasswordWithRandomKey(plaintext, randomKey string) {
	fmt.Printf("正在生成随机密钥并加密...\n")

	ciphertext, err := security.AESEncryptToString(randomKey, plaintext)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 加密失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("加密成功！")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Printf("\n随机密钥: %s\n", randomKey)
	fmt.Printf("密文: %s\n\n", ciphertext)
	fmt.Println("使用说明:")
	fmt.Println("  1. 请妥善保管上述密钥和密文")
	fmt.Println("  2. 将密文复制到配置文件或数据库中")
	fmt.Println("  3. 保存密钥用于后续解密（可配置到 app.yaml 或其他安全位置）")
	fmt.Println("  4. 使用指定密钥解密:")
	fmt.Printf("     %s -d -c \"%s\" -k \"%s\"\n", os.Args[0], ciphertext, randomKey)
	fmt.Println("  5. 如果密钥已配置到 app.yaml 的 app.encryption_key，可直接解密:")
	fmt.Printf("     %s -d -c \"%s\"\n", os.Args[0], ciphertext)
	fmt.Println()
}

// decryptPassword 解密密码
func decryptPassword(ciphertext, secretKey string) {
	if ciphertext == "" {
		// 交互式输入密文
		var err error
		ciphertext, err = readInput("请输入密文: ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: 读取密文失败: %v\n", err)
			os.Exit(1)
		}
		if ciphertext == "" {
			fmt.Fprintf(os.Stderr, "错误: 密文不能为空\n")
			os.Exit(1)
		}
	}

	var plaintext string
	var err error

	if secretKey != "" {
		// 使用指定的密钥
		fmt.Printf("使用指定的密钥进行解密...\n")
		plaintext, err = security.AESDecryptFromString(secretKey, ciphertext)
	} else {
		// 使用默认密钥
		fmt.Printf("使用默认密钥进行解密（从配置文件读取）...\n")
		plaintext, err = security.DecryptWithDefaultKey(ciphertext)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 解密失败: %v\n", err)
		fmt.Fprintf(os.Stderr, "提示: 请检查密文是否正确，或确认使用的密钥是否正确\n")
		os.Exit(1)
	}

	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("解密成功！")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Printf("\n明文: %s\n\n", plaintext)
}

// generateNewKey 生成新的随机密钥
func generateNewKey() {
	key, err := security.GenerateSecretKey()
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 生成密钥失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("新密钥生成成功！")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Printf("\n密钥: %s\n\n", key)
	fmt.Println("使用说明:")
	fmt.Println("  1. 将上述密钥配置到 app.yaml 的 app.encryption_key 中")
	fmt.Println("  2. 或通过环境变量 GATEWAY_APP_ENCRYPTION_KEY 设置")
	fmt.Println("  3. 示例配置文件内容:")
	fmt.Println("     app:")
	fmt.Println("       encryption_key: \"" + key + "\"")
	fmt.Println()
}

// readPassword 交互式读取密码（隐藏输入）
func readPassword(prompt string) (string, error) {
	fmt.Print(prompt)

	// 检查是否为终端
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		// 非终端环境（如管道），使用普通读取
		reader := bufio.NewReader(os.Stdin)
		password, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(password), nil
	}

	// 终端环境，隐藏输入
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}
	fmt.Println() // 换行
	return string(bytePassword), nil
}

// readInput 交互式读取普通输入
func readInput(prompt string) (string, error) {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(input), nil
}

// waitBeforeExit 在程序退出前等待（仅 Windows 下双击运行时）
func waitBeforeExit() {
	// 只在 Windows 平台下检查
	if runtime.GOOS != "windows" {
		return
	}

	// 检查是否有交互式终端（从命令行运行）
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return
	}

	// 检查是否从命令行运行（通过检查父进程判断）
	// 简单方法：如果没有重定向输出，说明可能是双击运行
	// 更可靠的方法：检查是否在控制台窗口运行

	fmt.Print("\n按回车键退出...")
	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')
}

// interactiveMenu 交互式菜单
func interactiveMenu() {
	fmt.Printf(banner, version)
	fmt.Println("\n欢迎使用 Gateway 密码加密工具！")
	fmt.Println("\n请选择操作：")
	fmt.Println("  1. 加密密码（使用默认密钥）")
	fmt.Println("  2. 加密密码（使用随机生成的密钥）")
	fmt.Println("  3. 解密密码")
	fmt.Println("  4. 生成新的随机密钥")
	fmt.Println("  5. 显示帮助信息")
	fmt.Println("  0. 退出")
	fmt.Println()

	choice, err := readInput("请输入选项 [0-5]: ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 读取输入失败: %v\n", err)
		os.Exit(1)
	}

	choice = strings.TrimSpace(choice)
	switch choice {
	case "1":
		// 加密密码（使用默认密钥）
		plaintext, err := readPassword("请输入密码: ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: 读取密码失败: %v\n", err)
			os.Exit(1)
		}
		if plaintext == "" {
			fmt.Fprintf(os.Stderr, "错误: 密码不能为空\n")
			os.Exit(1)
		}
		encryptPassword(plaintext, "")

	case "2":
		// 加密密码（使用随机密钥）
		plaintext, err := readPassword("请输入密码: ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: 读取密码失败: %v\n", err)
			os.Exit(1)
		}
		if plaintext == "" {
			fmt.Fprintf(os.Stderr, "错误: 密码不能为空\n")
			os.Exit(1)
		}
		randomKey, err := security.GenerateSecretKey()
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: 生成随机密钥失败: %v\n", err)
			os.Exit(1)
		}
		encryptPasswordWithRandomKey(plaintext, randomKey)

	case "3":
		// 解密密码
		decryptPassword("", "")

	case "4":
		// 生成新密钥
		generateNewKey()

	case "5":
		// 显示帮助
		flag.Usage()
		waitBeforeExit()
		return

	case "0":
		fmt.Println("再见！")
		return

	default:
		fmt.Fprintf(os.Stderr, "错误: 无效的选项，请输入 0-5\n")
		return
	}
}
