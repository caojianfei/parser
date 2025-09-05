package parsers

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	videosdk "github.com/caojianfei/parser"
	"github.com/go-resty/resty/v2"
	"github.com/tidwall/gjson"
)

// KuaishouParser 快手解析器
type KuaishouParser struct {
	client  *resty.Client
	baseURL string
}

// NewKuaishouParser 创建快手解析器
func NewKuaishouParser(baseURL string) videosdk.Parser {
	client := resty.New()
	client.SetTimeout(30 * time.Second)
	client.SetHeader("Content-Type", "application/json")
	client.SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	return &KuaishouParser{
		client:  client,
		baseURL: baseURL,
	}
}

// GetPlatform 获取平台类型
func (p *KuaishouParser) GetPlatform() videosdk.Platform {
	return videosdk.PlatformKuaishou
}

// ExtractVideoID 从URL提取视频ID（快手不需要提取ID，直接使用URL）
func (p *KuaishouParser) ExtractVideoID(url string) (string, error) {
	// 快手解析器不需要提取视频ID，直接使用完整URL
	// 这个方法保留是为了兼容接口
	return url, nil
}

// ValidateRequest 验证请求参数
func (p *KuaishouParser) ValidateRequest(req *videosdk.ParseRequest) error {
	if req.VideoID == "" && req.URL == "" {
		return fmt.Errorf("video_id 或 url 至少需要提供一个")
	}

	if req.Platform != videosdk.PlatformKuaishou {
		return fmt.Errorf("平台类型不匹配，期望: %s，实际: %s", videosdk.PlatformKuaishou, req.Platform)
	}

	return nil
}

// ParseVideo 解析视频信息
func (p *KuaishouParser) ParseVideo(ctx context.Context, req *videosdk.ParseRequest) (*videosdk.VideoInfo, error) {
	if err := p.ValidateRequest(req); err != nil {
		return nil, err
	}

	// 确定要解析的URL
	var targetURL string
	if req.URL != "" {
		targetURL = req.URL
	} else if req.VideoID != "" {
		// 如果只提供了VideoID，假设它是一个短链接或完整URL
		targetURL = req.VideoID
	} else {
		return nil, fmt.Errorf("必须提供URL或VideoID")
	}

	// 构建请求体，按照API文档规范
	requestBody := map[string]interface{}{
		"text":   targetURL,
		"cookie": req.Cookie,
		"proxy":  req.Proxy,
	}

	// 发送请求到快手API的 /detail/ 接口
	resp, err := p.client.R().
		SetContext(ctx).
		SetBody(requestBody).
		Post(p.baseURL + "/detail/")

	if err != nil {
		return nil, fmt.Errorf("请求快手API失败: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("快手API请求失败，状态码: %d", resp.StatusCode())
	}

	// 解析响应
	videoInfo, err := p.parseVideoData(resp.Body())
	if err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return videoInfo, nil
}

// parseVideoData 解析快手API返回的视频数据
func (p *KuaishouParser) parseVideoData(data []byte) (*videosdk.VideoInfo, error) {
	result := gjson.ParseBytes(data)

	// 检查响应是否成功
	message := result.Get("message").String()
	if !strings.Contains(message, "成功") {
		return nil, fmt.Errorf("API返回错误: %s", message)
	}

	videoData := result.Get("data")
	if !videoData.Exists() {
		return nil, errors.New("响应中缺少data字段")
	}

	// 解析基本信息
	videoID := videoData.Get("detailID").String()
	caption := videoData.Get("caption").String()
	photoType := videoData.Get("photoType").String()
	duration := videoData.Get("duration").String()
	coverURL := videoData.Get("coverUrl").String()
	downloadURL := videoData.Get("download").String()
	timestamp := videoData.Get("timestamp").String()

	// 解析统计数据
	realLikeCount := videoData.Get("realLikeCount").Int()
	viewCount := videoData.Get("viewCount")
	shareCount := videoData.Get("shareCount").Int()
	commentCount := videoData.Get("commentCount").Int()

	// 解析作者信息
	authorID := videoData.Get("authorID").String()
	authorName := videoData.Get("name").String()

	// 解析创建时间
	var createTime time.Time
	if timestamp != "" {
		if parsedTime, err := time.Parse("2006-01-02_15:04:05", timestamp); err == nil {
			createTime = parsedTime
		} else {
			createTime = time.Now()
		}
	} else {
		createTime = time.Now()
	}

	// 确定视频类型
	var videoType videosdk.VideoType
	switch photoType {
	case "视频":
		videoType = videosdk.VideoTypeVideo
	case "图片":
		videoType = videosdk.VideoTypeImage
	default:
		videoType = videosdk.VideoTypeUnknown
	}

	// 解析播放量（可能是字符串格式，如"203.7万"）
	var playCount int64
	if viewCount.Exists() {
		if viewCount.Type == gjson.Number {
			playCount = viewCount.Int()
		} else {
			// 处理"203.7万"这样的格式
			viewStr := viewCount.String()
			if strings.Contains(viewStr, "万") {
				numStr := strings.Replace(viewStr, "万", "", -1)
				if num, err := strconv.ParseFloat(numStr, 64); err == nil {
					playCount = int64(num * 10000)
				}
			} else if num, err := strconv.ParseInt(viewStr, 10, 64); err == nil {
				playCount = num
			}
		}
	}

	// 处理下载链接
	var downloadURLs []string
	if downloadURL != "" {
		if strings.Contains(downloadURL, " ") {
			// 多个链接用空格分隔
			downloadURLs = strings.Fields(downloadURL)
		} else {
			downloadURLs = []string{downloadURL}
		}
	}

	// 处理下载链接
	var downloads []videosdk.DownloadItem
	for _, url := range downloadURLs {
		if url != "" {
			mediaType := videosdk.MediaTypeVideo
			if videoType == videosdk.VideoTypeImage {
				mediaType = videosdk.MediaTypeImage
			}
			downloads = append(downloads, videosdk.DownloadItem{
				URL:  url,
				Type: mediaType,
			})
		}
	}

	return &videosdk.VideoInfo{
		ID:          videoID,
		Title:       caption,
		Description: caption,
		Type:        videoType,
		Platform:    videosdk.PlatformKuaishou,
		URL:         fmt.Sprintf("https://www.kuaishou.com/short-video/%s", videoID),
		CreateTime:  createTime,
		Duration:    duration,
		Downloads:   downloads,
		CoverURL:    coverURL,
		Width:       0,
		Height:      0,
		Author: videosdk.AuthorInfo{
			UID:      authorID,
			Nickname: authorName,
		},
		Stats: videosdk.VideoStats{
			PlayCount:    playCount,
			LikeCount:    realLikeCount,
			CommentCount: commentCount,
			ShareCount:   shareCount,
			CollectCount: 0,
		},
		Music: videosdk.MusicInfo{},
		Tags:  []string{},
		Extra: map[string]interface{}{
			"downloadURLs": downloadURLs,
			"photoType":    photoType,
		},
	}, nil
}
