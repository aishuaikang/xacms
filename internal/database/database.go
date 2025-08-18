package database

import (
	"errors"
	"fmt"
	"log"
	"new-spbatc-drone-platform/internal/database/models"
	"new-spbatc-drone-platform/internal/routes/dto"
	"os"

	"github.com/google/uuid"
	_ "github.com/joho/godotenv/autoload"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Service 与数据库交互的服务
type Service interface {
	// 获取用户列表
	GetUsers(req dto.UserQueryRequest) (*dto.PaginatedResponse[models.UserModel], error)
	// 创建用户
	CreateUser(req dto.CreateUserRequest) error
	// 获取用户
	GetUser(id uuid.UUID) (*models.UserModel, error)
	// 修改用户
	UpdateUser(userId uuid.UUID, req dto.UpdateUserRequest) error
	// 删除用户
	DeleteUser(userId uuid.UUID) error
}

type service struct {
	db *gorm.DB
}

var (
	dbname     = os.Getenv("BLUEPRINT_DB_DATABASE")
	password   = os.Getenv("BLUEPRINT_DB_PASSWORD")
	username   = os.Getenv("BLUEPRINT_DB_USERNAME")
	port       = os.Getenv("BLUEPRINT_DB_PORT")
	host       = os.Getenv("BLUEPRINT_DB_HOST")
	dbInstance *service
)

func New() Service {
	// Reuse Connection
	if dbInstance != nil {
		return dbInstance
	}

	db, err := gorm.Open(mysql.Open(fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", username, password, host, port, dbname)), &gorm.Config{
		// 打印日志
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal(err)
	}

	dbInstance = &service{
		db: db,
	}

	db.AutoMigrate(&models.RoleModel{}, &models.MenuModel{}, &models.UserModel{}, &models.TenantModel{}, &models.DepartmentModel{})

	return dbInstance
}

// GetUsers 获取用户列表
func (s *service) GetUsers(req dto.UserQueryRequest) (*dto.PaginatedResponse[models.UserModel], error) {
	var users []models.UserModel
	query := s.db.Model(&models.UserModel{})

	// 分页参数
	page := req.Page
	pageSize := req.PageSize

	offset := (page - 1) * pageSize

	if err := query.Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, err
	}
	return &dto.PaginatedResponse[models.UserModel]{
		Total: int64(len(users)),
		Items: users,
	}, nil
}

// CreateUser 创建用户
func (s *service) CreateUser(req dto.CreateUserRequest) error {
	// 转换为数据库模型
	userData := models.UserModel{
		ID:           uuid.New(),
		Nickname:     req.Nickname,
		Username:     req.Username,
		Password:     req.Password, // 注意：实际应用中应该加密密码
		Email:        req.Email,
		Phone:        req.Phone,
		Avatar:       req.Avatar,
		Status:       req.Status,
		RoleID:       req.RoleID,
		TenantID:     req.TenantID,
		DepartmentID: req.DepartmentID,
	}

	if err := s.db.Create(userData).Error; err != nil {
		return err
	}
	return nil
}

// GetUser 获取用户
func (s *service) GetUser(id uuid.UUID) (*models.UserModel, error) {
	var user models.UserModel
	if err := s.db.First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUser 修改用户
func (s *service) UpdateUser(userId uuid.UUID, req dto.UpdateUserRequest) error {
	user, err := s.GetUser(userId)
	if err != nil {
		return errors.New("用户不存在")
	}

	if req.Nickname != nil {
		user.Nickname = *req.Nickname
	}

	if req.Username != nil {
		user.Username = *req.Username
	}

	if req.Email != nil {
		user.Email = *req.Email
	}

	if req.Phone != nil {
		user.Phone = *req.Phone
	}

	if req.Avatar != nil {
		user.Avatar = *req.Avatar
	}

	if req.Status != nil {
		user.Status = req.Status
	}

	if req.RoleID != nil {
		user.RoleID = req.RoleID
	}

	if req.DepartmentID != nil {
		user.DepartmentID = req.DepartmentID
	}

	if err := s.db.Save(user).Error; err != nil {
		return err
	}
	return nil
}

// DeleteUser 删除用户
func (s *service) DeleteUser(userId uuid.UUID) error {
	if err := s.db.Delete(&models.UserModel{}, "id = ?", userId).Error; err != nil {
		return err
	}
	return nil
}
