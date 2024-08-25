package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/devalexandre/pdfbox"
	"github.com/imroc/req/v3"
	"github.com/sirupsen/logrus"
)

// FileProcessor 定义了处理文件的接口
type FileProcessor interface {
	Process(filePath string) error
}

// PDFProcessor 实现了 FileProcessor 接口，用于处理 PDF 文件
type PDFProcessor struct {
	outTxt string
}

func (p *PDFProcessor) Process(filePath string) error {
	logrus.Info("Processing PDF file:", filePath)
	text, err := pdfbox.ExtractTextFromPdf(filePath)
	if err != nil {
		return errors.New("failed to extract text from PDF:" + err.Error())
	}
	p.outTxt = text
	// 创建文件，如果文件已存在，会被覆盖
	file, err := os.Create(changeFileExtension(filePath, "txt"))
	if err != nil {
		return fmt.Errorf("failed to create text")
	}
	defer file.Close() // 确保文件在函数结束时关闭

	// 写入内容
	_, err = file.WriteString(text)
	return err
}

// FileDownloader 定义了下载文件的接口
type FileDownloader interface {
	Download(url string, filePath string) error
}

// HTTPFileDownloader 实现了 FileDownloader 接口，用于从 HTTP 源下载文件
type HTTPFileDownloader struct{}

func (d *HTTPFileDownloader) Download(url string, filePath string) error {
	// 检查文件是否已经存在
	if _, err := os.Stat(filePath); err == nil {
		logrus.Println("File already exists, skipping download.")
		return nil
	}
	client := req.C().DevMode()

	client.SetRedirectPolicy(
		// Only allow up to 5 redirects
		req.MaxRedirectPolicy(5),
		// Only allow redirect to same domain.
		// e.g. redirect "www.imroc.cc" to "imroc.cc" is allowed, but "google.com" is not
		req.SameDomainRedirectPolicy(),
	)

	resp, err := client.R().Get(url)
	if err != nil {
		return err
	}
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	return err
}

// FileService 结合了文件下载和处理的功能
type FileService struct {
	downloader FileDownloader
	processor  FileProcessor
}

func NewFileService() *FileService {
	return &FileService{
		downloader: &HTTPFileDownloader{},
		processor:  &PDFProcessor{},
	}
}

func (s *FileService) DownloadAndProcess(url string, filePath string) error {
	err := s.downloader.Download(url, filePath)
	if err != nil {
		return err
	}

	return s.processor.Process(filePath)
}

func GetTextInput() {
	pdfURL := "https://dl.acm.org/doi/pdf/10.1145/3544216.3544236"
	pdfPath := filepath.Join("./", "livenet.pdf")

	fileService := NewFileService()
	err := fileService.DownloadAndProcess(pdfURL, pdfPath)
	if err != nil {
		logrus.Println("Error downloading and processing PDF:", err)
	}
}

// changeFileExtension 函数接受原始文件名和新的后缀，返回更改后缀的新文件名
func changeFileExtension(fileName, newExt string) string {
	// 获取文件的基本名（不包括扩展名）
	baseName := strings.TrimSuffix(filepath.Base(fileName), filepath.Ext(fileName))

	// 构建新的文件名，添加新的后缀
	newFileName := baseName + "." + newExt

	return newFileName
}
