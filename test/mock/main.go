package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	mockToken       = "mock-jwt-token"
	defaultModel    = "deepseek-chat"
	defaultUserName = "mock_user"
)

type response struct {
	StatusCode int    `json:"status_code"`
	StatusMsg  string `json:"status_msg,omitempty"`
}

type sessionInfo struct {
	SessionID string    `json:"sessionId"`
	Name      string    `json:"name"`
	ModelType string    `json:"modelType"`
	UpdateAt  time.Time `json:"updateAt"`
}

type historyItem struct {
	IsUser  bool   `json:"is_user"`
	Content string `json:"content"`
}

type sessionData struct {
	Info    sessionInfo
	History []historyItem
}

type mockServer struct {
	mu       sync.RWMutex
	sessions map[string]*sessionData
	tasks    map[string]map[string]any
	nextID   int
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type registerRequest struct {
	Email    string `json:"email"`
	Captcha  string `json:"captcha"`
	Password string `json:"password"`
}

type captchaRequest struct {
	Email string `json:"email"`
}

type createSessionRequest struct {
	Question    string `json:"question"`
	ModelType   string `json:"modelType"`
	UsingGoogle bool   `json:"usingGoogle"`
	UsingRAG    bool   `json:"usingRAG"`
}

type chatSendRequest struct {
	Question    string `json:"question"`
	ModelType   string `json:"modelType"`
	SessionID   string `json:"sessionId"`
	UsingGoogle bool   `json:"usingGoogle"`
	UsingRAG    bool   `json:"usingRAG"`
}

type chatHistoryRequest struct {
	SessionID string `json:"sessionId"`
}

type travelPlanRequest struct {
	Description string `json:"description"`
}

func main() {
	port := os.Getenv("MOCK_PORT")
	if port == "" {
		port = "9090"
	}

	srv := newMockServer()
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/user/register", srv.withCORS(srv.handleRegister))
	mux.HandleFunc("/api/v1/user/login", srv.withCORS(srv.handleLogin))
	mux.HandleFunc("/api/v1/user/captcha", srv.withCORS(srv.handleCaptcha))

	mux.HandleFunc("/api/v1/AI/chat/sessions", srv.withCORS(srv.auth(srv.handleSessions)))
	mux.HandleFunc("/api/v1/AI/chat/send-new-session", srv.withCORS(srv.auth(srv.handleSendNewSession)))
	mux.HandleFunc("/api/v1/AI/chat/send", srv.withCORS(srv.auth(srv.handleSend)))
	mux.HandleFunc("/api/v1/AI/chat/history", srv.withCORS(srv.auth(srv.handleHistory)))
	mux.HandleFunc("/api/v1/AI/chat/send-stream-new-session", srv.withCORS(srv.auth(srv.handleStreamNewSession)))
	mux.HandleFunc("/api/v1/AI/chat/send-stream", srv.withCORS(srv.auth(srv.handleStreamSend)))
	mux.HandleFunc("/api/v1/AI/agent/travel_plan", srv.withCORS(srv.auth(srv.handleTravelPlan)))
	mux.HandleFunc("/api/v1/AI/agent/travel_plan/tasks", srv.withCORS(srv.auth(srv.handleTravelPlanTasks)))
	mux.HandleFunc("/api/v1/AI/agent/travel_plan/tasks/", srv.withCORS(srv.auth(srv.handleTravelPlanTaskDetail)))
	mux.HandleFunc("/api/v1/image/recognize", srv.withCORS(srv.auth(srv.handleRecognizeImage)))

	addr := ":" + port
	log.Printf("mock server listening on http://127.0.0.1%s", addr)
	log.Printf("auth token: %s", mockToken)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

func newMockServer() *mockServer {
	now := time.Now().UTC()
	return &mockServer{
		sessions: map[string]*sessionData{
			"session_mock_001": {
				Info: sessionInfo{
					SessionID: "session_mock_001",
					Name:      "东京三日游攻略",
					ModelType: defaultModel,
					UpdateAt:  now.Add(-30 * time.Minute),
				},
				History: []historyItem{
					{IsUser: true, Content: "帮我规划一次东京三日游"},
					{IsUser: false, Content: "当然可以。第一天建议游览浅草寺和晴空塔，第二天前往涩谷、原宿，第三天安排上野公园和秋叶原。"},
					{IsUser: true, Content: "预算控制在 5000 元以内"},
					{IsUser: false, Content: "如果预算控制在 5000 元以内，建议选择商务酒店，优先购买地铁通票，并减少高价景点和高端餐饮安排。"},
				},
			},
		},
		tasks:  map[string]map[string]any{},
		nextID: 2,
	}
}

func (s *mockServer) withCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next(w, r)
	}
}

func (s *mockServer) auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
		if token == "" {
			token = strings.TrimSpace(r.URL.Query().Get("token"))
		}
		if token != mockToken {
			writeJSON(w, http.StatusOK, response{StatusCode: 2006, StatusMsg: "无效的Token"})
			return
		}
		next(w, r)
	}
}

func (s *mockServer) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w)
		return
	}
	var req registerRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.Email == "" {
		writeJSON(w, http.StatusOK, response{StatusCode: 2001, StatusMsg: "请求参数错误"})
		return
	}
	if req.Captcha != "" && req.Captcha != "123456" {
		writeJSON(w, http.StatusOK, response{StatusCode: 2008, StatusMsg: "验证码错误"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"status_code": 1000,
		"status_msg":  "success",
		"token":       mockToken,
	})
}

func (s *mockServer) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w)
		return
	}
	var req loginRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.Username != "mock_user" || req.Password != "mock_password" {
		writeJSON(w, http.StatusOK, response{StatusCode: 2004, StatusMsg: "用户名或密码错误"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"status_code": 1000,
		"status_msg":  "success",
		"token":       mockToken,
	})
}

func (s *mockServer) handleCaptcha(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w)
		return
	}
	var req captchaRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.Email == "" {
		writeJSON(w, http.StatusOK, response{StatusCode: 2001, StatusMsg: "请求参数错误"})
		return
	}
	writeJSON(w, http.StatusOK, response{StatusCode: 1000, StatusMsg: "success"})
}

func (s *mockServer) handleSessions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	s.mu.RLock()
	sessions := make([]sessionInfo, 0, len(s.sessions))
	for _, item := range s.sessions {
		sessions = append(sessions, item.Info)
	}
	s.mu.RUnlock()

	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].UpdateAt.After(sessions[j].UpdateAt)
	})

	writeJSON(w, http.StatusOK, map[string]any{
		"status_code": 1000,
		"status_msg":  "success",
		"sessions":    sessions,
	})
}

func (s *mockServer) handleSendNewSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w)
		return
	}
	var req createSessionRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.Question == "" || req.ModelType == "" {
		writeJSON(w, http.StatusOK, response{StatusCode: 2001, StatusMsg: "请求参数错误"})
		return
	}

	reply := buildReply(req.Question)
	sessionID, name := s.createSession(req.ModelType, req.Question, reply)

	writeJSON(w, http.StatusOK, map[string]any{
		"status_code": 1000,
		"status_msg":  "success",
		"Information": reply,
		"sessionId":   sessionID,
		"name":        name,
	})
}

func (s *mockServer) handleSend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w)
		return
	}
	var req chatSendRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.Question == "" || req.ModelType == "" || req.SessionID == "" {
		writeJSON(w, http.StatusOK, response{StatusCode: 2001, StatusMsg: "请求参数错误"})
		return
	}

	reply := buildReply(req.Question)
	if !s.appendToSession(req.SessionID, req.Question, reply, req.ModelType) {
		writeJSON(w, http.StatusOK, response{StatusCode: 2009, StatusMsg: "记录不存在"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"status_code": 1000,
		"status_msg":  "success",
		"Information": reply,
	})
}

func (s *mockServer) handleHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w)
		return
	}
	var req chatHistoryRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.SessionID == "" {
		writeJSON(w, http.StatusOK, response{StatusCode: 2001, StatusMsg: "请求参数错误"})
		return
	}

	s.mu.RLock()
	data, ok := s.sessions[req.SessionID]
	if !ok {
		s.mu.RUnlock()
		writeJSON(w, http.StatusOK, response{StatusCode: 2009, StatusMsg: "记录不存在"})
		return
	}
	history := append([]historyItem(nil), data.History...)
	s.mu.RUnlock()

	writeJSON(w, http.StatusOK, map[string]any{
		"status_code": 1000,
		"status_msg":  "success",
		"history":     history,
	})
}

func (s *mockServer) handleTravelPlan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w)
		return
	}
	var req travelPlanRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.Description == "" {
		writeJSON(w, http.StatusOK, response{StatusCode: 2001, StatusMsg: "请求参数错误"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"status_code": 1000,
		"status_msg":  "success",
		"plan":        buildTravelPlanMock(req.Description),
	})
}

func (s *mockServer) handleTravelPlanTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w)
		return
	}
	var req travelPlanRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.Description == "" {
		writeJSON(w, http.StatusOK, response{StatusCode: 2001, StatusMsg: "请求参数错误"})
		return
	}

	taskID := fmt.Sprintf("travel_task_%d", time.Now().UnixNano())
	task := map[string]any{
		"task_id":             taskID,
		"state":               "succeeded",
		"description":         req.Description,
		"current_stage":       "json_structuring",
		"current_stage_label": "结构化整理",
		"current_detail":      "旅行规划已生成完成。",
		"progress_percent":    100,
		"stages": []map[string]any{
			{"key": "feasibility_check", "label": "可行性评估", "status": "completed", "detail": "需求可行性判断已完成。"},
			{"key": "overall_route", "label": "总体路线设计", "status": "completed", "detail": "总体路线已生成。"},
			{"key": "flight_planning", "label": "机票信息分析", "status": "completed", "detail": "机票建议已生成。"},
			{"key": "attraction_planning", "label": "景点亮点规划", "status": "completed", "detail": "景点亮点内容已整理。"},
			{"key": "plan_summary", "label": "行程汇总成文", "status": "completed", "detail": "汇总摘要已生成。"},
			{"key": "json_structuring", "label": "结构化整理", "status": "completed", "detail": "结构化结果已生成。"},
		},
		"plan": buildTravelPlanMock(req.Description),
	}

	s.mu.Lock()
	s.tasks[taskID] = task
	s.mu.Unlock()

	writeJSON(w, http.StatusOK, map[string]any{
		"status_code": 1000,
		"status_msg":  "success",
		"task":        task,
	})
}

func (s *mockServer) handleTravelPlanTaskDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	taskID := strings.TrimPrefix(r.URL.Path, "/api/v1/AI/agent/travel_plan/tasks/")
	if taskID == "" {
		writeJSON(w, http.StatusOK, response{StatusCode: 2001, StatusMsg: "请求参数错误"})
		return
	}

	s.mu.RLock()
	task, ok := s.tasks[taskID]
	s.mu.RUnlock()
	if !ok {
		writeJSON(w, http.StatusOK, response{StatusCode: 2001, StatusMsg: "请求参数错误"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"status_code": 1000,
		"status_msg":  "success",
		"task":        task,
	})
}

func (s *mockServer) handleRecognizeImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w)
		return
	}
	file, header, err := r.FormFile("image")
	if err != nil {
		writeJSON(w, http.StatusOK, response{StatusCode: 2001, StatusMsg: "请求参数错误"})
		return
	}
	defer file.Close()

	className := guessClassName(header)
	writeJSON(w, http.StatusOK, map[string]any{
		"status_code": 1000,
		"status_msg":  "success",
		"class_name":  className,
	})
}

func (s *mockServer) handleStreamNewSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w)
		return
	}
	var req createSessionRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.Question == "" || req.ModelType == "" {
		writeJSON(w, http.StatusOK, response{StatusCode: 2001, StatusMsg: "请求参数错误"})
		return
	}

	reply := buildReply(req.Question)
	sessionID, _ := s.createSession(req.ModelType, req.Question, reply)
	chunks := splitReply(reply)

	writeSSEHeaders(w)
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "stream unsupported", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "data: {\"sessionId\":\"%s\"}\n\n", sessionID)
	flusher.Flush()
	for _, chunk := range chunks {
		time.Sleep(250 * time.Millisecond)
		fmt.Fprintf(w, "data: {\"content\":%q}\n\n", chunk)
		flusher.Flush()
	}
	time.Sleep(150 * time.Millisecond)
	fmt.Fprint(w, "data: [DONE]\n\n")
	flusher.Flush()
}

func (s *mockServer) handleStreamSend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w)
		return
	}
	var req chatSendRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.Question == "" || req.ModelType == "" || req.SessionID == "" {
		writeJSON(w, http.StatusOK, response{StatusCode: 2001, StatusMsg: "请求参数错误"})
		return
	}

	reply := buildReply(req.Question)
	if !s.appendToSession(req.SessionID, req.Question, reply, req.ModelType) {
		writeSSEHeaders(w)
		if flusher, ok := w.(http.Flusher); ok {
			fmt.Fprint(w, "event: error\n")
			fmt.Fprint(w, "data: {\"message\":\"Failed to send message\"}\n\n")
			flusher.Flush()
		}
		return
	}

	chunks := splitReply(reply)
	writeSSEHeaders(w)
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "stream unsupported", http.StatusInternalServerError)
		return
	}
	for _, chunk := range chunks {
		time.Sleep(250 * time.Millisecond)
		fmt.Fprintf(w, "data: {\"content\":%q}\n\n", chunk)
		flusher.Flush()
	}
	time.Sleep(150 * time.Millisecond)
	fmt.Fprint(w, "data: [DONE]\n\n")
	flusher.Flush()
}

func (s *mockServer) createSession(modelType, question, reply string) (string, string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextID++
	sessionID := fmt.Sprintf("session_mock_%03d", s.nextID)
	name := truncateTitle(question)
	if modelType == "" {
		modelType = defaultModel
	}
	s.sessions[sessionID] = &sessionData{
		Info: sessionInfo{
			SessionID: sessionID,
			Name:      name,
			ModelType: modelType,
			UpdateAt:  time.Now().UTC(),
		},
		History: []historyItem{
			{IsUser: true, Content: question},
			{IsUser: false, Content: reply},
		},
	}
	return sessionID, name
}

func (s *mockServer) appendToSession(sessionID, question, reply, modelType string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, ok := s.sessions[sessionID]
	if !ok {
		return false
	}
	data.History = append(data.History,
		historyItem{IsUser: true, Content: question},
		historyItem{IsUser: false, Content: reply},
	)
	data.Info.UpdateAt = time.Now().UTC()
	if modelType != "" {
		data.Info.ModelType = modelType
	}
	return true
}

func buildReply(question string) string {
	switch {
	case strings.Contains(question, "东京"), strings.Contains(strings.ToLower(question), "travel"):
		return "当然可以。第一天建议游览浅草寺和晴空塔，第二天前往涩谷、原宿，第三天安排上野公园和秋叶原。"
	case strings.Contains(question, "预算"), strings.Contains(question, "5000"):
		return "如果预算控制在 5000 元以内，建议选择商务酒店，优先购买地铁通票，并减少高价景点和高端餐饮安排。"
	case strings.Contains(question, "感冒"), strings.Contains(question, "发烧"), strings.Contains(question, "咳嗽"):
		return "建议先充分休息并补充水分，可监测体温变化。如持续高烧、呼吸困难或症状加重，应尽快就医。"
	default:
		return "这是一个 mock 响应，用于前端联调。你可以继续发送问题来测试会话、历史记录和流式输出。"
	}
}

func splitReply(reply string) []string {
	parts := strings.FieldsFunc(reply, func(r rune) bool {
		return r == '，' || r == '。' || r == ',' || r == '.'
	})
	chunks := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			chunks = append(chunks, part)
		}
	}
	if len(chunks) == 0 {
		return []string{reply}
	}
	return chunks
}

func truncateTitle(question string) string {
	rs := []rune(strings.TrimSpace(question))
	if len(rs) == 0 {
		return "新的对话"
	}
	if len(rs) > 12 {
		return string(rs[:12])
	}
	return string(rs)
}

func guessClassName(header *multipart.FileHeader) string {
	name := strings.ToLower(header.Filename)
	switch {
	case strings.Contains(name, "cat"):
		return "cat"
	case strings.Contains(name, "dog"):
		return "golden_retriever"
	case strings.Contains(name, "flower"):
		return "sunflower"
	default:
		return "golden_retriever"
	}
}

func buildTravelPlanMock(description string) map[string]any {
	overallSummary := "东京 3 日行程以经典城市地标、商业街区和文化体验为主，节奏中等，适合第一次到东京旅行的用户。"
	if strings.Contains(description, "大阪") || strings.Contains(strings.ToLower(description), "kansai") {
		overallSummary = "关西 3 日行程以大阪城市体验为主，兼顾美食、商业街区与经典地标，适合首次体验关西都市风格的用户。"
	}

	dailyPlans := []map[string]any{
		{
			"day":       1,
			"title":     "浅草与东京晴空塔",
			"route":     "浅草寺 -> 仲见世商店街 -> 隅田公园 -> 东京晴空塔",
			"transport": "地铁 + 步行",
			"summary":   "第一天适合从东京传统街区开始，感受寺庙文化与城市天际线。",
			"attractions": []map[string]any{
				{
					"name":        "浅草寺",
					"description": "东京代表性的历史寺庙，适合体验传统建筑、参拜文化与街区氛围。",
					"highlights":  []string{"雷门地标", "传统参道氛围", "适合拍照与体验和风街景"},
					"images": []map[string]any{
						{
							"title":      "浅草寺正门",
							"url":        "https://images.unsplash.com/photo-1542051841857-5f90071e7989",
							"source":     "Unsplash",
							"source_url": "https://unsplash.com/photos/OQMZwNd3ThU",
						},
						{
							"title":      "东京浅草街景",
							"url":        "https://images.unsplash.com/photo-1513407030348-c983a97b98d8",
							"source":     "Unsplash",
							"source_url": "https://unsplash.com/photos/6I2dlBS1ewg",
						},
					},
				},
				{
					"name":        "东京晴空塔",
					"description": "俯瞰东京城市景观的标志性塔楼，夜景和黄昏时段尤其适合安排。",
					"highlights":  []string{"东京全景", "夜景优秀", "适合与浅草区域联动安排"},
					"images": []map[string]any{
						{
							"title":      "东京晴空塔远景",
							"url":        "https://images.unsplash.com/photo-1536098561742-ca998e48cbcc",
							"source":     "Unsplash",
							"source_url": "https://unsplash.com/photos/twukN12EN7c",
						},
					},
				},
			},
			"tips": []string{"浅草区域建议上午前往，人流更可控", "晴空塔建议提前预约傍晚时段"},
		},
		{
			"day":       2,
			"title":     "涩谷与原宿城市活力线",
			"route":     "涩谷十字路口 -> SHIBUYA SKY -> 表参道 -> 原宿竹下通",
			"transport": "JR + 步行",
			"summary":   "第二天以东京现代商业和潮流文化为主，适合逛街、拍照和夜景体验。",
			"attractions": []map[string]any{
				{
					"name":        "涩谷十字路口",
					"description": "东京都市感最强的代表场景之一，适合感受城市节奏。",
					"highlights":  []string{"城市地标", "人流视效强", "适合夜景与街头摄影"},
					"images": []map[string]any{
						{
							"title":      "涩谷街头夜景",
							"url":        "https://images.unsplash.com/photo-1540959733332-eab4deabeeaf",
							"source":     "Unsplash",
							"source_url": "https://unsplash.com/photos/JmuyB_LibRo",
						},
					},
				},
				{
					"name":        "原宿竹下通",
					"description": "适合体验东京年轻人文化、美食小店与潮流消费。",
					"highlights":  []string{"年轻潮流文化", "街头小吃多", "适合轻松步行游览"},
					"images": []map[string]any{
						{
							"title":      "原宿街区",
							"url":        "https://images.unsplash.com/photo-1554797589-7241bb691973",
							"source":     "Unsplash",
							"source_url": "https://unsplash.com/photos/2cFZ_FB08UM",
						},
					},
				},
			},
			"tips": []string{"涩谷建议安排到下午或傍晚", "原宿逛街建议避开周末中午高峰"},
		},
		{
			"day":       3,
			"title":     "上野与秋叶原文化收尾",
			"route":     "上野公园 -> 东京国立博物馆 -> 阿美横町 -> 秋叶原",
			"transport": "地铁 + 步行",
			"summary":   "第三天以博物馆、公园和二次元电子街区作为收尾，兼具文化与购物。",
			"attractions": []map[string]any{
				{
					"name":        "上野公园",
					"description": "东京经典城市公园，周边集中了博物馆和休闲空间。",
					"highlights":  []string{"博物馆集中", "散步舒适", "适合作为轻松安排"},
					"images": []map[string]any{
						{
							"title":      "上野公园景观",
							"url":        "https://images.unsplash.com/photo-1526481280695-3c4691f7f66c",
							"source":     "Unsplash",
							"source_url": "https://unsplash.com/photos/7Zjq7GxLqEc",
						},
					},
				},
				{
					"name":        "秋叶原",
					"description": "适合购买电子产品、动漫周边和体验东京亚文化。",
					"highlights":  []string{"电器购物", "动漫文化", "适合夜间街景体验"},
					"images": []map[string]any{
						{
							"title":      "秋叶原夜景",
							"url":        "https://images.unsplash.com/photo-1503899036084-c55cdd92da26",
							"source":     "Unsplash",
							"source_url": "https://unsplash.com/photos/W7b3eDUb_2I",
						},
					},
				},
			},
			"tips": []string{"博物馆建议提前查闭馆日", "秋叶原适合安排在傍晚以后"},
		},
	}

	return map[string]any{
		"mode":            "plan",
		"overall_summary": overallSummary,
		"flight_price": map[string]any{
			"summary":      "往返东京机票通常在淡季更划算，建议优先关注直飞与中转时长之间的平衡。",
			"currency":     "CNY",
			"price_range":  "1800-2600",
			"booking_tips": []string{"建议提前 2 到 4 周关注价格波动", "若预算敏感，可优先考虑非黄金时段航班"},
			"raw_text":     "Mock 航班价格区间：1800-2600 CNY，直飞更省时，中转更省预算。",
		},
		"daily_plans": dailyPlans,
		"sources": []string{
			"https://unsplash.com/photos/OQMZwNd3ThU",
			"https://unsplash.com/photos/6I2dlBS1ewg",
			"https://unsplash.com/photos/twukN12EN7c",
			"https://unsplash.com/photos/JmuyB_LibRo",
			"https://unsplash.com/photos/2cFZ_FB08UM",
			"https://unsplash.com/photos/7Zjq7GxLqEc",
			"https://unsplash.com/photos/W7b3eDUb_2I",
		},
		"notice":   "这是 mock 返回的结构化旅游方案，用于前端联调。",
		"raw_text": "",
	}
}

func writeSSEHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
}

func decodeJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusOK, response{StatusCode: 2001, StatusMsg: "请求参数错误"})
		return false
	}
	if err := json.Unmarshal(body, v); err != nil {
		writeJSON(w, http.StatusOK, response{StatusCode: 2001, StatusMsg: "请求参数错误"})
		return false
	}
	return true
}

func writeMethodNotAllowed(w http.ResponseWriter) {
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
