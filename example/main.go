package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	videosdk "github.com/resdownload/video-parser-sdk"
	"github.com/resdownload/video-parser-sdk/parsers"
)

func main() {
	// 创建SDK实例
	sdk := videosdk.NewSDK()

	// 设置超时时间
	sdk.SetTimeout(60 * time.Second)

	// 设置User-Agent
	sdk.SetUserAgent("VideoParserSDK-Example/1.0")

	// 注册抖音解析器
	douyinParser := parsers.NewDouyinParser("http://localhost:5555") // 替换为实际的API地址
	if err := sdk.RegisterParser(douyinParser); err != nil {
		log.Fatalf("注册抖音解析器失败: %v", err)
	}

	// 注册快手解析器（基于API接口）
	kuaishouParser := parsers.NewKuaishouParser("http://localhost:5557") // 快手API服务地址
	if err := sdk.RegisterParser(kuaishouParser); err != nil {
		log.Fatalf("注册快手解析器失败: %v", err)
	}

	// 注册小红书解析器（基于API接口）
	xiaohongshuParser := parsers.NewXiaohongshuParser("http://localhost:5556") // 小红书API服务地址
	if err := sdk.RegisterParser(xiaohongshuParser); err != nil {
		log.Fatalf("注册小红书解析器失败: %v", err)
	}

	// 显示支持的平台
	platforms := sdk.GetSupportedPlatforms()
	fmt.Printf("支持的平台: %v\n\n", platforms)

	// 示例1: 解析抖音视频（使用作品ID）
	fmt.Println("=== 示例1: 解析抖音视频（使用作品ID） ===")
	parseDouyinByID(sdk)

	// 示例2: 解析抖音视频（使用分享链接，三步流程）
	fmt.Println("\n=== 示例2: 解析抖音视频（使用分享链接） ===")
	parseDouyinByURL(sdk)
	//
	//// 示例3: 解析快手视频（直接URL解析）
	fmt.Println("\n=== 示例3: 解析快手视频（直接URL解析） ===")
	parseKuaishou(sdk)

	// 示例4: 解析小红书内容（直接URL解析）
	fmt.Println("\n=== 示例4: 解析小红书内容（直接URL解析） ===")
	parseXiaohongshu(sdk)

	// 示例5: 批量解析不同平台的视频
	fmt.Println("\n=== 示例5: 批量解析 ===")
	batchParse(sdk)

	// 显示使用说明
	fmt.Println("\n=== 使用说明 ===")
	fmt.Println("1. 抖音解析器: 三步流程 - 分享链接→完整URL→作品ID→视频数据")
	fmt.Println("   - 需要启动抖音API服务(localhost:5555)")
	fmt.Println("   - 需要提供有效的抖音Cookie")
	fmt.Println("2. 快手解析器: 直接通过视频URL获取完整数据")
	fmt.Println("   - 需要启动快手API服务(localhost:5557)")
	fmt.Println("   - 需要提供有效的快手Cookie")
	fmt.Println("3. 小红书解析器: 直接通过视频URL获取完整数据")
	fmt.Println("   - 需要启动小红书API服务(localhost:5556)")
	fmt.Println("   - 需要提供有效的小红书Cookie")
	fmt.Println("4. Cookie获取: 请在浏览器中登录对应平台，然后复制Cookie值")
	fmt.Println("5. API服务: 请确保对应的下载器服务正在运行")
}

// parseDouyinByID 使用视频ID解析抖音视频（需要先通过URL获取ID）
func parseDouyinByID(sdk videosdk.SDK) {
	req := &videosdk.ParseRequest{
		Platform: videosdk.PlatformDouyin,
		VideoID:  "", // 从完整URL中提取的作品ID
		Cookie:   "", // 实际使用时需要提供有效的抖音cookie
		Proxy:    "",
		Source:   false,
	}

	ctx := context.Background()
	resp, err := sdk.ParseVideo(ctx, req)
	if err != nil {
		fmt.Printf("解析失败: %v\n", err)
		return
	}

	printResponse(resp)
}

// parseDouyinByURL 使用URL解析抖音视频（三步流程：短链接→完整URL→作品ID→视频数据）
func parseDouyinByURL(sdk videosdk.SDK) {
	req := &videosdk.ParseRequest{
		Platform: videosdk.PlatformDouyin,
		URL:      "", // 抖音分享短链接
		Cookie:   "", // 实际使用时需要提供有效的抖音cookie
		Proxy:    "", // 可选的代理设置
		Source:   false,
	}

	ctx := context.Background()
	resp, err := sdk.ParseVideo(ctx, req)
	if err != nil {
		fmt.Printf("解析失败: %v\n", err)
		return
	}

	printResponse(resp)
}

// parseKuaishou 解析快手视频（直接通过URL获取数据）
func parseKuaishou(sdk videosdk.SDK) {
	req := &videosdk.ParseRequest{
		Platform: videosdk.PlatformKuaishou,
		URL:      "", // 快手分享链接
		Cookie:   "", // 快手Cookie（必需）
		Proxy:    "", // 可选的代理设置
		Source:   false,
	}

	ctx := context.Background()
	resp, err := sdk.ParseVideo(ctx, req)
	if err != nil {
		fmt.Printf("解析失败: %v\n", err)
		return
	}

	printResponse(resp)
	if resp.Data != nil {
		fmt.Printf("视频ID: %s\n", resp.Data.ID)
		fmt.Printf("标题: %s\n", resp.Data.Title)
		fmt.Printf("作者: %s\n", resp.Data.Author.Nickname)
		fmt.Printf("点赞数: %d\n", resp.Data.Stats.LikeCount)
		fmt.Printf("播放数: %d\n", resp.Data.Stats.PlayCount)
	}
}

// parseXiaohongshu 解析小红书内容（直接通过URL获取数据）
func parseXiaohongshu(sdk videosdk.SDK) {
	req := &videosdk.ParseRequest{
		Platform: videosdk.PlatformXiaohongshu,
		URL:      "", // 小红书作品链接
		Cookie:   "", // 小红书Cookie（必需）
		Proxy:    "", // 可选的代理设置
		Source:   false,
	}

	ctx := context.Background()
	resp, err := sdk.ParseVideo(ctx, req)
	if err != nil {
		fmt.Printf("解析失败: %v\n", err)
		return
	}

	printResponse(resp)
	if resp.Data != nil {
		fmt.Printf("作品ID: %s\n", resp.Data.ID)
		fmt.Printf("标题: %s\n", resp.Data.Title)
		fmt.Printf("作者: %s\n", resp.Data.Author.Nickname)
		fmt.Printf("类型: %s\n", resp.Data.Type)
		fmt.Printf("点赞数: %d\n", resp.Data.Stats.LikeCount)
		fmt.Printf("收藏数: %d\n", resp.Data.Stats.CollectCount)
		fmt.Printf("标签: %v\n", resp.Data.Tags)
	}
}

// batchParse 批量解析不同平台的视频
func batchParse(sdk videosdk.SDK) {
	requests := []*videosdk.ParseRequest{
		{
			Platform: videosdk.PlatformDouyin,
			URL:      "https://v.douyin.com/iFRMqmyv/", // 抖音分享短链接
			Cookie:   "your_douyin_cookie",
		},
		{
			Platform: videosdk.PlatformKuaishou,
			URL:      "https://v.kuaishou.com/3xMsre", // 快手分享链接
			Cookie:   "your_kuaishou_cookie",
		},
		{
			Platform: videosdk.PlatformXiaohongshu,
			URL:      "https://www.xiaohongshu.com/explore/65e6b4b3000000001203e5b7", // 小红书作品链接
			Cookie:   "your_xiaohongshu_cookie",
		},
	}

	ctx := context.Background()
	for i, req := range requests {
		fmt.Printf("解析第%d个视频...\n", i+1)
		resp, err := sdk.ParseVideo(ctx, req)
		if err != nil {
			fmt.Printf("  失败: %v\n", err)
		} else {
			fmt.Printf("  成功: 平台=%s, 标题=%s\n", resp.Data.Platform, resp.Data.Title)
		}
	}
}

// printResponse 打印响应结果
func printResponse(resp *videosdk.ParseResponse) {
	jsonData, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		fmt.Printf("JSON序列化失败: %v\n", err)
		return
	}

	fmt.Printf("响应结果:\n%s\n", string(jsonData))

	// 展示下载链接信息
	if resp.Data != nil && len(resp.Data.Downloads) > 0 {
		fmt.Printf("\n下载链接信息:\n")
		for i, download := range resp.Data.Downloads {
			fmt.Printf("  [%d] 类型: %s, 链接: %s\n", i+1, download.Type, download.URL)
		}
	}
}
