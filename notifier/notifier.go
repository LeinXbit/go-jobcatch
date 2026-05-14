package notifier

import (
    "go-catch/logger"
    "go-catch/model"
    "go.uber.org/zap"
    "strings"
)

type ConsoleNotifier struct{}

func NewConsoleNotifier() *ConsoleNotifier {
    return &ConsoleNotifier{}
}

func (n *ConsoleNotifier) NotifyNewJob(job model.Job) {
    separator := strings.Repeat("=", 50)

    // 输出分隔线
    logger.Log.Info(separator)

    // 输出岗位标题
    logger.Log.Info("New job found",
        zap.String("title", job.Title),
        zap.String("city", job.City),
    )

    // 输出薪资
    if job.Salary != "" {
        logger.Log.Info("Salary",
            zap.String("range", job.Salary),
        )
    } else {
        logger.Log.Info("Salary",
            zap.String("range", "negotiable"),
        )
    }

    // 输出公司名称
    logger.Log.Info("Company",
        zap.String("name", job.Company),
    )

    // 输出岗位链接
    if job.URL != "" {
        logger.Log.Info("Job URL",
            zap.String("url", job.URL),
        )
    }

    // 输出分隔线
    logger.Log.Info(separator)
}

func (n *ConsoleNotifier) NotifyBatch(jobs []model.Job) {
    if len(jobs) == 0 {
        return
    }

    logger.Log.Info("Batch notification",
        zap.Int("count", len(jobs)),
    )

    for _, job := range jobs {
        n.NotifyNewJob(job)
    }
}