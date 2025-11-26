# FRPCGUI 便携版构建指南

本项目已简化为仅支持便携版构建，无需安装WiX或Windows SDK。

## 系统要求

### 必需
- **Go 1.24.0 或更高版本**
  - 下载地址: https://golang.org/dl/
  - 安装后确保 `go` 命令在PATH中可用

### 可选
- **MinGW (推荐)**
  - 下载地址: https://github.com/mstorsjo/llvm-mingw/releases
  - 用于编译C代码，某些功能可能需要

## 快速开始

### 1. 克隆项目
```bash
git clone https://github.com/hzcrv1911/frpcgui
cd frpcgui
```

### 2. 构建便携版
```bash
# 构建所有架构 (x64 和 x86)
build_portable.bat

# 或者只构建特定架构
build_portable.bat x64
build_portable.bat x86
```

### 3. 运行程序
```bash
# 直接运行
bin\x64\frpcgui.exe

# 或
bin\x86\frpcgui.exe
```

## 构建输出

构建完成后，您将在 `bin` 目录中找到：

- `bin\x64\frpcgui.exe` - 64位版本
- `bin\x86\frpcgui.exe` - 32位版本
- `bin\frpcgui-{version}-x64.zip` - 64位压缩包
- `bin\frpcgui-{version}-x86.zip` - 32位压缩包

## 分发

您可以直接分发：
1. 单个可执行文件 (`frpcgui.exe`)
2. 压缩包 (`frpcgui-{version}-*.zip`)

接收方无需安装任何依赖，可以直接运行。

## 故障排除

### 问题：找不到go命令
确保已安装Go 1.24.0+并将其添加到PATH环境变量。

### 问题：编译失败
1. 确保网络连接正常，需要下载依赖包
2. 检查Go版本是否为1.24.0或更高
3. 尝试运行 `go mod tidy` 更新依赖

### 问题：某些功能不工作
安装MinGW并确保 `gcc` 命令在PATH中可用。

## 与原版的区别

此简化版本：
- ✅ 移除了WiX相关代码
- ✅ 移除了Windows SDK依赖
- ✅ 移除了MSI安装包创建
- ✅ 保留了所有核心功能
- ✅ 保留了多语言支持

如果您需要创建Windows安装包，可以使用原版本的 `installer_old` 目录中的文件。

## 开发

如果您想修改代码并重新构建：

```bash
# 修改代码后重新构建
build_portable.bat

# 或者直接运行（不编译资源）
go run ./cmd/frpcgui