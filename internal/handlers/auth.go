package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"go-devops/internal/config"
	"go-devops/internal/logger"
	"go-devops/internal/models"
	"gorm.io/gorm"
)

type AuthHandler struct {
	db *gorm.DB
}

func NewAuthHandler(db *gorm.DB) *AuthHandler {
	return &AuthHandler{db: db}
}

// 用户登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warnf("登录请求参数错误: %v, IP: %s", err, c.ClientIP())
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	logger.Infof("用户登录尝试: %s, IP: %s", req.Username, c.ClientIP())

	var user models.User
	if err := h.db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		logger.Warnf("登录失败 - 用户不存在: %s, IP: %s", req.Username, c.ClientIP())
		logger.LogUserAction(0, req.Username, "login", "auth", false, "用户不存在")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		logger.Warnf("登录失败 - 密码错误: %s, IP: %s", req.Username, c.ClientIP())
		logger.LogUserAction(user.ID, req.Username, "login", "auth", false, "密码错误")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	token, err := h.generateToken(&user)
	if err != nil {
		logger.Errorf("生成令牌失败: %v, 用户: %s", err, req.Username)
		logger.LogUserAction(user.ID, req.Username, "login", "auth", false, "令牌生成失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成令牌失败"})
		return
	}

	logger.Infof("用户登录成功: %s, IP: %s", req.Username, c.ClientIP())
	logger.LogUserAction(user.ID, req.Username, "login", "auth", true, "登录成功")

	c.JSON(http.StatusOK, models.AuthResponse{
		Token: token,
		User:  user,
	})
}

// 用户注册
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warnf("注册请求参数错误: %v, IP: %s", err, c.ClientIP())
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	logger.Infof("用户注册尝试: %s, 邮箱: %s, IP: %s", req.Username, req.Email, c.ClientIP())

	// 检查用户名是否已存在
	var existingUser models.User
	if err := h.db.Where("username = ? OR email = ?", req.Username, req.Email).First(&existingUser).Error; err == nil {
		logger.Warnf("注册失败 - 用户名或邮箱已存在: %s, %s, IP: %s", req.Username, req.Email, c.ClientIP())
		logger.LogUserAction(0, req.Username, "register", "auth", false, "用户名或邮箱已存在")
		c.JSON(http.StatusConflict, gin.H{"error": "用户名或邮箱已存在"})
		return
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Errorf("密码加密失败: %v, 用户: %s", err, req.Username)
		logger.LogUserAction(0, req.Username, "register", "auth", false, "密码加密失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}

	user := models.User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
		Role:     "user",
	}

	if err := h.db.Create(&user).Error; err != nil {
		logger.Errorf("创建用户失败: %v, 用户: %s", err, req.Username)
		logger.LogDBOperation("create", "users", false, err.Error())
		logger.LogUserAction(0, req.Username, "register", "auth", false, "数据库创建失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建用户失败"})
		return
	}

	logger.LogDBOperation("create", "users", true, "")

	token, err := h.generateToken(&user)
	if err != nil {
		logger.Errorf("生成令牌失败: %v, 用户: %s", err, req.Username)
		logger.LogUserAction(user.ID, req.Username, "register", "auth", false, "令牌生成失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成令牌失败"})
		return
	}

	logger.Infof("用户注册成功: %s, ID: %d, IP: %s", req.Username, user.ID, c.ClientIP())
	logger.LogUserAction(user.ID, req.Username, "register", "auth", true, "注册成功")

	c.JSON(http.StatusCreated, models.AuthResponse{
		Token: token,
		User:  user,
	})
}

// 获取用户信息
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID := c.GetUint("user_id")
	
	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// 更新用户信息
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID := c.GetUint("user_id")
	
	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	var updateData map[string]interface{}
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	// 如果要更新密码，需要加密
	if password, exists := updateData["password"]; exists {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password.(string)), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
			return
		}
		updateData["password"] = string(hashedPassword)
	}

	if err := h.db.Model(&user).Updates(updateData).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新用户信息失败"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// 获取所有用户（管理员）
func (h *AuthHandler) GetUsers(c *gin.Context) {
	var users []models.User
	if err := h.db.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户列表失败"})
		return
	}

	c.JSON(http.StatusOK, users)
}

// 删除用户（管理员）
func (h *AuthHandler) DeleteUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	if err := h.db.Delete(&models.User{}, uint(id)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除用户失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "用户删除成功"})
}

// 更新用户角色（管理员）
func (h *AuthHandler) UpdateUserRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	var req struct {
		Role string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	if err := h.db.Model(&models.User{}).Where("id = ?", uint(id)).Update("role", req.Role).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新用户角色失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "用户角色更新成功"})
}

// 修改密码
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID := c.GetUint("user_id")
	
	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warnf("修改密码请求参数错误: %v, 用户ID: %d", err, userID)
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		logger.Warnf("修改密码失败 - 旧密码错误: 用户ID: %d, IP: %s", userID, c.ClientIP())
		logger.LogUserAction(userID, user.Username, "change_password", "auth", false, "旧密码错误")
		c.JSON(http.StatusBadRequest, gin.H{"error": "旧密码错误"})
		return
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		logger.Errorf("新密码加密失败: %v, 用户ID: %d", err, userID)
		logger.LogUserAction(userID, user.Username, "change_password", "auth", false, "密码加密失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}

	// 更新密码
	if err := h.db.Model(&user).Update("password", string(hashedPassword)).Error; err != nil {
		logger.Errorf("更新密码失败: %v, 用户ID: %d", err, userID)
		logger.LogDBOperation("update", "users", false, err.Error())
		logger.LogUserAction(userID, user.Username, "change_password", "auth", false, "数据库更新失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新密码失败"})
		return
	}

	logger.Infof("用户修改密码成功: %s, ID: %d, IP: %s", user.Username, userID, c.ClientIP())
	logger.LogDBOperation("update", "users", true, "")
	logger.LogUserAction(userID, user.Username, "change_password", "auth", true, "密码修改成功")

	c.JSON(http.StatusOK, gin.H{"message": "密码修改成功"})
}

// 获取用户统计信息
func (h *AuthHandler) GetUserStats(c *gin.Context) {
	userID := c.GetUint("user_id")
	
	var stats struct {
		ScriptCount    int64 `json:"script_count"`
		ExecutionCount int64 `json:"execution_count"`
		LastLoginDays  int   `json:"last_login_days"`
	}
	
	// 获取用户创建的脚本数量
	h.db.Model(&models.Script{}).Where("created_by = ?", userID).Count(&stats.ScriptCount)
	
	// 获取用户的执行次数
	h.db.Model(&models.JobExecution{}).Where("executed_by = ?", userID).Count(&stats.ExecutionCount)
	
	// 获取用户信息计算最后登录天数
	var user models.User
	if err := h.db.First(&user, userID).Error; err == nil {
		lastUpdate := user.UpdatedAt
		daysSince := int((time.Now().Sub(lastUpdate)).Hours() / 24)
		stats.LastLoginDays = daysSince
	}
	
	c.JSON(http.StatusOK, stats)
}

// 生成JWT令牌
func (h *AuthHandler) generateToken(user *models.User) (string, error) {
	cfg := config.Load()
	
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWTSecret))
}
