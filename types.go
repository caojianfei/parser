package videosdk

import (
	"context"
	"time"
)

// Platform 平台类型
type Platform string

const (
	PlatformDouyin      Platform = "douyin"      // 抖音
	PlatformKuaishou    Platform = "kuaishou"    // 快手
	PlatformXiaohongshu Platform = "xiaohongshu" // 小红书
	PlatformBilibili    Platform = "bilibili"    // B站（预留）
	PlatformYoutube     Platform = "youtube"     // YouTube（预留）
)

// VideoType 视频类型
type VideoType string

const (
	VideoTypeVideo   VideoType = "video" // 视频
	VideoTypeImage   VideoType = "image" // 图文
	VideoTypeLive    VideoType = "live"  // 实况
	VideoTypeUnknown VideoType = "unknown"
)

// MediaType 媒体类型
type MediaType string

const (
	MediaTypeVideo MediaType = "video" // 视频文件
	MediaTypeImage MediaType = "image" // 图片文件
	MediaTypeGif   MediaType = "gif"   // 动图文件
)

// DownloadItem 下载项
type DownloadItem struct {
	URL  string    `json:"url"`  // 下载链接
	Type MediaType `json:"type"` // 媒体类型
}

// ParseRequest 解析请求参数
type ParseRequest struct {
	Platform Platform `json:"platform"` // 平台
	VideoID  string   `json:"video_id"` // 视频ID
	URL      string   `json:"url"`      // 视频URL（可选，用于从URL提取ID）
	Cookie   string   `json:"cookie"`   // Cookie（某些平台需要）
	Proxy    string   `json:"proxy"`    // 代理地址（可选）
	Source   bool     `json:"source"`   // 是否获取原始数据
}

// VideoInfo 统一的视频信息结构
type VideoInfo struct {
	// 基础信息
	ID          string    `json:"id"`          // 视频ID
	Title       string    `json:"title"`       // 视频标题
	Description string    `json:"description"` // 视频描述
	Type        VideoType `json:"type"`        // 视频类型
	Platform    Platform  `json:"platform"`    // 平台
	URL         string    `json:"url"`         // 视频页面URL
	CreateTime  time.Time `json:"create_time"` // 创建时间
	Duration    string    `json:"duration"`    // 视频时长

	// 媒体信息
	Downloads []DownloadItem `json:"downloads"` // 媒体下载链接列表
	CoverURL  string         `json:"cover_url"` // 封面图片URL
	Width     int            `json:"width"`     // 视频宽度
	Height    int            `json:"height"`    // 视频高度

	// 作者信息
	Author AuthorInfo `json:"author"` // 作者信息

	// 统计信息
	Stats VideoStats `json:"stats"` // 统计数据

	// 音乐信息
	Music MusicInfo `json:"music"` // 音乐信息

	// 标签信息
	Tags []string `json:"tags"` // 标签列表

	// 扩展信息
	Extra map[string]interface{} `json:"extra"` // 平台特有的扩展信息
}

// AuthorInfo 作者信息
type AuthorInfo struct {
	UID       string `json:"uid"`       // 用户ID
	SecUID    string `json:"sec_uid"`   // 安全用户ID
	UniqueID  string `json:"unique_id"` // 唯一ID
	Nickname  string `json:"nickname"`  // 昵称
	Avatar    string `json:"avatar"`    // 头像URL
	Signature string `json:"signature"` // 个人签名
	Age       int    `json:"age"`       // 年龄
}

// VideoStats 视频统计信息
type VideoStats struct {
	PlayCount    int64 `json:"play_count"`    // 播放量
	LikeCount    int64 `json:"like_count"`    // 点赞数
	CommentCount int64 `json:"comment_count"` // 评论数
	ShareCount   int64 `json:"share_count"`   // 分享数
	CollectCount int64 `json:"collect_count"` // 收藏数
}

// MusicInfo 音乐信息
type MusicInfo struct {
	ID     string `json:"id"`     // 音乐ID
	Title  string `json:"title"`  // 音乐标题
	Author string `json:"author"` // 音乐作者
	URL    string `json:"url"`    // 音乐URL
}

// ParseResponse 解析响应
type ParseResponse struct {
	Success bool       `json:"success"`         // 是否成功
	Message string     `json:"message"`         // 响应消息
	Data    *VideoInfo `json:"data,omitempty"`  // 视频信息
	Error   string     `json:"error,omitempty"` // 错误信息
	Time    time.Time  `json:"time"`            // 响应时间
}

// Parser 平台解析器接口
type Parser interface {
	// GetPlatform 获取平台类型
	GetPlatform() Platform

	// ParseVideo 解析视频信息
	ParseVideo(ctx context.Context, req *ParseRequest) (*VideoInfo, error)

	// ExtractVideoID 从URL提取视频ID
	ExtractVideoID(url string) (string, error)

	// ValidateRequest 验证请求参数
	ValidateRequest(req *ParseRequest) error
}

// SDK 主SDK接口
type SDK interface {
	// RegisterParser 注册平台解析器
	RegisterParser(parser Parser) error

	// ParseVideo 解析视频信息
	ParseVideo(ctx context.Context, req *ParseRequest) (*ParseResponse, error)

	// GetSupportedPlatforms 获取支持的平台列表
	GetSupportedPlatforms() []Platform

	// SetTimeout 设置请求超时时间
	SetTimeout(timeout time.Duration)

	// SetUserAgent 设置User-Agent
	SetUserAgent(userAgent string)
}
