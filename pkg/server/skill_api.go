package server

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/tianniu-ai/tianniu/pkg/skill"
)

type SkillAPI struct {
	skillManager *skill.Manager
}

func NewSkillAPI(skillManager *skill.Manager) *SkillAPI {
	return &SkillAPI{skillManager: skillManager}
}

func (api *SkillAPI) RegisterRoutes(g *gin.RouterGroup) {
	g.GET("/skills", api.listSkillsForUser)
	g.GET("/skills/system", api.listSystemSkills)
	g.GET("/skills/user", api.listUserSkills)
	g.GET("/skills/:id", api.getSkill)
	g.POST("/skills/install", api.installUserSkill)
	g.POST("/skills/:id/uninstall", api.uninstallSkill)
	g.POST("/skills/:id/enable", api.enableSkill)
	g.POST("/skills/:id/disable", api.disableSkill)
}

func (api *SkillAPI) getUserID(c *gin.Context) string {
	userID, exists := c.Get("userID")
	if !exists {
		return ""
	}
	return userID.(string)
}

func (api *SkillAPI) listSkillsForUser(c *gin.Context) {
	userID := api.getUserID(c)
	if userID == "" {
		respondError(c, StatusUsernameError, nil)
		return
	}

	skills, err := api.skillManager.GetSkillsForUser(userID)
	if err != nil {
		respondError(c, StatusInternalServerError, err)
		return
	}
	respondSuccess(c, skills)
}

func (api *SkillAPI) listSystemSkills(c *gin.Context) {
	skills, err := api.skillManager.GetSystemSkills()
	if err != nil {
		respondError(c, StatusInternalServerError, err)
		return
	}
	respondSuccess(c, skills)
}

func (api *SkillAPI) listUserSkills(c *gin.Context) {
	userID := api.getUserID(c)
	if userID == "" {
		respondError(c, StatusUsernameError, nil)
		return
	}

	skills, err := api.skillManager.GetUserSkills(userID)
	if err != nil {
		respondError(c, StatusInternalServerError, err)
		return
	}
	respondSuccess(c, skills)
}

func (api *SkillAPI) getSkill(c *gin.Context) {
	id := c.Param("id")
	skillData, err := api.skillManager.GetSkillByID(id)
	if err != nil {
		respondError(c, StatusInternalServerError, err)
		return
	}
	respondSuccess(c, skillData)
}

func (api *SkillAPI) installUserSkill(c *gin.Context) {
	userID := api.getUserID(c)
	if userID == "" {
		respondError(c, StatusUsernameError, nil)
		return
	}

	file, err := c.FormFile("skill_file")
	if err != nil {
		respondError(c, StatusInvalidParam, fmt.Errorf("skill file is required"))
		return
	}

	force := false
	if forceStr := c.PostForm("force"); forceStr == "true" || forceStr == "1" {
		force = true
	}

	fileContent, err := file.Open()
	if err != nil {
		respondError(c, StatusInternalServerError, fmt.Errorf("failed to open uploaded file: %v", err))
		return
	}
	defer fileContent.Close()

	content := make([]byte, file.Size)
	if _, err := fileContent.Read(content); err != nil {
		respondError(c, StatusInternalServerError, fmt.Errorf("failed to read uploaded file: %v", err))
		return
	}

	options := skill.InstallOptions{Force: force}
	installedSkill, err := api.skillManager.InstallUserSkillFromContent(userID, string(content), options)
	if err != nil {
		respondError(c, StatusInternalServerError, err)
		return
	}
	respondSuccess(c, installedSkill)
}

type UninstallSkillRequest struct {
	KeepConfig bool `json:"keep_config"`
}

func (api *SkillAPI) uninstallSkill(c *gin.Context) {
	id := c.Param("id")
	var req UninstallSkillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, StatusInvalidParam, err)
		return
	}

	options := skill.UninstallOptions{KeepConfig: req.KeepConfig}
	if err := api.skillManager.Uninstall(id, options); err != nil {
		respondError(c, StatusInternalServerError, err)
		return
	}
	respondSuccess(c, nil)
}

func (api *SkillAPI) enableSkill(c *gin.Context) {
	id := c.Param("id")
	if err := api.skillManager.Enable(id); err != nil {
		respondError(c, StatusInternalServerError, err)
		return
	}
	respondSuccess(c, nil)
}

func (api *SkillAPI) disableSkill(c *gin.Context) {
	id := c.Param("id")
	if err := api.skillManager.Disable(id); err != nil {
		respondError(c, StatusInternalServerError, err)
		return
	}
	respondSuccess(c, nil)
}
