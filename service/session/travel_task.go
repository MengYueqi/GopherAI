package session

import (
	"GopherAI/common/aihelper"
	"GopherAI/common/code"
	"GopherAI/model"
	"context"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	travelTaskStatePending   = "pending"
	travelTaskStateRunning   = "running"
	travelTaskStateSucceeded = "succeeded"
	travelTaskStateFailed    = "failed"

	travelStagePending   = "pending"
	travelStageRunning   = "running"
	travelStageCompleted = "completed"
	travelStageSkipped   = "skipped"
	travelStageFailed    = "failed"
)

var travelTaskStageDefs = []struct {
	Key   string
	Label string
}{
	{Key: "feasibility_check", Label: "可行性评估"},
	{Key: "requirements_feedback", Label: "需求补充建议"},
	{Key: "overall_route", Label: "总体路线设计"},
	{Key: "flight_planning", Label: "机票信息分析"},
	{Key: "attraction_planning", Label: "景点亮点规划"},
	{Key: "plan_summary", Label: "行程汇总成文"},
	{Key: "json_structuring", Label: "结构化整理"},
}

type travelTaskManager struct {
	mu    sync.RWMutex
	tasks map[string]*model.TravelPlanningTaskSnapshot
}

var globalTravelTaskManager = &travelTaskManager{
	tasks: make(map[string]*model.TravelPlanningTaskSnapshot),
}

func StartTravelPlanningTask(description string) (model.TravelPlanningTaskSnapshot, code.Code) {
	taskID := uuid.New().String()
	now := time.Now().Unix()
	task := &model.TravelPlanningTaskSnapshot{
		TaskID:          taskID,
		State:           travelTaskStatePending,
		Description:     description,
		CurrentDetail:   "任务已创建，等待开始规划。",
		ProgressPercent: 0,
		CreatedAt:       now,
		UpdatedAt:       now,
		Stages:          buildTravelPlanningStages(),
	}

	globalTravelTaskManager.mu.Lock()
	globalTravelTaskManager.tasks[taskID] = task
	globalTravelTaskManager.mu.Unlock()

	go runTravelPlanningTask(taskID, description)

	return cloneTravelTask(task), code.CodeSuccess
}

func GetTravelPlanningTask(taskID string) (model.TravelPlanningTaskSnapshot, code.Code) {
	globalTravelTaskManager.mu.RLock()
	task, ok := globalTravelTaskManager.tasks[taskID]
	globalTravelTaskManager.mu.RUnlock()
	if !ok {
		return model.TravelPlanningTaskSnapshot{}, code.CodeInvalidParams
	}
	return cloneTravelTask(task), code.CodeSuccess
}

func runTravelPlanningTask(taskID string, description string) {
	manager := aihelper.GetGlobalManager()
	modelType := "1"
	config := map[string]interface{}{
		"apiKey": "your-api-key",
	}
	helper, err := manager.GetOrCreateAIHelper("system", "medical_advice_session", modelType, config)
	if err != nil {
		log.Println("runTravelPlanningTask GetOrCreateAIHelper error:", err)
		failTravelTask(taskID, "初始化规划助手失败。")
		return
	}

	updateTravelTask(taskID, func(task *model.TravelPlanningTaskSnapshot) {
		task.State = travelTaskStateRunning
		task.CurrentDetail = "开始生成旅行规划。"
		task.UpdatedAt = time.Now().Unix()
	})

	aiResponse, err := helper.GenerateMedicalAdviceResponseWithProgress(context.Background(), description, func(progress aihelper.TravelPlanningProgress) {
		applyTravelTaskProgress(taskID, progress)
	})
	if err != nil {
		log.Println("runTravelPlanningTask GenerateMedicalAdviceResponseWithProgress error:", err)
		failTravelTask(taskID, "生成旅行规划失败，请稍后重试。")
		return
	}

	payload := parseTravelPlanPayload(aiResponse.Content)
	updateTravelTask(taskID, func(task *model.TravelPlanningTaskSnapshot) {
		now := time.Now().Unix()
		task.State = travelTaskStateSucceeded
		task.ProgressPercent = 100
		task.Advice = payload
		task.CompletedAt = now
		task.UpdatedAt = now
		if payload.Mode == "raw" {
			task.CurrentDetail = "已生成补充建议。"
		} else {
			task.CurrentDetail = "旅行规划已生成完成。"
		}
		markPendingStagesSkipped(task)
	})
}

func failTravelTask(taskID string, message string) {
	updateTravelTask(taskID, func(task *model.TravelPlanningTaskSnapshot) {
		now := time.Now().Unix()
		task.State = travelTaskStateFailed
		task.ErrorMessage = message
		task.CurrentDetail = message
		task.CompletedAt = now
		task.UpdatedAt = now
		for i := range task.Stages {
			if task.Stages[i].Status == travelStageRunning {
				task.Stages[i].Status = travelStageFailed
				task.Stages[i].Detail = message
				task.Stages[i].FinishedAt = now
			}
		}
		markPendingStagesSkipped(task)
	})
}

func applyTravelTaskProgress(taskID string, progress aihelper.TravelPlanningProgress) {
	updateTravelTask(taskID, func(task *model.TravelPlanningTaskSnapshot) {
		now := time.Now().Unix()
		task.State = travelTaskStateRunning
		task.CurrentStage = progress.Stage
		task.CurrentStageLabel = progress.Label
		task.CurrentDetail = progress.Detail
		if progress.Percent > task.ProgressPercent {
			task.ProgressPercent = progress.Percent
		}
		task.UpdatedAt = now

		for i := range task.Stages {
			stage := &task.Stages[i]
			if stage.Key != progress.Stage {
				continue
			}
			stage.Label = progress.Label
			stage.Status = progress.Status
			stage.Detail = progress.Detail
			if progress.Status == travelStageRunning && stage.StartedAt == 0 {
				stage.StartedAt = now
			}
			if progress.Status == travelStageCompleted || progress.Status == travelStageFailed || progress.Status == travelStageSkipped {
				if stage.StartedAt == 0 {
					stage.StartedAt = now
				}
				stage.FinishedAt = now
			}
			break
		}
	})
}

func updateTravelTask(taskID string, fn func(task *model.TravelPlanningTaskSnapshot)) {
	globalTravelTaskManager.mu.Lock()
	defer globalTravelTaskManager.mu.Unlock()

	task, ok := globalTravelTaskManager.tasks[taskID]
	if !ok {
		return
	}
	fn(task)
}

func buildTravelPlanningStages() []model.TravelPlanningStage {
	stages := make([]model.TravelPlanningStage, 0, len(travelTaskStageDefs))
	for _, stage := range travelTaskStageDefs {
		stages = append(stages, model.TravelPlanningStage{
			Key:    stage.Key,
			Label:  stage.Label,
			Status: travelStagePending,
		})
	}
	return stages
}

func markPendingStagesSkipped(task *model.TravelPlanningTaskSnapshot) {
	now := time.Now().Unix()
	for i := range task.Stages {
		if task.Stages[i].Status == travelStagePending {
			task.Stages[i].Status = travelStageSkipped
			task.Stages[i].FinishedAt = now
		}
	}
}

func cloneTravelTask(task *model.TravelPlanningTaskSnapshot) model.TravelPlanningTaskSnapshot {
	if task == nil {
		return model.TravelPlanningTaskSnapshot{}
	}
	cloned := *task
	if task.Stages != nil {
		cloned.Stages = append([]model.TravelPlanningStage(nil), task.Stages...)
	}
	if task.Advice.DailyPlans != nil {
		cloned.Advice.DailyPlans = append([]model.TravelDayPlan(nil), task.Advice.DailyPlans...)
	}
	if task.Advice.Sources != nil {
		cloned.Advice.Sources = append([]string(nil), task.Advice.Sources...)
	}
	if task.Advice.FlightPrice.BookingTips != nil {
		cloned.Advice.FlightPrice.BookingTips = append([]string(nil), task.Advice.FlightPrice.BookingTips...)
	}
	return cloned
}
