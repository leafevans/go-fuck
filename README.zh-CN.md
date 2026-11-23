# go-fuck

[English](README.md) | **简体中文**

[thefuck](https://github.com/nvbn/thefuck) 的高性能 Go 重写版。go-fuck 会根据上一条失败命令及其
终端输出给出修正建议；确认后，由 shell 函数执行选中的正确命令。

**项目状态：** go-fuck 自 **3.0.0** 起进入稳定阶段。设置文件和 `--json` 输出受兼容性承诺保护，
详见 [UPGRADING.md](UPGRADING.md)。

> 本页是中文快速上手与实践指南。英文 [README.md](README.md) 是完整、权威的用户文档，包含全部
> 设置项、JSON 契约和自动生成的完整内置规则列表。为避免两份内容长期不一致，本页不复制这些
> 大段参考表。

```console
$ git push
fatal: The current branch main has no upstream branch.

$ fuck
› git push --set-upstream origin main  [enter · ↑↓ · ctrl+c]
```

## 快速安装

| 平台 | 推荐方式 | 命令 |
| --- | --- | --- |
| Windows | Scoop | `scoop bucket add go-fuck https://github.com/leafevans/go-fuck`<br>`scoop install go-fuck` |
| macOS / Linux | Homebrew | `brew tap leafevans/go-fuck https://github.com/leafevans/go-fuck`<br>`brew trust leafevans/go-fuck`<br>`brew install go-fuck` |
| Arch Linux | AUR | `yay -S go-fuck` 或 `paru -S go-fuck` |
| 已安装 Go 1.27+ | `go install` | `go install github.com/leafevans/go-fuck/v3/cmd/go-fuck@latest` |

也可以从 [GitHub Releases](https://github.com/leafevans/go-fuck/releases) 下载预编译包。发布页同时
提供 `checksums.txt`；安装前建议校验下载文件。预编译二进制没有运行时依赖。

## 初始化 shell

安装后运行：

```console
go-fuck --setup
```

setup 会检测当前 shell，将 alias 写入对应配置文件，并打印重新加载命令。完成后重新打开终端，或
执行 setup 输出的 reload 命令。

支持的 shell：

- bash、zsh、fish、tcsh
- Windows PowerShell 5.1 和 PowerShell 7
- Windows 的 `cmd.exe` 可使用 release 包中的 `scripts/fuck.bat`

默认的 `capture_mode=auto` 会在支持的 shell 中读取上一条命令的终端输出，避免重复执行已失败的
命令。当前支持无需重复执行的组合是：

- Linux / macOS 上的 bash、zsh
- Windows 上的 PowerShell，使用 ConPTY，需要 Windows 10 1809 或更高版本

其他 shell 会自动回退到 `rerun`，也就是在受控超时内重新执行失败命令以取得输出。

## 基本用法

命令失败后输入：

```console
fuck
```

使用方向键选择建议，按 Enter 执行，按 Ctrl+C 或 `q` 取消。常用选项：

| 选项 | 作用 |
| --- | --- |
| `fuck -y` | 直接执行第一条建议，不再确认 |
| `fuck -r` | 修正后的命令仍失败时再次调用 go-fuck |
| `fuck --explain` | 显示捕获来源、命中的规则、副作用和执行计划，不执行修正 |
| `fuck --preview` | 只预览建议，不执行原命令、不应用修正、不写 shell 历史 |
| `fuck --json` | 输出结构化 JSON，适合编辑器、脚本和 AI 工具调用 |
| `fuck --doctor` | 检查 shell、alias、设置文件、历史记录和可选集成 |
| `fuck --list-rules` | 列出可用于 `rules`、`exclude_rules` 和 `priority` 的规则名 |
| `fuck -v` | 显示版本和构建信息 |
| `fuck -h` | 显示完整帮助 |
| `fuck -a` | 输出可由 shell 直接求值的 alias 定义 |

`--json`、`--list-rules --json` 和 `--doctor --json` 都是机器可读输出，不会附加颜色或装饰。
检查某条命令而不重新执行它：

```console
go-fuck --json --no-rerun --force-command "git brnch"
```

如果调用方已经拿到了失败输出，可通过文件或标准输入传入：

```console
go-fuck --json --force-command "git push" --output-from failed-output.log
cat failed-output.log | go-fuck --json --force-command "git push" --output-from -
```

## 终端捕获策略

`settings.json` 中的 `capture_mode` 控制输出采集方式：

| 模式 | 行为 |
| --- | --- |
| `auto` | 优先读取终端日志；不可用时重新执行命令 |
| `instant` | 只读取终端日志，绝不自动重新执行命令 |
| `rerun` | 不读取终端日志，重新执行上一条失败命令 |

常用初始化方式：

```bash
# bash / zsh
eval "$(go-fuck --alias --capture-mode=auto)"

# fish
go-fuck --alias --capture-mode=auto | source

# tcsh
eval `go-fuck --alias --capture-mode=auto`
```

```powershell
# PowerShell
iex "$(go-fuck --alias --capture-mode=auto)"
```

如果重复执行命令可能产生副作用，可显式使用 `--capture-mode=instant`；在不支持终端捕获的环境中，
这时不会自动回退到 `rerun`。

## 设置

设置文件由首次运行自动创建：

| 平台 | 路径 |
| --- | --- |
| Linux / BSD | `~/.config/go-fuck/settings.json` 或 `$XDG_CONFIG_HOME/go-fuck/settings.json` |
| macOS | `~/Library/Application Support/go-fuck/settings.json` |
| Windows | `%AppData%\go-fuck\settings.json` |

最小示例：

```json
{
    "capture_mode": "auto",
    "require_confirmation": true,
    "wait_command": 10,
    "exclude_rules": ["git_push"],
    "priority": {
        "cd_mkdir": 100
    }
}
```

所有设置都可用 `GOFUCK_*` 环境变量覆盖，并兼容旧版 `THEFUCK_*` 前缀。优先级从低到高依次为：
内置默认值、`settings.json`、环境变量、命令行选项。

完整键名、默认值和环境变量映射见英文 README 的
[Settings](README.md#settings) 章节。

## 常见问题

### `fuck` 没有反应

先运行：

```console
go-fuck --doctor
```

重点检查当前 shell、alias 是否已写入正确配置文件、`go-fuck` 实际从哪个路径启动，以及历史记录
是否可读。PowerShell 会检查 `powershell.exe` 和 `pwsh` 的 profile。

### 提示 `Your shell alias is outdated`

这表示当前终端里加载的 `fuck` 函数比已安装的二进制旧。基本纠错可能仍能工作，但多行命令和
`--list-rules`、`--json` 等信息命令的透传可能不完整。无需重装，在当前终端重新加载 alias：

```bash
# bash / zsh
eval "$(go-fuck --alias)"

# fish
go-fuck --alias | source

# tcsh
eval `go-fuck --alias`
```

```powershell
# PowerShell
iex "$(go-fuck --alias)"
```

裸 `--alias` 会遵循当前配置的 `capture_mode`。需要跳过终端捕获时，可以在命令后加
`--capture-mode=rerun`。

### `fuck -v` 显示的版本不对

通常是 `PATH` 中还有另一个更早安装的 `go-fuck`。运行：

```console
go-fuck --debug --version
```

它会在调试输出中显示实际可执行文件路径和构建模块。v3 应显示
`github.com/leafevans/go-fuck/v3`。在 WSL 中应优先使用 Linux 二进制，而不是从 Windows
interop `PATH` 中发现的 `.exe`。

### `git brnch` 没有建议 `git branch`

可以使用强制命令和已保存的失败输出复现，不执行原命令：

```console
go-fuck --explain --capture-mode=instant --force-command "git brnch"
go-fuck --json --no-rerun --force-command "git brnch"
```

如果仍未匹配，请提交 issue，并附上操作系统、shell、`go-fuck --version`、失败命令输出和
`--debug` 信息。粘贴前请移除密码、令牌和私有主机名。

## 自定义规则

无需重新编译的方式是在配置目录的 `rules/` 下放置 JSON 文件：

| 平台 | 目录 |
| --- | --- |
| Linux / BSD | `~/.config/go-fuck/rules` |
| macOS | `~/Library/Application Support/go-fuck/rules` |
| Windows | `%AppData%\go-fuck\rules` |

示例：

```json
{
    "name": "git_remote_typo",
    "enabled": true,
    "priority": 900,
    "match": {
        "script": "^git remote add (\\w+)$",
        "output": "No such remote '([^']+)'"
    },
    "commands": ["git remote remove {1}"]
}
```

每次调用都会重新读取规则目录，因此新增规则后不需要重启。JSON 规则是纯声明式的，不执行代码，
不能访问网络或产生副作用。需要检查文件、运行外部命令或贡献给主项目的规则应编译进二进制；
接口和测试约定见 [PORTING.md](PORTING.md)。

## 与 thefuck 的主要差异

- 单一静态二进制，无 Python 解释器或运行时依赖。
- Linux、macOS、Windows 的 amd64/arm64 发布预编译版本。
- bash、zsh 和 Windows PowerShell 支持无需重复执行的终端捕获。
- 设置使用 `settings.json`，同时兼容 `GOFUCK_*` 和旧 `THEFUCK_*` 环境变量。
- 自定义规则支持编译内置和 JSON 声明式扩展。
- 提供稳定的 `--json`、`--explain`、`--preview`、`--doctor` 和 `--list-rules` 接口。

完整差异和兼容性边界见英文 README 的
[Differences from thefuck](README.md#differences-from-thefuck)。

## 文档与安全

| 文档 | 内容 |
| --- | --- |
| [README.md](README.md) | 完整英文文档、全部设置和自动生成的规则列表 |
| [UPGRADING.md](UPGRADING.md) | 兼容性承诺和版本迁移 |
| [CONTRIBUTING.md](CONTRIBUTING.md) | 开发、测试和提交流程 |
| [PORTING.md](PORTING.md) | 从 thefuck 移植内置规则 |
| [packaging/README.md](packaging/README.md) | Homebrew、Scoop、AUR 和发布流程 |
| [SECURITY.md](SECURITY.md) | 漏洞报告范围和数据处理说明 |

发现安全问题请不要创建公开 issue，使用
[GitHub 私有漏洞报告](https://github.com/leafevans/go-fuck/security/advisories/new)。

## 构建与测试

从源码构建需要 Go 1.27 或更高版本：

```bash
git clone https://github.com/leafevans/go-fuck.git
cd go-fuck
go build ./...
go test ./... -count=1
go vet ./...
```

项目只使用 Go 标准库。便捷任务和完整检查流程见 [CONTRIBUTING.md](CONTRIBUTING.md)。

## 许可证

MIT，详见 [LICENSE](LICENSE)。go-fuck 是独立项目，与 thefuck 官方无隶属关系。
