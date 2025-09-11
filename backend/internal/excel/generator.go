package excel

import (
	"fmt"
	"jobView-backend/internal/model"
	"time"

	"github.com/xuri/excelize/v2"
)

// Generator Excel 文件生成器
// 位置：/backend/internal/excel/generator.go
// 功能：负责将求职投递数据生成为 Excel 文件，支持流式处理和样式设置
// 依赖：依赖 Excelize v2 库和内部数据模型
type Generator struct {
	file        *excelize.File
	sheetName   string
	currentRow  int
	styleConfig *StyleConfig
}

// StyleConfig 样式配置结构
type StyleConfig struct {
	HeaderStyle  int
	DataStyle    int
	DateStyle    int
	StatusStyles map[model.ApplicationStatus]int
}

// NewGenerator 创建新的 Excel 生成器
func NewGenerator() *Generator {
	return &Generator{
		file:      excelize.NewFile(),
		sheetName: "求职投递记录",
		currentRow: 1,
	}
}

// InitializeWorkbook 初始化工作簿，设置样式和表头
func (g *Generator) InitializeWorkbook() error {
	// 重命名默认工作表
	if err := g.file.SetSheetName("Sheet1", g.sheetName); err != nil {
		return fmt.Errorf("设置工作表名称失败: %v", err)
	}

	// 初始化样式配置
	if err := g.initializeStyles(); err != nil {
		return fmt.Errorf("初始化样式失败: %v", err)
	}

	// 设置表头
	if err := g.setHeaders(); err != nil {
		return fmt.Errorf("设置表头失败: %v", err)
	}

	return nil
}

// initializeStyles 初始化所有样式
func (g *Generator) initializeStyles() error {
	var err error
	g.styleConfig = &StyleConfig{
		StatusStyles: make(map[model.ApplicationStatus]int),
	}

	// 表头样式：粗体，背景色，居中
	g.styleConfig.HeaderStyle, err = g.file.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
			Size: 12,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#4CAF50"}, // 绿色背景
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
		},
	})
	if err != nil {
		return err
	}

	// 数据行样式：普通格式，边框
	g.styleConfig.DataStyle, err = g.file.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Horizontal: "left",
			Vertical:   "center",
			WrapText:   true, // 自动换行
		},
		Border: []excelize.Border{
			{Type: "left", Color: "#CCCCCC", Style: 1},
			{Type: "top", Color: "#CCCCCC", Style: 1},
			{Type: "bottom", Color: "#CCCCCC", Style: 1},
			{Type: "right", Color: "#CCCCCC", Style: 1},
		},
	})
	if err != nil {
		return err
	}

	// 日期样式
	g.styleConfig.DateStyle, err = g.file.NewStyle(&excelize.Style{
		NumFmt: 14, // 日期格式 m/d/yy
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "#CCCCCC", Style: 1},
			{Type: "top", Color: "#CCCCCC", Style: 1},
			{Type: "bottom", Color: "#CCCCCC", Style: 1},
			{Type: "right", Color: "#CCCCCC", Style: 1},
		},
	})
	if err != nil {
		return err
	}

	// 状态颜色编码样式
	statusColors := map[model.ApplicationStatus]string{
		model.StatusApplied:          "#E3F2FD", // 浅蓝色
		model.StatusResumeScreening:  "#FFF3E0", // 浅橙色
		model.StatusWrittenTest:      "#F3E5F5", // 浅紫色
		model.StatusFirstInterview:   "#E8F5E8", // 浅绿色
		model.StatusSecondInterview:  "#E8F5E8", // 浅绿色
		model.StatusThirdInterview:   "#E8F5E8", // 浅绿色
		model.StatusHRInterview:      "#E8F5E8", // 浅绿色
		model.StatusOfferReceived:    "#C8E6C9", // 绿色
		model.StatusOfferAccepted:    "#4CAF50", // 深绿色
		model.StatusRejected:         "#FFCDD2", // 浅红色
		model.StatusProcessFinished:  "#F5F5F5", // 灰色
	}

	for status, color := range statusColors {
		styleID, err := g.file.NewStyle(&excelize.Style{
			Fill: excelize.Fill{
				Type:    "pattern",
				Color:   []string{color},
				Pattern: 1,
			},
			Alignment: &excelize.Alignment{
				Horizontal: "center",
				Vertical:   "center",
			},
			Border: []excelize.Border{
				{Type: "left", Color: "#CCCCCC", Style: 1},
				{Type: "top", Color: "#CCCCCC", Style: 1},
				{Type: "bottom", Color: "#CCCCCC", Style: 1},
				{Type: "right", Color: "#CCCCCC", Style: 1},
			},
		})
		if err != nil {
			return err
		}
		g.styleConfig.StatusStyles[status] = styleID
	}

	return nil
}

// setHeaders 设置表头
func (g *Generator) setHeaders() error {
	headers := []string{
		"序号", "公司名称", "职位标题", "投递日期", "当前状态", "薪资范围",
		"工作地点", "面试时间", "面试地点", "面试类型", "HR姓名", "HR电话",
		"HR邮箱", "提醒时间", "跟进日期", "备注", "创建时间", "更新时间",
	}

	// 设置列宽
	columnWidths := []float64{6, 20, 25, 12, 15, 15, 15, 18, 20, 10, 12, 15, 20, 18, 12, 30, 20, 20}

	for i, header := range headers {
		colName, err := excelize.ColumnNumberToName(i + 1)
		if err != nil {
			return err
		}

		// 设置表头值
		cell := fmt.Sprintf("%s%d", colName, g.currentRow)
		if err := g.file.SetCellValue(g.sheetName, cell, header); err != nil {
			return err
		}

		// 应用表头样式
		if err := g.file.SetCellStyle(g.sheetName, cell, cell, g.styleConfig.HeaderStyle); err != nil {
			return err
		}

		// 设置列宽
		if err := g.file.SetColWidth(g.sheetName, colName, colName, columnWidths[i]); err != nil {
			return err
		}
	}

	g.currentRow++
	return nil
}

// WriteJobApplications 批量写入求职投递数据
func (g *Generator) WriteJobApplications(applications []model.JobApplication) error {
	for i, app := range applications {
		if err := g.writeJobApplication(i+1, &app); err != nil {
			return fmt.Errorf("写入第%d条记录失败: %v", i+1, err)
		}
	}
	return nil
}

// WriteJobApplicationStream 流式写入求职投递数据
func (g *Generator) WriteJobApplicationStream(applications <-chan model.JobApplication, totalCount int) error {
	sequenceNumber := 1
	
	for app := range applications {
		if err := g.writeJobApplication(sequenceNumber, &app); err != nil {
			return fmt.Errorf("写入第%d条记录失败: %v", sequenceNumber, err)
		}
		sequenceNumber++
	}
	
	return nil
}

// writeJobApplication 写入单条求职投递记录
func (g *Generator) writeJobApplication(sequenceNumber int, app *model.JobApplication) error {
	row := g.currentRow

	// 数据映射：按照表头顺序
	values := []interface{}{
		sequenceNumber,                          // 序号
		app.CompanyName,                         // 公司名称
		app.PositionTitle,                       // 职位标题
		app.ApplicationDate,                     // 投递日期
		string(app.Status),                      // 当前状态
		g.getString(app.SalaryRange),            // 薪资范围
		g.getString(app.WorkLocation),           // 工作地点
		g.getTimeString(app.InterviewTime),      // 面试时间
		g.getString(app.InterviewLocation),      // 面试地点
		g.getString(app.InterviewType),          // 面试类型
		g.getString(app.HRName),                 // HR姓名
		g.getString(app.HRPhone),                // HR电话
		g.getString(app.HREmail),                // HR邮箱
		g.getTimeString(app.ReminderTime),       // 提醒时间
		g.getString(app.FollowUpDate),           // 跟进日期
		g.getString(app.Notes),                  // 备注
		app.CreatedAt.Format("2006-01-02 15:04:05"), // 创建时间
		app.UpdatedAt.Format("2006-01-02 15:04:05"), // 更新时间
	}

	// 写入数据并应用样式
	for colIndex, value := range values {
		colName, err := excelize.ColumnNumberToName(colIndex + 1)
		if err != nil {
			return err
		}

		cell := fmt.Sprintf("%s%d", colName, row)
		if err := g.file.SetCellValue(g.sheetName, cell, value); err != nil {
			return err
		}

		// 应用特殊样式
		var styleID int
		switch colIndex {
		case 4: // 状态列
			if statusStyle, exists := g.styleConfig.StatusStyles[app.Status]; exists {
				styleID = statusStyle
			} else {
				styleID = g.styleConfig.DataStyle
			}
		case 3, 7, 13, 16, 17: // 日期列
			styleID = g.styleConfig.DateStyle
		default:
			styleID = g.styleConfig.DataStyle
		}

		if err := g.file.SetCellStyle(g.sheetName, cell, cell, styleID); err != nil {
			return err
		}
	}

	g.currentRow++
	return nil
}

// AddStatisticsSheet 添加统计工作表
func (g *Generator) AddStatisticsSheet(stats map[string]interface{}) error {
	statsSheetName := "统计概览"
	
	// 创建新工作表
	index, err := g.file.NewSheet(statsSheetName)
	if err != nil {
		return fmt.Errorf("创建统计工作表失败: %v", err)
	}

	// 设置为活动工作表
	g.file.SetActiveSheet(index)

	// 设置统计表头
	if err := g.setStatisticsHeaders(statsSheetName); err != nil {
		return fmt.Errorf("设置统计表头失败: %v", err)
	}

	// 写入统计数据
	if err := g.writeStatisticsData(statsSheetName, stats); err != nil {
		return fmt.Errorf("写入统计数据失败: %v", err)
	}

	// 切回主工作表
	g.file.SetActiveSheet(0)
	
	return nil
}

// setStatisticsHeaders 设置统计表头
func (g *Generator) setStatisticsHeaders(sheetName string) error {
	headers := [][]string{
		{"状态统计", "", ""},
		{"状态", "数量", "百分比"},
	}

	for rowIndex, rowData := range headers {
		for colIndex, header := range rowData {
			colName, err := excelize.ColumnNumberToName(colIndex + 1)
			if err != nil {
				return err
			}
			
			cell := fmt.Sprintf("%s%d", colName, rowIndex+1)
			if err := g.file.SetCellValue(sheetName, cell, header); err != nil {
				return err
			}
			
			// 应用样式
			if rowIndex == 0 {
				// 合并标题行
				if colIndex == 0 {
					if err := g.file.MergeCell(sheetName, "A1", "C1"); err != nil {
						return err
					}
				}
			}
			
			if err := g.file.SetCellStyle(sheetName, cell, cell, g.styleConfig.HeaderStyle); err != nil {
				return err
			}
		}
	}

	return nil
}

// writeStatisticsData 写入统计数据
func (g *Generator) writeStatisticsData(sheetName string, stats map[string]interface{}) error {
	currentRow := 3

	if statusDistribution, ok := stats["statusDistribution"].(map[string]int); ok {
		total := 0
		for _, count := range statusDistribution {
			total += count
		}

		for status, count := range statusDistribution {
			percentage := float64(count) / float64(total) * 100

			// 状态
			if err := g.file.SetCellValue(sheetName, fmt.Sprintf("A%d", currentRow), status); err != nil {
				return err
			}
			// 数量
			if err := g.file.SetCellValue(sheetName, fmt.Sprintf("B%d", currentRow), count); err != nil {
				return err
			}
			// 百分比
			if err := g.file.SetCellValue(sheetName, fmt.Sprintf("C%d", currentRow), fmt.Sprintf("%.1f%%", percentage)); err != nil {
				return err
			}

			// 应用样式
			for col := 1; col <= 3; col++ {
				colName, _ := excelize.ColumnNumberToName(col)
				cell := fmt.Sprintf("%s%d", colName, currentRow)
				if err := g.file.SetCellStyle(sheetName, cell, cell, g.styleConfig.DataStyle); err != nil {
					return err
				}
			}

			currentRow++
		}
	}

	return nil
}

// SaveToFile 保存Excel文件到指定路径
func (g *Generator) SaveToFile(filePath string) error {
	// 保护工作表
	if err := g.file.ProtectSheet(g.sheetName, &excelize.SheetProtectionOptions{
		Password:      "",
		EditScenarios: false,
	}); err != nil {
		return fmt.Errorf("保护工作表失败: %v", err)
	}

	// 保存文件
	if err := g.file.SaveAs(filePath); err != nil {
		return fmt.Errorf("保存文件到 %s 失败: %v", filePath, err)
	}

	return nil
}

// GetBuffer 获取Excel文件的字节缓冲区，用于直接下载
func (g *Generator) GetBuffer() ([]byte, error) {
	buffer, err := g.file.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("生成Excel缓冲区失败: %v", err)
	}
	
	return buffer.Bytes(), nil
}

// Close 关闭Excel文件并释放资源
func (g *Generator) Close() error {
	if g.file != nil {
		return g.file.Close()
	}
	return nil
}

// getString 安全获取字符串指针的值
func (g *Generator) getString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// getTimeString 安全格式化时间指针
func (g *Generator) getTimeString(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02 15:04")
}

// GenerateFilename 生成标准化的文件名
func GenerateFilename(username string, timestamp time.Time) string {
	timeStr := timestamp.Format("20060102_150405")
	return fmt.Sprintf("求职投递记录_%s_%s.xlsx", username, timeStr)
}

// EstimateFileSize 估算文件大小（字节）
func EstimateFileSize(recordCount int) int64 {
	// 基础文件大小约 10KB
	baseSize := int64(10240)
	
	// 每条记录大约 500 字节
	recordSize := int64(recordCount * 500)
	
	return baseSize + recordSize
}

// ValidateData 验证数据完整性
func (g *Generator) ValidateData(applications []model.JobApplication) error {
	if len(applications) == 0 {
		return fmt.Errorf("没有可导出的数据")
	}
	
	// 检查必填字段
	for i, app := range applications {
		if app.CompanyName == "" {
			return fmt.Errorf("第%d条记录缺少公司名称", i+1)
		}
		if app.PositionTitle == "" {
			return fmt.Errorf("第%d条记录缺少职位标题", i+1)
		}
	}
	
	return nil
}