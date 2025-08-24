package services

import (
	"errors"
	"new-spbatc-drone-platform/internal/models"
	"new-spbatc-drone-platform/internal/routes/dto"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserService 用户服务接口
type UserService interface {
	GetUsers(req dto.UserQueryRequest) (*dto.PaginatedResponse[models.UserModel], error)
	CreateUser(req dto.CreateUserRequest) (*models.UserModel, error)
	UpdateUser(userId uuid.UUID, req dto.UpdateUserRequest) (*models.UserModel, error)
}

// userService 用户服务实现
type userService struct {
	db            *gorm.DB
	commonService CommonService
}

// NewUserService 创建用户服务实例
func NewUserService(db *gorm.DB, commonService CommonService) UserService {
	return &userService{
		db:            db,
		commonService: commonService,
	}
}

// GetUsers 获取用户列表
func (s *userService) GetUsers(req dto.UserQueryRequest) (*dto.PaginatedResponse[models.UserModel], error) {
	var users []models.UserModel
	query := s.db.Model(&models.UserModel{})

	// 分页参数
	page := req.Page
	pageSize := req.PageSize

	offset := (page - 1) * pageSize

	if err := query.Offset(offset).Limit(pageSize).Find(&users).Order("created_at DESC").Error; err != nil {
		return nil, err
	}
	return &dto.PaginatedResponse[models.UserModel]{
		Total: int64(len(users)),
		Items: users,
	}, nil
}

// CreateUser 创建用户
func (s *userService) CreateUser(req dto.CreateUserRequest) (*models.UserModel, error) {
	// 转换为数据库模型
	userData := &models.UserModel{
		ID:       uuid.New(),
		Nickname: req.Nickname,
		Username: req.Username,
		Password: req.Password, // TODO：实际应用中应该加密密码
		Email:    req.Email,
		Phone:    req.Phone,
		Avatar:   req.Avatar,
		Status:   req.Status,
		RoleID:   req.RoleID,
	}

	if err := s.db.Create(userData).Error; err != nil {
		return nil, err
	}
	return userData, nil
}

// UpdateUser 修改用户
func (s *userService) UpdateUser(userId uuid.UUID, req dto.UpdateUserRequest) (*models.UserModel, error) {
	var user models.UserModel
	if err := s.commonService.GetItemByID(&user, userId); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("用户不存在")
		}
		return nil, err
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
		user.Avatar = req.Avatar
	}

	if req.Status != nil {
		user.Status = req.Status
	}

	if req.RoleID != nil {
		user.RoleID = req.RoleID
	}

	if err := s.db.Save(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
