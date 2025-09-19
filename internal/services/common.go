package services

import (
	"errors"
	"net/http"
	"sort"
	"uav_defender/internal/dto"
	"uav_defender/internal/pkg/global"
	"uav_defender/internal/pkg/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CommonService 公共服务接口
type CommonService interface {
	GetItems(model any) error
	GetItemByID(id uuid.UUID, model any) error
	IsExistByID(id uuid.UUID, model any) (bool, error)
	DeleteItemByID(model any, id uuid.UUID) error
	ValidateBody(c *gin.Context, model any) error
	ValidateQuery(c *gin.Context, model any) error
	GetAPIs() []dto.APIInfo
}

// commonService 公共服务实现
type commonService struct {
	db        *gorm.DB
	validator *utils.ValidationMiddleware
}

// NewCommonService 创建公共服务实例
func NewCommonService(db *gorm.DB, validator *utils.ValidationMiddleware) CommonService {
	return &commonService{
		db:        db,
		validator: validator,
	}
}

// GetItems 获取多个数据
func (s *commonService) GetItems(model any) error {
	// 先尝试按 order 字段排序，如果失败则只按创建时间排序
	if err := s.db.Order("`order` ASC, created_at DESC").Find(model).Error; err != nil {
		// 如果 order 字段不存在，回退到只按创建时间排序
		if err := s.db.Order("created_at DESC").Find(model).Error; err != nil {
			return err
		}
	}
	return nil
}

// GetItemByID 根据ID获取单个数据
func (s *commonService) GetItemByID(id uuid.UUID, model any) error {
	if err := s.db.First(model, "id = ?", id).Error; err != nil {
		return err
	}
	return nil
}

// IsExistByID 检查ID是否存在
func (s *commonService) IsExistByID(id uuid.UUID, model any) (bool, error) {
	var count int64
	if err := s.db.Model(model).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// DeleteItemByID 根据ID删除单个数据
func (s *commonService) DeleteItemByID(model any, id uuid.UUID) error {
	if err := s.db.Delete(model, "id = ?", id).Error; err != nil {
		return err
	}
	return nil
}

// ValidateBody 验证请求体
func (s *commonService) ValidateBody(c *gin.Context, model any) error {
	// 解析请求体
	if err := c.ShouldBindJSON(model); err != nil {
		global.Logger.Error("解析请求体失败", zap.Error(err))
		return errors.New("请求体格式错误")
	}

	// 验证请求数据
	if errs := s.validator.ValidateStruct(model); len(errs) > 0 {
		return errors.New(errs[0])
	}
	return nil
}

// ValidateQuery 验证查询参数
func (s *commonService) ValidateQuery(c *gin.Context, model any) error {
	// 解析查询参数
	if err := c.ShouldBindQuery(model); err != nil {
		global.Logger.Error("解析查询参数失败", zap.Error(err))
		return errors.New("查询参数格式错误")
	}

	// 验证查询数据
	if errs := s.validator.ValidateStruct(model); len(errs) > 0 {
		return errors.New(errs[0])
	}
	return nil
}

// GetAPIs 获取API列表
func (s *commonService) GetAPIs() []dto.APIInfo {
	routeMap := make(map[string][]dto.APIInfo) // 键: 路径+名称, 值: 具有相同路径+名称的路由

	// allroutes := s.ginServer.Routes()

	// log.Debugf("所有路由: %+v", allroutes)

	// // 按路径+名称分组路由
	// for _, route := range allroutes {
	// 	key := route.Path + "|" + route.Method
	// 	routeMap[key] = append(routeMap[key], dto.APIInfo{
	// 		Method:  route.Method,
	// 		Path:    route.Path,
	// 		Handler: route.Handler,
	// 	})
	// }

	var result []dto.APIInfo
	// 处理每个分组
	for _, routes := range routeMap {
		if len(routes) == 1 {
			// 只有一个路由，无论方法如何都保留它
			result = append(result, routes[0])

		} else {
			// 具有相同路径+名称的多个路由
			hasNonHead := false
			var headRoute *dto.APIInfo

			for i := range routes {
				if routes[i].Method == http.MethodHead {
					if headRoute == nil {
						headRoute = &routes[i]
					}
				} else {
					hasNonHead = true
					result = append(result, routes[i])
				}
			}

			// 如果没有找到非HEAD路由，保留HEAD路由
			if !hasNonHead && headRoute != nil {
				result = append(result, *headRoute)
			}
		}
	}

	// 排序
	sort.Slice(result, func(i, j int) bool {
		return result[i].Path < result[j].Path
	})
	return result
}

// paginate 分页辅助函数
func paginate(query *gorm.DB, page, pageSize int) *gorm.DB {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize
	return query.Offset(offset).Limit(pageSize)
}
