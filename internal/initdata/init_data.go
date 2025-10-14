package initdata

import (
	"embed"
	"fmt"
	"log"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"uav_defender/internal/models"
	"uav_defender/internal/pkg/global"

	"gorm.io/gorm"
)

//go:embed sql/*.sql
var sqlFiles embed.FS

type InitData struct {
	db *gorm.DB
}

func NewInitData(db *gorm.DB) *InitData {
	return &InitData{db: db}
}

// InitFromSQL 从嵌入的SQL文件中执行初始化
func (i *InitData) InitFromSQL() error {
	// 检查是否需要初始化
	if !i.shouldInitialize() {
		global.Logger.Info("数据库已初始化，跳过SQL初始化")
		return nil
	}

	global.Logger.Info("开始执行SQL初始化...")

	// 读取sql目录下的所有SQL文件
	entries, err := sqlFiles.ReadDir("sql")
	if err != nil {
		return fmt.Errorf("读取SQL目录失败: %w", err)
	}

	// 收集所有SQL文件并排序(按文件名排序,确保执行顺序)
	var sqlFileNames []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".sql" {
			sqlFileNames = append(sqlFileNames, entry.Name())
		}
	}
	sort.Strings(sqlFileNames)

	// 按顺序执行每个SQL文件
	for _, fileName := range sqlFileNames {
		if err := i.executeSQLFile(fileName); err != nil {
			return fmt.Errorf("执行SQL文件 %s 失败: %w", fileName, err)
		}
		log.Printf("成功执行SQL文件: %s", fileName)
	}

	log.Println("SQL初始化完成")
	return nil
}

// shouldInitialize 检查是否需要初始化（通过检查是否存在超级管理员）
func (i *InitData) shouldInitialize() bool {
	var count int64
	// 检查是否存在ID为1的角色（超级管理员）
	i.db.Model(&models.Role{}).Where("id = ?", 1).Count(&count)
	return count == 0
}

// executeSQLFile 执行单个SQL文件
func (i *InitData) executeSQLFile(fileName string) error {
	// 读取SQL文件内容
	// 注意: embed.FS 总是使用正斜杠 /，即使在 Windows 上也是如此
	filePath := path.Join("sql", fileName)
	content, err := sqlFiles.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("读取文件失败: %w", err)
	}

	sqlContent := string(content)

	// 分割SQL语句(以分号分隔,忽略注释)
	statements := i.splitSQLStatements(sqlContent)

	// 执行每条SQL语句
	for index, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		if err := i.db.Exec(stmt).Error; err != nil {
			return fmt.Errorf("执行第 %d 条SQL语句失败: %w\nSQL: %s", index+1, err, stmt)
		}
	}

	return nil
}

// splitSQLStatements 分割SQL语句
func (i *InitData) splitSQLStatements(sqlContent string) []string {
	var statements []string
	var currentStmt strings.Builder
	lines := strings.Split(sqlContent, "\n")

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// 跳过注释行
		if strings.HasPrefix(trimmedLine, "--") || trimmedLine == "" {
			continue
		}

		currentStmt.WriteString(line)
		currentStmt.WriteString("\n")

		// 如果行以分号结尾,表示一条SQL语句结束
		if strings.HasSuffix(trimmedLine, ";") {
			statements = append(statements, strings.TrimSpace(currentStmt.String()))
			currentStmt.Reset()
		}
	}

	// 添加最后一条语句(如果有)
	if currentStmt.Len() > 0 {
		stmt := strings.TrimSpace(currentStmt.String())
		if stmt != "" {
			statements = append(statements, stmt)
		}
	}

	return statements
}
