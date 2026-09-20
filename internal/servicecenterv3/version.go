package servicecenterv3

// Version 是本包实现版本，遵循 semver。-dev 表示尚未冻结默认引擎切换。
const Version = "v3.0.0-dev"

const (
	// APILevelExperimental 表示契约可加字段/方法，但不得静默改语义。
	APILevelExperimental = "experimental"
	// APILevelStable 预留给默认引擎切到本包、并删除旧树之后。
	APILevelStable = "stable"
)

// APILevel 当前对外承诺等级。大厂评审与开源声明以此为准，勿与 Version 混用。
const APILevel = APILevelExperimental

// DefaultBootstrapTenant 是 StartAll 在配置未给租户时的引导租户。
// 仅用于进程启动加载中心实例，不是数据面客户端可传的租户。
const DefaultBootstrapTenant = "default"
