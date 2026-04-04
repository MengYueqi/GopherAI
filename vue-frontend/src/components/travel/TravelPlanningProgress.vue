<template>
  <section class="progress-board">
    <div class="board-header">
      <div>
        <p class="board-eyebrow">行程推演台</p>
        <h3>{{ titleText }}</h3>
      </div>
      <div class="board-meta">
        <span class="progress-pill">{{ safePercent }}%</span>
        <span class="time-pill">已用时 {{ elapsedSeconds }}秒</span>
      </div>
    </div>

    <p class="board-detail">{{ detailText }}</p>

    <div class="progress-track">
      <div class="progress-track-fill" :style="{ width: `${safePercent}%` }"></div>
    </div>

    <div class="stage-list">
      <article
        v-for="stage in stages"
        :key="stage.key"
        class="stage-card"
        :class="stage.status || 'pending'"
      >
        <span class="stage-index">{{ stageIndex(stage.key) }}</span>
        <div class="stage-copy">
          <div class="stage-title-row">
            <strong>{{ stage.label }}</strong>
            <span class="stage-status">{{ statusText(stage.status) }}</span>
          </div>
          <p>{{ stage.detail || fallbackDetail(stage.status) }}</p>
        </div>
      </article>
    </div>
  </section>
</template>

<script>
import { computed } from 'vue'

export default {
  name: 'TravelPlanningProgress',
  props: {
    task: {
      type: Object,
      default: () => ({})
    },
    elapsedSeconds: {
      type: Number,
      default: 0
    }
  },
  setup(props) {
    const stages = computed(() => Array.isArray(props.task?.stages) ? props.task.stages : [])
    const safePercent = computed(() => {
      const value = Number(props.task?.progress_percent || 0)
      return Math.max(0, Math.min(100, value))
    })

    const titleText = computed(() => {
      if (props.task?.state === 'failed') return '规划执行失败'
      if (props.task?.state === 'succeeded') return '规划执行完成'
      return props.task?.current_stage_label || '正在启动规划流程'
    })

    const detailText = computed(() => {
      return props.task?.error_message || props.task?.current_detail || '系统正在整理行程信息。'
    })

    const statusText = (status) => {
      if (status === 'completed') return '已完成'
      if (status === 'running') return '进行中'
      if (status === 'failed') return '失败'
      if (status === 'skipped') return '已跳过'
      return '待执行'
    }

    const fallbackDetail = (status) => {
      if (status === 'completed') return '该阶段已执行完毕。'
      if (status === 'running') return '当前正在处理中。'
      if (status === 'skipped') return '本次规划未进入该阶段。'
      if (status === 'failed') return '该阶段执行失败。'
      return '等待前置阶段完成。'
    }

    const stageIndex = (key) => {
      const index = stages.value.findIndex(stage => stage.key === key)
      return `${index + 1}`.padStart(2, '0')
    }

    return {
      stages,
      safePercent,
      titleText,
      detailText,
      statusText,
      fallbackDetail,
      stageIndex
    }
  }
}
</script>

<style scoped>
.progress-board {
  border: 1px solid rgba(157, 41, 51, 0.16);
  border-radius: 24px;
  padding: 24px;
  background:
    radial-gradient(circle at top right, rgba(212, 175, 55, 0.16), transparent 34%),
    linear-gradient(135deg, rgba(157, 41, 51, 0.08), rgba(255, 250, 242, 0.96));
  box-shadow: 0 20px 45px rgba(70, 38, 24, 0.08);
}

.board-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 16px;
}

.board-eyebrow {
  margin: 0 0 10px;
  color: var(--primary);
  letter-spacing: 0.18em;
  font-size: 12px;
}

.board-header h3 {
  margin: 0;
  font-family: 'Noto Serif SC', serif;
  font-size: 26px;
  color: var(--secondary);
}

.board-meta {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
  justify-content: flex-end;
}

.progress-pill,
.time-pill {
  border-radius: 999px;
  padding: 8px 14px;
  background: rgba(255, 255, 255, 0.74);
  border: 1px solid rgba(157, 41, 51, 0.12);
  color: var(--secondary);
  font-size: 13px;
}

.progress-pill {
  background: rgba(157, 41, 51, 0.9);
  color: #fff;
}

.board-detail {
  margin: 14px 0 18px;
  line-height: 1.8;
  color: rgba(52, 38, 33, 0.86);
}

.progress-track {
  position: relative;
  height: 12px;
  border-radius: 999px;
  overflow: hidden;
  background: rgba(157, 41, 51, 0.08);
}

.progress-track-fill {
  height: 100%;
  border-radius: inherit;
  background: linear-gradient(90deg, #9d2933, #d4af37);
  transition: width 0.35s ease;
}

.stage-list {
  margin-top: 20px;
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14px;
}

.stage-card {
  display: flex;
  gap: 14px;
  border-radius: 18px;
  padding: 16px;
  background: rgba(255, 255, 255, 0.72);
  border: 1px solid rgba(28, 28, 28, 0.08);
  transition: transform 0.25s ease, box-shadow 0.25s ease, border-color 0.25s ease;
}

.stage-card.running {
  border-color: rgba(157, 41, 51, 0.36);
  box-shadow: 0 10px 24px rgba(157, 41, 51, 0.12);
  transform: translateY(-1px);
}

.stage-card.completed {
  border-color: rgba(212, 175, 55, 0.3);
}

.stage-card.failed {
  border-color: rgba(157, 41, 51, 0.42);
  background: rgba(157, 41, 51, 0.08);
}

.stage-card.skipped {
  opacity: 0.72;
}

.stage-index {
  display: inline-flex;
  width: 38px;
  height: 38px;
  border-radius: 50%;
  align-items: center;
  justify-content: center;
  background: rgba(157, 41, 51, 0.1);
  color: var(--primary);
  font-weight: 600;
  font-family: 'Noto Serif SC', serif;
  flex-shrink: 0;
}

.stage-copy {
  min-width: 0;
}

.stage-title-row {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  align-items: baseline;
}

.stage-title-row strong {
  font-family: 'Noto Serif SC', serif;
  color: var(--secondary);
  font-size: 16px;
}

.stage-status {
  color: rgba(89, 63, 52, 0.74);
  font-size: 12px;
  white-space: nowrap;
}

.stage-copy p {
  margin: 8px 0 0;
  line-height: 1.7;
  color: rgba(52, 38, 33, 0.78);
  font-size: 14px;
}

@media (max-width: 900px) {
  .stage-list {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 640px) {
  .progress-board {
    padding: 18px;
  }

  .board-header {
    flex-direction: column;
  }

  .board-meta {
    justify-content: flex-start;
  }
}
</style>
