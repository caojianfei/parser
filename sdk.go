package videosdk

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// VideoSDK SDK主实现
type VideoSDK struct {
	parsers   map[Platform]Parser
	mu        sync.RWMutex
	timeout   time.Duration
	userAgent string
}

// NewSDK 创建新的SDK实例
func NewSDK() SDK {
	return &VideoSDK{
		parsers:   make(map[Platform]Parser),
		timeout:   30 * time.Second,
		userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/139.0.0.0 Safari/537.36",
	}
}

// RegisterParser 注册平台解析器
func (s *VideoSDK) RegisterParser(parser Parser) error {
	if parser == nil {
		return fmt.Errorf("parser cannot be nil")
	}

	platform := parser.GetPlatform()
	if platform == "" {
		return fmt.Errorf("parser platform cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.parsers[platform] = parser
	return nil
}

// ParseVideo 解析视频信息
func (s *VideoSDK) ParseVideo(ctx context.Context, req *ParseRequest) (*ParseResponse, error) {
	start := time.Now()
	response := &ParseResponse{
		Time: start,
	}

	// 参数验证
	if req == nil {
		response.Success = false
		response.Error = "request cannot be nil"
		return response, fmt.Errorf("request cannot be nil")
	}

	if req.Platform == "" {
		response.Success = false
		response.Error = "platform is required"
		return response, fmt.Errorf("platform is required")
	}

	// 获取解析器
	s.mu.RLock()
	parser, exists := s.parsers[req.Platform]
	s.mu.RUnlock()

	if !exists {
		response.Success = false
		response.Error = fmt.Sprintf("platform %s is not supported", req.Platform)
		return response, fmt.Errorf("platform %s is not supported", req.Platform)
	}

	// 验证请求参数
	if err := parser.ValidateRequest(req); err != nil {
		response.Success = false
		response.Error = fmt.Sprintf("request validation failed: %v", err)
		return response, fmt.Errorf("request validation failed: %w", err)
	}

	// 设置超时上下文
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	// 解析视频信息
	videoInfo, err := parser.ParseVideo(ctx, req)
	if err != nil {
		response.Success = false
		response.Error = fmt.Sprintf("failed to parse video: %v", err)
		return response, fmt.Errorf("failed to parse video: %w", err)
	}

	// 设置平台信息
	videoInfo.Platform = req.Platform

	response.Success = true
	response.Message = "解析成功"
	response.Data = videoInfo

	return response, nil
}

// GetSupportedPlatforms 获取支持的平台列表
func (s *VideoSDK) GetSupportedPlatforms() []Platform {
	s.mu.RLock()
	defer s.mu.RUnlock()

	platforms := make([]Platform, 0, len(s.parsers))
	for platform := range s.parsers {
		platforms = append(platforms, platform)
	}

	return platforms
}

// SetTimeout 设置请求超时时间
func (s *VideoSDK) SetTimeout(timeout time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.timeout = timeout
}

// SetUserAgent 设置User-Agent
func (s *VideoSDK) SetUserAgent(userAgent string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.userAgent = userAgent
}

// GetTimeout 获取超时时间
func (s *VideoSDK) GetTimeout() time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.timeout
}

// GetUserAgent 获取User-Agent
func (s *VideoSDK) GetUserAgent() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.userAgent
}
