package parsers

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	videosdk "github.com/resdownload/video-parser-sdk"
	"github.com/tidwall/gjson"
)

// DouyinParser 抖音解析器
type DouyinParser struct {
	client  *resty.Client
	baseURL string
}

// NewDouyinParser 创建抖音解析器
func NewDouyinParser(baseURL string) videosdk.Parser {
	client := resty.New()
	client.SetTimeout(30 * time.Second)
	client.SetHeader("Content-Type", "application/json")
	client.SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	return &DouyinParser{
		client:  client,
		baseURL: baseURL,
	}
}

// GetPlatform 获取平台类型
func (p *DouyinParser) GetPlatform() videosdk.Platform {
	return videosdk.PlatformDouyin
}

// ExtractVideoID 从URL提取视频ID
func (p *DouyinParser) ExtractVideoID(url string) (string, error) {
	// 支持多种抖音URL格式
	patterns := []string{
		`https?://www\.douyin\.com/video/(\d+)`,
		`https?://www\.iesdouyin\.com/share/video/(\d+)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(url)
		if len(matches) > 1 {
			return matches[1], nil
		}
	}

	return "", fmt.Errorf("无法从URL中提取视频ID: %s", url)
}

// resolveShortURL 解析短链接获取完整URL
func (p *DouyinParser) resolveShortURL(shortURL string, proxy string) (string, error) {
	req := map[string]interface{}{
		"text":  shortURL,
		"proxy": proxy,
	}

	resp, err := p.client.R().
		SetBody(req).
		Post(p.baseURL + "/douyin/share")

	if err != nil {
		return "", fmt.Errorf("请求分享链接解析失败: %w", err)
	}

	if resp.StatusCode() != 200 {
		return "", fmt.Errorf("分享链接解析请求失败，状态码: %d", resp.StatusCode())
	}

	// 解析响应
	result := gjson.ParseBytes(resp.Body())
	if !result.Get("url").Exists() {
		return "", fmt.Errorf("分享链接解析响应中未找到URL")
	}

	return result.Get("url").String(), nil
}

// ValidateRequest 验证请求参数
func (p *DouyinParser) ValidateRequest(req *videosdk.ParseRequest) error {
	if req.VideoID == "" && req.URL == "" {
		return fmt.Errorf("video_id 或 url 至少需要提供一个")
	}

	if req.Platform != videosdk.PlatformDouyin {
		return fmt.Errorf("平台类型不匹配，期望: %s，实际: %s", videosdk.PlatformDouyin, req.Platform)
	}

	return nil
}

// ParseVideo 解析视频信息
func (p *DouyinParser) ParseVideo(ctx context.Context, req *videosdk.ParseRequest) (*videosdk.VideoInfo, error) {
	var videoID string
	var err error

	// 步骤1: 如果提供的是URL，需要先获取视频ID
	if req.URL != "" {
		// 检查是否为短链接
		if strings.Contains(req.URL, "v.douyin.com") {
			// 步骤1a: 解析短链接获取完整URL
			fullURL, err := p.resolveShortURL(req.URL, req.Proxy)
			if err != nil {
				return nil, fmt.Errorf("解析短链接失败: %w", err)
			}
			// 步骤1b: 从完整URL提取视频ID
			videoID, err = p.ExtractVideoID(fullURL)
			if err != nil {
				return nil, fmt.Errorf("从完整URL提取视频ID失败: %w", err)
			}
		} else {
			// 直接从完整URL提取视频ID
			videoID, err = p.ExtractVideoID(req.URL)
			if err != nil {
				return nil, fmt.Errorf("从URL提取视频ID失败: %w", err)
			}
		}
	} else if req.VideoID != "" {
		// 直接使用提供的视频ID
		videoID = req.VideoID
	} else {
		return nil, fmt.Errorf("必须提供URL或VideoID")
	}

	// 步骤2: 使用视频ID获取详细数据
	requestBody := map[string]interface{}{
		"detail_id": videoID,
		"cookie":    req.Cookie,
		"proxy":     req.Proxy,
		"source":    req.Source,
	}

	// 发送请求到 /douyin/detail 接口
	resp, err := p.client.R().
		SetContext(ctx).
		SetBody(requestBody).
		Post(p.baseURL + "/douyin/detail")

	if err != nil {
		return nil, fmt.Errorf("请求抖音API失败: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("抖音API请求失败，状态码: %d", resp.StatusCode())
	}

	// 解析响应
	result := gjson.ParseBytes(resp.Body())
	if !result.Get("data").Exists() {
		return nil, fmt.Errorf("响应中未找到data字段")
	}

	data := result.Get("data")
	return p.parseVideoData(data)
}

// parseVideoData 解析视频数据
func (p *DouyinParser) parseVideoData(data gjson.Result) (*videosdk.VideoInfo, error) {
	videoInfo := &videosdk.VideoInfo{
		Extra: make(map[string]interface{}),
	}

	// 基础信息
	videoInfo.ID = data.Get("id").String()
	videoInfo.Title = data.Get("desc").String()
	videoInfo.Description = data.Get("desc").String()
	videoInfo.URL = data.Get("share_url").String()
	videoInfo.Duration = data.Get("duration").String()

	// 视频类型
	videoType := data.Get("type").String()
	switch videoType {
	case "视频":
		videoInfo.Type = videosdk.VideoTypeVideo
	case "图集":
		videoInfo.Type = videosdk.VideoTypeImage
	case "实况":
		videoInfo.Type = videosdk.VideoTypeLive
	default:
		videoInfo.Type = videosdk.VideoTypeUnknown
	}

	// 创建时间
	if createTime := data.Get("create_time").String(); createTime != "" {
		if t, err := time.Parse("2006-01-02 15:04:05", createTime); err == nil {
			videoInfo.CreateTime = t
		}
	}

	// 媒体信息
	videoInfo.CoverURL = data.Get("static_cover").String()
	videoInfo.Width = int(data.Get("width").Int())
	videoInfo.Height = int(data.Get("height").Int())

	// 处理下载链接
	downloads := data.Get("downloads")
	if downloads.IsArray() {
		// 图文类型，多个下载链接
		downloads.ForEach(func(key, value gjson.Result) bool {
			// 如果链接包含douyinpic.com则是图片，否则是视频
			if strings.Contains(value.String(), "douyinpic.com") {
				videoInfo.Downloads = append(videoInfo.Downloads, videosdk.DownloadItem{
					URL:  value.String(),
					Type: videosdk.MediaTypeImage,
				})
			} else {
				videoInfo.Downloads = append(videoInfo.Downloads, videosdk.DownloadItem{
					URL:  value.String(),
					Type: videosdk.MediaTypeVideo,
				})
			}
			return true
		})
	} else if downloads.Exists() && downloads.String() != "" {
		// 视频类型，单个下载链接
		videoInfo.Downloads = append(videoInfo.Downloads, videosdk.DownloadItem{
			URL:  downloads.String(),
			Type: videosdk.MediaTypeVideo,
		})
	}

	// 作者信息
	videoInfo.Author = videosdk.AuthorInfo{
		UID:       data.Get("uid").String(),
		SecUID:    data.Get("sec_uid").String(),
		UniqueID:  data.Get("unique_id").String(),
		Nickname:  data.Get("nickname").String(),
		Signature: data.Get("signature").String(),
		Age:       int(data.Get("user_age").Int()),
	}

	// 统计信息
	videoInfo.Stats = videosdk.VideoStats{
		PlayCount:    data.Get("play_count").Int(),
		LikeCount:    data.Get("digg_count").Int(),
		CommentCount: data.Get("comment_count").Int(),
		ShareCount:   data.Get("share_count").Int(),
		CollectCount: data.Get("collect_count").Int(),
	}

	// 音乐信息
	videoInfo.Music = videosdk.MusicInfo{
		Title:  data.Get("music_title").String(),
		Author: data.Get("music_author").String(),
		URL:    data.Get("music_url").String(),
	}

	// 标签信息
	tags := data.Get("text_extra").Array()
	for _, tag := range tags {
		videoInfo.Tags = append(videoInfo.Tags, tag.String())
	}

	// 分类标签
	categoryTags := data.Get("tag").Array()
	for _, tag := range categoryTags {
		videoInfo.Tags = append(videoInfo.Tags, tag.String())
	}

	// 扩展信息
	videoInfo.Extra["collection_time"] = data.Get("collection_time").String()
	videoInfo.Extra["create_timestamp"] = data.Get("create_timestamp").Int()
	videoInfo.Extra["uri"] = data.Get("uri").String()
	videoInfo.Extra["dynamic_cover"] = data.Get("dynamic_cover").String()
	videoInfo.Extra["mark"] = data.Get("mark").String()

	return videoInfo, nil
}
