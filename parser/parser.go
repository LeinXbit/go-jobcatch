package parser

import (
    "bytes"
    "strconv"
    "strings"

    "go-catch/model"
    "github.com/PuerkitoBio/goquery"
)

// Job51Parser 51job 解析器
type Job51Parser struct {
    Keywords []string
}

// NewJob51Parser 创建解析器
func NewJob51Parser(keywords []string) *Job51Parser {
    return &Job51Parser{Keywords: keywords}
}

// Parse 解析 HTML 数据
func (p *Job51Parser) Parse(data []byte, city string) ([]model.Job, error) {
    doc, err := goquery.NewDocumentFromReader(bytes.NewReader(data))
    if err != nil {
        return nil, err
    }

    var jobs []model.Job

    doc.Find(".jod").Each(func(i int, s *goquery.Selection) {
        // 提取标题
        title := strings.TrimSpace(s.Find(".jname at").Text())

        // 提取公司
        company := strings.TrimSpace(s.Find(".cname at").Text())

        // 提取薪资
        salary := strings.TrimSpace(s.Find(".sal").Text())

        // 提取岗位链接
        jobHref, exists := s.Find(".jname at").Attr("href")
        jobURL := ""
        if exists {
            jobURL = "https://search.51job.com" + jobHref
        }

        // 提取岗位 ID（字符串类型）
        idStr, exists := s.Attr("id")
        if !exists {
            idStr = strconv.Itoa(i)
        }

        // 转换 ID 为 uint
        var idUint uint
        if idStr != "" {
            parsedID, err := strconv.ParseUint(idStr, 10, 64)
            if err == nil {
                idUint = uint(parsedID)
            }
        }

        // 关键词过滤
        if title != "" && p.matchKeywords(title) {
            job := model.Job{
                ID:      idUint,      // 数据库自增 ID
                JobID:   idStr,       // 岗位唯一标识（字符串）
                Title:   title,
                Company: company,
                City:    city,
                Salary:  salary,
                URL:     jobURL,
            }
            jobs = append(jobs, job)
        }
    })

    return jobs, nil
}

// matchKeywords 匹配关键词
func (p *Job51Parser) matchKeywords(title string) bool {
    lower := strings.ToLower(title)
    for _, kw := range p.Keywords {
        if strings.Contains(lower, strings.ToLower(kw)) {
            return true
        }
    }
    return false
}