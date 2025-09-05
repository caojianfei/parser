# 跨平台视频数据解析SDK (Go版本)

这是一个专注于获取单个作品数据的跨平台视频数据解析SDK，支持抖音、快手、小红书等多个平台的视频数据解析。

## 特性

- 🚀 **模块化设计**: 采用插件式架构，便于扩展新平台
- 🔧 **统一接口**: 提供标准化的API接口，屏蔽平台差异
- 📱 **多平台支持**: 支持抖音、快手、小红书等主流平台
- ⚡ **高性能**: 基于Go语言，支持并发处理
- 🛡️ **类型安全**: 完整的类型定义，减少运行时错误
- 🔄 **易扩展**: 新增平台只需实现Parser接口
- 📦 **混合媒体**: 统一的Downloads字段支持视频和图片混合内容
- 🎯 **智能解析**: 抖音三步流程，快手小红书直接URL解析

## 支持的平台

| 平台 | 状态 | 说明 |
|------|------|------|
| 抖音 | ✅ 已实现 | 支持视频和图文解析 |
| 快手 | ✅ 已实现 | 支持视频解析，需要Cookie |
| 小红书 | ✅ 已实现 | 支持视频和图文解析，需要Cookie |
| B站 | 📋 计划中 | 后续版本支持 |
| YouTube | 📋 计划中 | 后续版本支持 |

## 安装

```bash
go mod init your-project
go get github.com/caojianfei/parser
```

## 快速开始

### 基础用法

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    videosdk "github.com/resdownload/video-parser-sdk"
    "github.com/resdownload/video-parser-sdk/parsers"
)

func main() {
    // 创建SDK实例
    sdk := videosdk.NewSDK()
    
    // 注册抖音解析器
    douyinParser := parsers.NewDouyinParser("http://localhost:5555")
    sdk.RegisterParser(douyinParser)
    
    // 注册快手解析器
    kuaishouParser := parsers.NewKuaishouParser("http://localhost:5557")
    sdk.RegisterParser(kuaishouParser)
    
    // 注册小红书解析器
    xiaohongshuParser := parsers.NewXiaohongshuParser("http://localhost:5556")
    sdk.RegisterParser(xiaohongshuParser)
    
    // 解析视频
    req := &videosdk.ParseRequest{
        Platform: videosdk.PlatformDouyin,
        VideoID:  "7535398881601817894",
        Cookie:   "your_cookie_here",
    }
    
    resp, err := sdk.ParseVideo(context.Background(), req)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("视频标题: %s\n", resp.Data.Title)
    fmt.Printf("作者: %s\n", resp.Data.Author.Nickname)
    fmt.Printf("播放量: %d\n", resp.Data.Stats.PlayCount)
    
    // 获取下载链接
    for i, download := range resp.Data.Downloads {
        fmt.Printf("下载链接[%d]: %s (类型: %s)\n", i+1, download.URL, download.Type)
    }
}
```

### 多平台解析示例

#### 抖音解析（三步流程）

抖音解析采用三步流程：分享链接解析 → 完整URL获取 → 作品ID提取 → 视频数据获取

```go
// 方式1：使用分享短链接
req := &videosdk.ParseRequest{
    Platform: videosdk.PlatformDouyin,
    URL:      "https://v.douyin.com/iFRMqmyv/", // 分享短链接
    Cookie:   "your_douyin_cookie",
}

// 方式2：使用完整URL
req := &videosdk.ParseRequest{
    Platform: videosdk.PlatformDouyin,
    URL:      "https://www.douyin.com/video/7535398881601817894",
    Cookie:   "your_douyin_cookie",
}

// 方式3：直接使用作品ID
req := &videosdk.ParseRequest{
    Platform: videosdk.PlatformDouyin,
    VideoID:  "7535398881601817894",
    Cookie:   "your_douyin_cookie",
}

resp, err := sdk.ParseVideo(context.Background(), req)
if err != nil {
    log.Fatal(err)
}

// 获取下载链接
for i, download := range resp.Data.Downloads {
    fmt.Printf("下载链接[%d]: %s (类型: %s)\n", i+1, download.URL, download.Type)
}
```

#### 快手解析（直接URL解析）

```go
// 解析快手视频（需要Cookie和API服务）
req := &videosdk.ParseRequest{
    Platform: videosdk.PlatformKuaishou,
    URL:      "https://v.kuaishou.com/3xMsre", // 快手分享链接
    Cookie:   "your_kuaishou_cookie_here",
    Proxy:    "", // 可选
}

resp, err := sdk.ParseVideo(context.Background(), req)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("视频标题: %s\n", resp.Data.Title)
fmt.Printf("播放量: %d\n", resp.Data.Stats.PlayCount)

// 获取下载链接
for i, download := range resp.Data.Downloads {
    fmt.Printf("下载链接[%d]: %s (类型: %s)\n", i+1, download.URL, download.Type)
}
```

#### 小红书解析（直接URL解析）

```go
// 解析小红书内容（需要Cookie和API服务）
req := &videosdk.ParseRequest{
    Platform: videosdk.PlatformXiaohongshu,
    URL:      "https://www.xiaohongshu.com/explore/65e6b4b3000000001203e5b7", // 小红书作品链接
    Cookie:   "your_xiaohongshu_cookie_here",
    Proxy:    "", // 可选
}

resp, err := sdk.ParseVideo(context.Background(), req)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("作品标题: %s\n", resp.Data.Title)
fmt.Printf("作品类型: %s\n", resp.Data.Type)
fmt.Printf("点赞数: %d\n", resp.Data.Stats.LikeCount)
fmt.Printf("标签: %v\n", resp.Data.Tags)

// 获取下载链接（支持视频和图片混合内容）
for i, download := range resp.Data.Downloads {
    fmt.Printf("下载链接[%d]: %s (类型: %s)\n", i+1, download.URL, download.Type)
}
```

## 架构设计

### 核心组件

```
┌─────────────────┐
│   SDK Interface │  ← 统一的SDK接口
└─────────────────┘
         │
┌─────────────────┐
│   VideoSDK      │  ← SDK主实现类
└─────────────────┘
         │
┌─────────────────┐
│ Parser Registry │  ← 解析器注册中心
└─────────────────┘
         │
    ┌────┴────┐
    │ Parsers │  ← 各平台解析器
    └─────────┘
    │    │    │
 ┌──▼─┐ ┌▼─┐ ┌▼──┐
 │抖音│ │快手│ │小红书│
 └────┘ └──┘ └───┘
```

### 接口设计

#### Parser接口

每个平台解析器都需要实现以下接口：

```go
type Parser interface {
    GetPlatform() Platform
    ParseVideo(ctx context.Context, req *ParseRequest) (*VideoInfo, error)
    ExtractVideoID(url string) (string, error)
    ValidateRequest(req *ParseRequest) error
}
```

#### 数据结构

```go
type VideoInfo struct {
    ID          string         `json:"id"`
    Title       string         `json:"title"`
    Description string         `json:"description"`
    Type        VideoType      `json:"type"`
    Platform    Platform       `json:"platform"`
    URL         string         `json:"url"`
    CreateTime  time.Time      `json:"create_time"`
    Duration    string         `json:"duration"`
    Downloads   []DownloadItem `json:"downloads"`
    CoverURL    string         `json:"cover_url"`
    Author      AuthorInfo     `json:"author"`
    Stats       VideoStats     `json:"stats"`
    Music       MusicInfo      `json:"music"`
    Tags        []string       `json:"tags"`
    Extra       map[string]interface{} `json:"extra"`
}

type DownloadItem struct {
    URL  string    `json:"url"`
    Type MediaType `json:"type"`
}

type MediaType string

const (
    MediaTypeVideo MediaType = "video"
    MediaTypeImage MediaType = "image"
)
```

## 扩展新平台

要添加新平台支持，只需要实现`Parser`接口：

```go
package parsers

import (
    "context"
    videosdk "github.com/resdownload/video-parser-sdk"
)

type NewPlatformParser struct {
    // 平台特有的配置
}

func NewNewPlatformParser() videosdk.Parser {
    return &NewPlatformParser{}
}

func (p *NewPlatformParser) GetPlatform() videosdk.Platform {
    return "new_platform"
}

func (p *NewPlatformParser) ParseVideo(ctx context.Context, req *videosdk.ParseRequest) (*videosdk.VideoInfo, error) {
    // 实现解析逻辑
    return nil, nil
}

func (p *NewPlatformParser) ExtractVideoID(url string) (string, error) {
    // 实现URL解析逻辑
    return "", nil
}

func (p *NewPlatformParser) ValidateRequest(req *videosdk.ParseRequest) error {
    // 实现参数验证逻辑
    return nil
}
```

然后注册到SDK：

```go
sdk := videosdk.NewSDK()
newParser := parsers.NewNewPlatformParser()
sdk.RegisterParser(newParser)
```

## 配置选项

### 超时设置

```go
sdk.SetTimeout(30 * time.Second)
```

### User-Agent设置

```go
sdk.SetUserAgent("VideoParser-SDK/1.0")
```

### API服务配置

对于快手和小红书平台，需要启动对应的API服务：

```bash
# 启动快手API服务（端口5555）
# 启动小红书API服务（端口5556）
```

### Cookie配置

快手和小红书平台需要提供有效的Cookie：

```go
request := &videosdk.ParseRequest{
    Platform: videosdk.PlatformKuaishou,
    URL:      "https://v.kuaishou.com/example",
    Cookie:   "your_cookie_here", // 从浏览器获取
    Proxy:    "", // 可选的代理设置
}
```

### 获取支持的平台

```go
platforms := sdk.GetSupportedPlatforms()
for _, platform := range platforms {
    fmt.Printf("支持平台: %s\n", platform)
}
```

## 错误处理

SDK提供了详细的错误信息：

```go
resp, err := sdk.ParseVideo(ctx, req)
if err != nil {
    fmt.Printf("解析失败: %v\n", err)
    return
}

if !resp.Success {
    fmt.Printf("解析失败: %s\n", resp.Error)
    return
}

// 使用解析结果
videoInfo := resp.Data
```

## 依赖项

- `github.com/go-resty/resty/v2`: HTTP客户端
- `github.com/tidwall/gjson`: JSON解析

## 项目结构

```
sdk/go/
├── go.mod              # Go模块定义
├── README.md           # 项目文档
├── types.go            # 类型定义
├── sdk.go              # SDK主实现
├── parsers/            # 解析器目录
│   ├── douyin.go       # 抖音解析器
│   ├── kuaishou.go     # 快手解析器（预留）
│   └── xiaohongshu.go  # 小红书解析器（预留）
└── example/            # 示例代码
    └── main.go         # 使用示例
```

## 许可证

MIT License

## 贡献

欢迎提交Issue和Pull Request来帮助改进这个项目。

## 更新日志

### v1.2.0 (2024-01-XX)

- ✅ 重构VideoInfo数据结构，使用Downloads字段替代VideoURL和Images
- ✅ 支持混合媒体内容（视频+图片）的统一下载链接管理
- ✅ 抖音解析器实现三步流程：短链接解析→完整URL→作品ID→视频数据
- ✅ 快手和小红书解析器优化为直接URL解析
- ✅ 新增MediaType枚举，支持video和image类型区分
- ✅ 完善示例代码，展示Downloads字段的使用方式
- ✅ 更新文档，提供详细的平台解析说明

### v1.1.0 (2024-01-XX)

- ✅ 实现快手平台解析器（基于API接口）
- ✅ 实现小红书平台解析器（基于API接口）
- ✅ 支持Cookie和代理配置
- ✅ 完善多平台解析示例
- ✅ 更新文档和使用说明

### v1.0.0
- ✅ 初始版本发布
- ✅ 实现抖音平台解析器
- ✅ 定义统一接口规范
- ✅ 提供完整的示例代码
- ✅ 建立模块化架构设计