package services

import (
	"errors"
	"uav_defender/internal/dto"
	"uav_defender/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// UserService 用户服务接口
type UserService interface {
	GetUsers(req dto.UserQueryRequest) (*dto.PaginatedResponse[models.User], error)
	CreateUser(req dto.CreateUserRequest) (*models.User, error)
	UpdateUser(userId uint, req dto.UpdateUserRequest) (*models.User, error)
	AssignRole(userId uint, req dto.AssignRoleRequest) (*models.User, error)
	Login(req dto.LoginRequest) (*dto.LoginResponse, error)
}

// userService 用户服务实现
type userService struct {
	db            *gorm.DB
	commonService CommonService
	roleService   RoleService
}

// NewUserService 创建用户服务实例
func NewUserService(db *gorm.DB, commonService CommonService, roleService RoleService) UserService {
	return &userService{
		db:            db,
		commonService: commonService,
		roleService:   roleService,
	}
}

// GetUsers 获取用户列表
func (s *userService) GetUsers(req dto.UserQueryRequest) (*dto.PaginatedResponse[models.User], error) {
	query := s.db.Model(&models.User{}).Preload(clause.Associations)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var users []models.User
	if err := paginate(query, req.Page, req.PageSize).Order("created_at DESC").Find(&users).Error; err != nil {
		return nil, err
	}

	return &dto.PaginatedResponse[models.User]{
		Total: total,
		Items: users,
	}, nil
}

// CreateUser 创建用户
func (s *userService) CreateUser(req dto.CreateUserRequest) (*models.User, error) {
	userData := &models.User{
		Nickname: req.Nickname,
		Username: req.Username,
		Password: req.Password, // TODO：实际应用中应该加密密码
		Email:    req.Email,
		Phone:    req.Phone,
		Avatar:   req.Avatar,
	}

	if err := s.db.Create(userData).Error; err != nil {
		return nil, err
	}
	return userData, nil
}

// UpdateUser 修改用户
func (s *userService) UpdateUser(userId uint, req dto.UpdateUserRequest) (*models.User, error) {
	var user models.User
	if err := s.commonService.GetItemByID(userId, &user); err != nil {
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

	if err := s.db.Save(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// AssignRole 分配角色
func (s *userService) AssignRole(userId uint, req dto.AssignRoleRequest) (*models.User, error) {
	var user models.User
	if err := s.commonService.GetItemByID(userId, &user); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}

	// 检查角色是否存在
	exists, err := s.roleService.IsRoleExist(req.RoleID)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, errors.New("角色不存在")
	}

	user.RoleID = &req.RoleID

	if err := s.db.Save(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *userService) Login(req dto.LoginRequest) (*dto.LoginResponse, error) {
	var user models.User
	if err := s.db.Where("username = ? AND password = ?", req.Username, req.Password).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("用户名或密码错误")
		}
		return nil, err
	}

	// 判断用户是否有分配角色
	if user.RoleID == nil {
		return nil, errors.New("用户未分配角色")
	}

	return &dto.LoginResponse{
		User:  user,
		Token: "some-jwt-token",
	}, nil
}
