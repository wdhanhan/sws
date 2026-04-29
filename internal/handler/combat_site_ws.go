package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/websocket"

	"github.com/starfall-warsong/sws/internal/service"
	"github.com/starfall-warsong/sws/pkg/auth"
)

type wsMsg struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data,omitempty"`
}

type CombatSiteWS struct {
	svc        *service.CombatSiteService
	jwtManager *auth.JWTManager
}

func NewCombatSiteWS(svc *service.CombatSiteService, jwtManager *auth.JWTManager) *CombatSiteWS {
	return &CombatSiteWS{svc: svc, jwtManager: jwtManager}
}

func (h *CombatSiteWS) Handler() websocket.Handler {
	return websocket.Handler(h.serve)
}

func (h *CombatSiteWS) serve(ws *websocket.Conn) {
	defer ws.Close()

	charID, ok := h.authenticate(ws)
	if !ok {
		return
	}

	slog.Info("ws connected", "char_id", charID)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cmdCh := make(chan wsMsg, 8)

	// Read loop: receive client commands
	go func() {
		defer cancel()
		for {
			var msg wsMsg
			if err := websocket.JSON.Receive(ws, &msg); err != nil {
				return
			}
			select {
			case cmdCh <- msg:
			case <-ctx.Done():
				return
			}
		}
	}()

	autoTicker := (*time.Ticker)(nil)
	defer func() {
		if autoTicker != nil {
			autoTicker.Stop()
		}
	}()

	// Check if there's an existing session and send state
	if result, err := h.svc.SiteNextTick(ctx, charID); err == nil {
		h.send(ws, "tick", result)
	}

	for {
		select {
		case <-ctx.Done():
			return

		case msg, ok := <-cmdCh:
			if !ok {
				return
			}
			switch msg.Type {
			case "enter":
				var req struct {
					SiteID int64 `json:"site_id"`
				}
				json.Unmarshal(msg.Data, &req)
				result, err := h.svc.EnterSite(ctx, charID, req.SiteID)
				if err != nil {
					h.sendErr(ws, err.Error())
				} else {
					h.send(ws, "entered", result)
				}

			case "tick":
				result, err := h.svc.SiteNextTick(ctx, charID)
				if err != nil {
					h.sendErr(ws, err.Error())
				} else {
					h.send(ws, "tick", result)
				}

			case "auto_start":
				if autoTicker != nil {
					autoTicker.Stop()
				}
				autoTicker = time.NewTicker(1 * time.Second)
				// Send first tick immediately
				if result, err := h.svc.SiteNextTick(ctx, charID); err == nil {
					h.send(ws, "tick", result)
					if result.Completed || result.Failed {
						autoTicker.Stop()
						autoTicker = nil
					}
				}

			case "auto_stop":
				if autoTicker != nil {
					autoTicker.Stop()
					autoTicker = nil
				}
				h.send(ws, "auto_stopped", nil)

			case "leave":
				if autoTicker != nil {
					autoTicker.Stop()
					autoTicker = nil
				}
				h.svc.LeaveSite(ctx, charID)
				h.send(ws, "left", nil)
			}

		case <-func() <-chan time.Time {
			if autoTicker != nil {
				return autoTicker.C
			}
			return nil
		}():
			result, err := h.svc.SiteNextTick(ctx, charID)
			if err != nil {
				h.sendErr(ws, err.Error())
				if autoTicker != nil {
					autoTicker.Stop()
					autoTicker = nil
				}
			} else {
				h.send(ws, "tick", result)
				if result.Completed || result.Failed {
					if autoTicker != nil {
						autoTicker.Stop()
						autoTicker = nil
					}
				}
			}
		}
	}
}

func (h *CombatSiteWS) authenticate(ws *websocket.Conn) (int64, bool) {
	// Client must send auth message first: {"type":"auth","data":{"token":"...","char_id":123}}
	ws.SetDeadline(time.Now().Add(10 * time.Second))
	var msg wsMsg
	if err := websocket.JSON.Receive(ws, &msg); err != nil || msg.Type != "auth" {
		h.sendErr(ws, "需要先发送认证消息")
		return 0, false
	}
	ws.SetDeadline(time.Time{}) // clear deadline

	var authData struct {
		Token  string `json:"token"`
		CharID int64  `json:"char_id"`
	}
	if err := json.Unmarshal(msg.Data, &authData); err != nil {
		h.sendErr(ws, "认证数据格式错误")
		return 0, false
	}

	// Validate JWT
	token := authData.Token
	if strings.HasPrefix(token, "Bearer ") {
		token = token[7:]
	}
	claims, err := h.jwtManager.ValidateToken(token)
	if err != nil || claims.TokenType != "access" {
		h.sendErr(ws, "令牌无效或已过期")
		return 0, false
	}

	charID := authData.CharID
	if charID == 0 {
		charID = claims.AccountID
	}

	h.send(ws, "authed", map[string]int64{"char_id": charID})
	return charID, true
}

func (h *CombatSiteWS) send(ws *websocket.Conn, msgType string, data any) {
	raw, _ := json.Marshal(data)
	websocket.JSON.Send(ws, wsMsg{Type: msgType, Data: raw})
}

func (h *CombatSiteWS) sendErr(ws *websocket.Conn, msg string) {
	raw, _ := json.Marshal(map[string]string{"message": msg})
	websocket.JSON.Send(ws, wsMsg{Type: "error", Data: raw})
}

// Helper to extract charID from query params (for simple auth via URL)
func charIDFromQuery(ws *websocket.Conn) int64 {
	q := ws.Request().URL.Query()
	id, _ := strconv.ParseInt(q.Get("char_id"), 10, 64)
	return id
}
