package main

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func NewGoRAGServiceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "langchain-ollama-rag",
		Short: "学习基于langchaingo构建的rag应用",
		Long:  `学习基于langchaingo构建的rag应用`,

		// stop printing usage when the command errors
		SilenceUsage: true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run()
		},
		PostRun: func(cmd *cobra.Command, args []string) {

		},
	}

	cobra.OnInitialize(initLog)
	// ========
	cmd.AddCommand(FileToChunksCmd)
	FileToChunksCmd.Flags().StringP("filepath", "f", "livenet.txt", "指定文件路径, 默认 livenet.txt")
	FileToChunksCmd.Flags().IntP("chunksize", "c", 200, "指定块大小，默认为100")
	FileToChunksCmd.Flags().IntP("chunkoverlap", "o", 50, "指定块重叠大小，默认为10")
	// ========
	cmd.AddCommand(EmbeddingCmd)
	EmbeddingCmd.Flags().StringP("filepath", "f", "livenet.txt", "指定文件路径, 默认为 livenet.txt")
	EmbeddingCmd.Flags().IntP("chunksize", "c", 200, "指定块大小，默认为100")
	EmbeddingCmd.Flags().IntP("chunkoverlap", "o", 50, "指定块重叠大小，默认为10")
	// ========
	cmd.AddCommand(RetrieverCmd)
	RetrieverCmd.Flags().IntP("topk", "t", 5, "召回数据的数量，默认为20")
	// ========
	cmd.AddCommand(GetAnwserCmd)
	GetAnwserCmd.Flags().IntP("topk", "t", 5, "召回数据的数量，默认为20")
	return cmd
}

func initLog() {
	logrus.SetReportCaller(true)
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
		PadLevelText:  true,
		CallerPrettyfier: func(frame *runtime.Frame) (function string, file string) {
			function = fmt.Sprintf("%s:%d", frame.Function, frame.Line)
			function = strings.TrimPrefix(function, "github.com/kiosk404/")
			return function, ""
		},
		TimestampFormat: time.DateTime,
	})
}

var FileToChunksCmd = &cobra.Command{
	Use:   "filetochunks",
	Short: "将文件转换为块",
	Run: func(cmd *cobra.Command, args []string) {
		filepath, _ := cmd.Flags().GetString("filepath")
		chunkSize, _ := cmd.Flags().GetInt("chunksize")
		chunkOverlap, _ := cmd.Flags().GetInt("chunkoverlap")
		docs, err := TextToChunks(filepath, chunkSize, chunkOverlap)
		if err != nil {
			logrus.Errorf("转换文件为块失败，错误信息: %v", err)
		}
		logrus.Infof("转换文件为块成功，块数量: %d", len(docs))
		for _, v := range docs {
			fmt.Printf("🗂 块内容==> %v\n", v.PageContent)
		}
		fmt.Println("输入:", filepath, chunkSize, chunkOverlap)
	},
}

var EmbeddingCmd = &cobra.Command{
	Use:   "embedding",
	Short: "将文档块转换为向量",
	Run: func(cmd *cobra.Command, args []string) {
		filepath, _ := cmd.Flags().GetString("filepath")
		chunkSize, _ := cmd.Flags().GetInt("chunksize")
		chunkOverlap, _ := cmd.Flags().GetInt("chunkoverlap")
		docs, err := TextToChunks(filepath, chunkSize, chunkOverlap)
		if err != nil {
			logrus.Errorf("转换文件为块失败，错误信息: %v", err.Error())
		}
		err = storeDocs(docs, getStore())
		if err != nil {
			logrus.Errorf("转换块为向量失败，错误信息: %v", err.Error())
		} else {
			logrus.Info("转换块为向量成功")
		}
	},
}

var RetrieverCmd = &cobra.Command{
	Use:   "retriever",
	Short: "将用户问题转换为向量并检索文档",
	Run: func(cmd *cobra.Command, args []string) {
		topk, _ := cmd.Flags().GetInt("topk")

		// 获取用户输入的问题
		prompt, err := GetUserInput("请输入你的问题")
		if err != nil {
			logrus.Error("获取用户输入失败，错误信息: %v", err)
		}
		rst, err := useRetriaver(getStore(), prompt, topk)
		if err != nil {
			logrus.Error("检索文档失败，错误信息: %v", err)
		}
		for _, v := range rst {
			fmt.Printf("🗂 根据输入的内容检索出的块内容==> %v\n", v.PageContent)
		}
	},
}

var GetAnwserCmd = &cobra.Command{
	Use:   "getanswer",
	Short: "获取回答",
	Run: func(cmd *cobra.Command, args []string) {
		topk, _ := cmd.Flags().GetInt("topk")

		prompt, err := GetUserInput("请输入你的问题")
		if err != nil {
			logrus.Error("获取用户输入失败，错误信息: %v", err)
		}
		rst, err := useRetriaver(getStore(), prompt, topk)
		if err != nil {
			logrus.Error("检索文档失败，错误信息: %v", err)
		}
		answer, err := GetAnswer(context.Background(), getOllamaMistral(), rst, prompt)
		if err != nil {
			logrus.Error("获取回答失败，错误信息: %v", err)
		} else {
			fmt.Printf("🗂 原始回答==> %s\n\n", answer)
			rst, err := Translate(getOllamaLlama2(), answer)
			if err != nil {
				logrus.Error("翻译回答失败，错误信息: %v", err)
			} else {
				fmt.Printf("🗂 翻译后的回答==> %s\n", rst)
			}
		}
	},
}

func run() error {
	logrus.Infoln("start langchaingo-ollama-rag app")
	GetTextInput()
	return nil
}
