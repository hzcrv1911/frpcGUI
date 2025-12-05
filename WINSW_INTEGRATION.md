# WinSW 集成说明

本文档说明了如何使用 WinSW 集成来运行 frpc.exe 作为 Windows 服务，而不是内嵌 frpc 代码。

## 概述

修改后的 frpcgui 不再内嵌 frpc 代码,而是使用独立的 frpc.exe 和 WinSW 来管理 Windows 服务。这种方式有以下优势:

1. **更轻量级的应用程序**：不再需要内嵌整个 frpc 代码库
2. **更容易更新 frpc**：只需替换 frpc.exe 文件即可
3. **更好的隔离性**：frpc 运行在独立的进程中
4. **兼容性更好**：使用官方 frpc.exe，确保完全兼容

## 部署要求

在使用修改后的 frpcgui 之前,需要准备以下文件:

1. **frpcgui.exe**:主应用程序
2. **frpc.exe**：FRP 客户端可执行文件
3. **winsw.exe** 或 **winsw-x64.exe**：WinSW 可执行文件

这些文件应该放在同一个目录中。

## 配置文件

配置文件格式保持不变，支持两种格式：

- **TOML 格式**（推荐）：新格式，更灵活
- **INI 格式**：旧格式，向后兼容

配置文件由 GUI 生成，与独立 frpc.exe 完全兼容。

## 服务管理

### 安装服务

1. 在 GUI 中创建或导入配置
2. 点击"启动"按钮
3. 系统会自动：
   - 复制frpc.exe到配置文件夹
   - 生成 WinSW 配置文件
   - 使用 WinSW 安装服务
   - 启动服务

### 卸载服务

1. 在 GUI 中点击"停止"按钮
2. 系统会自动：
   - 停止服务
   - 使用 WinSW 卸载服务
   - 删除相关配置文件
   - 删除配置文件夹里面的frpc.exe

### 重载配置

1. 修改配置后
2. 系统会自动重启服务以应用新配置

## 日志管理

日志文件存储在配置文件所在目录的`logs`子目录中：

- **frpc.log**：主日志文件
- **frpc.YYYYMMDD-HHMMSS.log**：历史日志文件

## 故障排除

### 常见问题

1. **WinSW 未找到**

   - 错误：`WinSW executable not found`
   - 解决：确保 winsw.exe 或 winsw-x64.exe 在应用程序目录中

2. **frpc.exe 未找到**

   - 错误：`frpc.exe not found`
   - 解决：确保 frpc.exe 在应用程序目录中

3. **服务启动失败**
   - 检查配置文件是否正确
   - 检查日志文件中的错误信息
   - 确保端口未被占用

### 手动管理服务

如果需要手动管理服务，可以使用以下命令：

```bash
# 安装服务
winsw.exe install config.xml

# 启动服务
winsw.exe start servicename

# 停止服务
winsw.exe stop servicename

# 重启服务
winsw.exe restart servicename

# 卸载服务
winsw.exe uninstall servicename

# 查看服务状态
winsw.exe status servicename
```

## 技术细节

### WinSW 配置文件

WinSW 配置文件自动生成，包含以下内容：

- 服务 ID 和名称
- 可执行文件路径（frpc.exe）
- 启动参数（-c 配置文件路径）
- 日志配置
- 停止命令

### 服务命名规则

服务名称基于配置文件路径的 MD5 哈希生成，确保唯一性：

```
frpcgui_<配置文件路径的MD5哈希>
```

### 状态跟踪

服务状态通过定期查询 WinSW 管理的服务来获取，而不是通过 Windows 服务 API。
