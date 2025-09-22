package parsers

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	videosdk "github.com/caojianfei/parser"
	"github.com/go-resty/resty/v2"
	"github.com/tidwall/gjson"
)

// XiaohongshuParser 小红书解析器
type XiaohongshuParser struct {
	client  *resty.Client
	baseURL string
}

// NewXiaohongshuParser 创建小红书解析器
func NewXiaohongshuParser(baseURL string) videosdk.Parser {
	client := resty.New()
	client.SetTimeout(30 * time.Second)
	client.SetHeader("Content-Type", "application/json")
	client.SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	return &XiaohongshuParser{
		client:  client,
		baseURL: baseURL,
	}
}

// GetPlatform 获取平台类型
func (p *XiaohongshuParser) GetPlatform() videosdk.Platform {
	return videosdk.PlatformXiaohongshu
}

// ExtractVideoID 返回视频URL（小红书API直接接受URL参数）
func (p *XiaohongshuParser) ExtractVideoID(url string) (string, error) {
	// 小红书API直接接受URL参数，无需提取ID
	if url == "" {
		return "", fmt.Errorf("URL不能为空")
	}
	return url, nil
}

// ValidateRequest 验证请求参数
func (p *XiaohongshuParser) ValidateRequest(req *videosdk.ParseRequest) error {
	if req.VideoID == "" && req.URL == "" {
		return fmt.Errorf("video_id 或 url 至少需要提供一个")
	}

	if req.Platform != videosdk.PlatformXiaohongshu {
		return fmt.Errorf("平台类型不匹配，期望: %s，实际: %s", videosdk.PlatformXiaohongshu, req.Platform)
	}

	return nil
}

// ParseVideo 解析视频信息
func (p *XiaohongshuParser) ParseVideo(ctx context.Context, req *videosdk.ParseRequest) (*videosdk.VideoInfo, error) {
	if err := p.ValidateRequest(req); err != nil {
		return nil, err
	}

	// 确定要使用的URL
	url := req.URL
	if url == "" {
		url = req.VideoID // VideoID在小红书中实际就是URL
	}

	if url == "" {
		return nil, fmt.Errorf("URL不能为空")
	}

	// 构建请求体，严格按照API文档规范
	requestBody := map[string]interface{}{
		"url":      url,
		"download": false,
		"index":    []string{},
		"cookie":   req.Cookie,
		"proxy":    req.Proxy,
		"skip":     false,
	}

	// 发送POST请求到小红书API的/xhs/接口
	resp, err := p.client.R().
		SetContext(ctx).
		SetBody(requestBody).
		Post(p.baseURL + "/xhs/detail")

	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("API请求失败，状态码: %d", resp.StatusCode())
	}

	// 解析响应
	videoInfo, err := p.parseVideoData(resp.Body())
	if err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	return videoInfo, nil
}

// parseVideoData 解析小红书API返回的视频数据
func (p *XiaohongshuParser) parseVideoData(data []byte) (*videosdk.VideoInfo, error) {
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
	videoID := videoData.Get("作品ID").String()
	title := videoData.Get("作品标题").String()
	description := videoData.Get("作品描述").String()
	workType := videoData.Get("作品类型").String()
	workLink := videoData.Get("作品链接").String()
	publishTime := videoData.Get("发布时间").String()
	updateTime := videoData.Get("最后更新时间").String()
	timestamp := videoData.Get("时间戳").String()

	// 解析统计数据
	collectCount := videoData.Get("收藏数量").Int()
	commentCount := videoData.Get("评论数量").Int()
	shareCount := videoData.Get("分享数量").Int()
	likeCount := videoData.Get("点赞数量").Int()

	// 解析作者信息
	authorNickname := videoData.Get("作者昵称").String()
	authorID := videoData.Get("作者ID").String()
	authorLink := videoData.Get("作者链接").String()

	// 解析下载地址
	downloadURLs := []string{}
	downloadData := videoData.Get("下载地址")
	if downloadData.IsArray() {
		downloadData.ForEach(func(key, value gjson.Result) bool {
			downloadURLs = append(downloadURLs, value.String())
			return true
		})
	} else if downloadData.Exists() {
		downloadURLs = append(downloadURLs, downloadData.String())
	}

	// 解析动图地址
	gifURLs := []string{}
	gifData := videoData.Get("动图地址")
	if gifData.IsArray() {
		gifData.ForEach(func(key, value gjson.Result) bool {
			gifURLs = append(gifURLs, value.String())
			return true
		})
	} else if gifData.Exists() {
		gifURLs = append(gifURLs, gifData.String())
	}

	// 解析标签
	tags := []string{}
	tagsData := videoData.Get("作品标签")
	if tagsData.IsArray() {
		tagsData.ForEach(func(key, value gjson.Result) bool {
			tags = append(tags, value.String())
			return true
		})
	} else if tagsData.Exists() {
		tagsStr := tagsData.String()
		if tagsStr != "" {
			tags = strings.Split(tagsStr, ",")
		}
	}

	// 解析创建时间
	var createTime time.Time
	if publishTime != "" {
		// 尝试多种时间格式
		timeFormats := []string{
			"2006-01-02 15:04:05",
			"2006-01-02T15:04:05Z",
			"2006-01-02",
		}
		for _, format := range timeFormats {
			if parsedTime, err := time.Parse(format, publishTime); err == nil {
				createTime = parsedTime
				break
			}
		}
		if createTime.IsZero() {
			createTime = time.Now()
		}
	} else {
		createTime = time.Now()
	}

	// 确定视频类型
	var videoType videosdk.VideoType
	switch workType {
	case "视频":
		videoType = videosdk.VideoTypeVideo
	case "图文":
		videoType = videosdk.VideoTypeImage
	default:
		videoType = videosdk.VideoTypeVideo
	}

	// 获取封面图片（通常是第一个下载地址或动图地址）
	var coverURL string
	if len(downloadURLs) > 0 {
		coverURL = downloadURLs[0]
	} else if len(gifURLs) > 0 {
		coverURL = gifURLs[0]
	}

	// 处理下载链接
	var downloads []videosdk.DownloadItem
	// 添加视频下载链接
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
	// 添加GIF下载链接
	for _, url := range gifURLs {
		if url != "" {
			downloads = append(downloads, videosdk.DownloadItem{
				URL:  url,
				Type: videosdk.MediaTypeImage, // GIF作为图片类型
			})
		}
	}

	return &videosdk.VideoInfo{
		ID:          videoID,
		Title:       title,
		Description: description,
		Type:        videoType,
		Platform:    videosdk.PlatformXiaohongshu,
		URL:         workLink,
		CreateTime:  createTime,
		Duration:    "00:00:00",
		Downloads:   downloads,
		CoverURL:    coverURL,
		Width:       0,
		Height:      0,
		Author: videosdk.AuthorInfo{
			UID:      authorID,
			Nickname: authorNickname,
		},
		Stats: videosdk.VideoStats{
			PlayCount:    0,
			LikeCount:    int64(int(likeCount)),
			CommentCount: int64(int(commentCount)),
			ShareCount:   int64(int(shareCount)),
			CollectCount: int64(int(collectCount)),
		},
		Music: videosdk.MusicInfo{},
		Tags:  tags,
		Extra: map[string]interface{}{
			"publishTime":  publishTime,
			"updateTime":   updateTime,
			"timestamp":    timestamp,
			"workType":     workType,
			"downloadURLs": downloadURLs,
			"gifURLs":      gifURLs,
			"authorLink":   authorLink,
		},
	}, nil
}
