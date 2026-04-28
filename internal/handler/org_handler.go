package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/starfall-warsong/sws/internal/service"
	"github.com/starfall-warsong/sws/pkg/response"
)

type OrgHandler struct {
	corpSvc *service.CorpService
}

func NewOrgHandler(corpSvc *service.CorpService) *OrgHandler {
	return &OrgHandler{corpSvc: corpSvc}
}

func (h *OrgHandler) CreateCorp(c *gin.Context) {
	charID := getCharIDFromContext(c)
	var req service.CreateCorpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	corp, err := h.corpSvc.Create(c.Request.Context(), charID, &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, corp)
}

func (h *OrgHandler) JoinCorp(c *gin.Context) {
	charID := getCharIDFromContext(c)
	corpID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的军团ID")
		return
	}
	if err := h.corpSvc.Join(c.Request.Context(), charID, corpID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, gin.H{"message": "已加入军团"})
}

func (h *OrgHandler) LeaveCorp(c *gin.Context) {
	charID := getCharIDFromContext(c)
	if err := h.corpSvc.Leave(c.Request.Context(), charID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, gin.H{"message": "已离开军团"})
}

func (h *OrgHandler) GetCorpInfo(c *gin.Context) {
	corpID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的军团ID")
		return
	}
	corp, err := h.corpSvc.GetInfo(c.Request.Context(), corpID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	members, _ := h.corpSvc.GetMembers(c.Request.Context(), corpID)
	response.OK(c, gin.H{"corp": corp, "members": members})
}

func (h *OrgHandler) GetMyCorp(c *gin.Context) {
	charID := getCharIDFromContext(c)
	corp, role, err := h.corpSvc.GetCharCorp(c.Request.Context(), charID)
	if err != nil {
		response.NotFound(c, "你还没有加入任何军团")
		return
	}
	response.OK(c, gin.H{"corp": corp, "your_role": role})
}

func (h *OrgHandler) CreateAlliance(c *gin.Context) {
	charID := getCharIDFromContext(c)
	var req struct {
		Name   string `json:"name" binding:"required"`
		Ticker string `json:"ticker" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	allianceID, err := h.corpSvc.CreateAlliance(c.Request.Context(), charID, req.Name, req.Ticker)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, gin.H{"alliance_id": allianceID, "message": "联盟创建成功"})
}

// 聊天
func (h *OrgHandler) SendChat(c *gin.Context) {
	charID := getCharIDFromContext(c)
	var req struct {
		Channel string `json:"channel" binding:"required"`
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	var name string
	getDB(h).GetContext(c.Request.Context(), &name, `SELECT name FROM characters WHERE id = $1`, charID)
	if name == "" {
		name = "未知玩家"
	}

	_, err := getDB(h).ExecContext(c.Request.Context(),
		`INSERT INTO chat_messages (channel, sender_id, sender_name, content) VALUES ($1, $2, $3, $4)`,
		req.Channel, charID, name, req.Content)
	if err != nil {
		response.InternalError(c, "发送失败")
		return
	}
	response.OK(c, gin.H{"message": "已发送"})
}

func (h *OrgHandler) GetChat(c *gin.Context) {
	channel := c.Query("channel")
	if channel == "" {
		response.BadRequest(c, "需要指定频道")
		return
	}
	limit := 50
	if l, err := strconv.Atoi(c.Query("limit")); err == nil && l > 0 {
		limit = l
	}

	type Msg struct {
		ID         int64  `db:"id" json:"id"`
		SenderName string `db:"sender_name" json:"sender_name"`
		Content    string `db:"content" json:"content"`
		CreatedAt  string `db:"created_at" json:"created_at"`
	}
	var msgs []Msg
	getDB(h).SelectContext(c.Request.Context(), &msgs,
		`SELECT id, sender_name, content, created_at FROM chat_messages
		 WHERE channel = $1 ORDER BY created_at DESC LIMIT $2`, channel, limit)

	response.OK(c, gin.H{"messages": msgs, "channel": channel})
}

// SendMail sends a mail
func (h *OrgHandler) SendMail(c *gin.Context) {
	charID := getCharIDFromContext(c)
	var req struct {
		ToID    int64  `json:"to_id" binding:"required"`
		Subject string `json:"subject" binding:"required"`
		Body    string `json:"body" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	var fromName string
	getDB(h).GetContext(c.Request.Context(), &fromName, `SELECT name FROM characters WHERE id = $1`, charID)

	getDB(h).ExecContext(c.Request.Context(),
		`INSERT INTO mails (from_id, from_name, to_id, subject, body) VALUES ($1, $2, $3, $4, $5)`,
		charID, fromName, req.ToID, req.Subject, req.Body)

	response.OK(c, gin.H{"message": "邮件已发送"})
}

func (h *OrgHandler) GetMails(c *gin.Context) {
	charID := getCharIDFromContext(c)
	type Mail struct {
		ID       int64  `db:"id" json:"id"`
		FromName string `db:"from_name" json:"from_name"`
		Subject  string `db:"subject" json:"subject"`
		IsRead   bool   `db:"is_read" json:"is_read"`
		CreatedAt string `db:"created_at" json:"created_at"`
	}
	var mails []Mail
	getDB(h).SelectContext(c.Request.Context(), &mails,
		`SELECT id, from_name, subject, is_read, created_at FROM mails WHERE to_id = $1 ORDER BY created_at DESC LIMIT 50`,
		charID)
	response.OK(c, gin.H{"mails": mails, "count": len(mails)})
}

// helper to access db through corp service (simplified)
func getDB(h *OrgHandler) *sqlx.DB {
	// Access db through reflection-free approach
	return h.corpSvc.DB()
}
