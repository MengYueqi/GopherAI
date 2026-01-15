<template>
  <div class="travel-planning-page">
    <section class="input-panel glass-panel">
      <div class="panel-header">
        <div>
          <span class="pill">Travel Planner</span>
          <h1>AI 旅游规划助手</h1>
          <p>输入出发地、目的地、天数与偏好，我们会为你生成可执行的行程建议。</p>
        </div>
        <button type="button" class="back-btn" @click="$router.push('/menu')">← 返回</button>
      </div>

      <label for="travel-input" class="field-label">行程需求描述</label>
      <textarea
        id="travel-input"
        v-model="description"
        placeholder="例如：从上海出发，7 天游云南，喜欢自然风景，预算 6k 左右"
        :disabled="loading"
        rows="6"
      ></textarea>

      <div class="actions">
        <button
          type="button"
          class="primary-btn"
          :disabled="loading || !description.trim()"
          @click="submitAdvice"
        >
          {{ loading ? '生成中…' : '生成行程' }}
        </button>
        <span v-if="loading" class="timer">已用时 {{ elapsedSeconds }}s</span>
      </div>

      <div class="result-card" v-if="currentAdvice">
        <h3>最新行程</h3>
        <MdPreview
          :modelValue="currentAdvice"
          previewTheme="github"
          :showCodeRowNumber="false"
        />
      </div>
      <div class="empty-state" v-else>
        <p>提交需求后，这里会展示 AI 返回的行程规划。</p>
      </div>
    </section>
  </div>
</template>

<script>
import { ref, onUnmounted } from 'vue'
import { ElMessage } from 'element-plus'
import { MdPreview } from 'md-editor-v3'
import api from '../utils/api'

export default {
  name: 'TravelPlanning',
  components: {
    MdPreview
  },
  setup() {
    const description = ref('')
    const loading = ref(false)
    const currentAdvice = ref('')
    const elapsedSeconds = ref(0)
    let timerId = null

    const startTimer = () => {
      elapsedSeconds.value = 0
      if (timerId) clearInterval(timerId)
      timerId = setInterval(() => {
        elapsedSeconds.value += 1
      }, 1000)
    }

    const stopTimer = () => {
      if (timerId) {
        clearInterval(timerId)
        timerId = null
      }
    }

    const submitAdvice = async () => {
      if (!description.value.trim() || loading.value) return
      loading.value = true
      startTimer()
      try {
        const payload = { description: description.value.trim() }
        const response = await api.post('/AI/agent/medical_advice', payload)

        if (response.data && response.data.status_code === 1000) {
          const advice = response.data.advice || '暂无建议'
          currentAdvice.value = advice
          description.value = ''
        } else {
          const message = response.data?.status_msg || '生成失败，请稍后再试'
          ElMessage.error(message)
        }
      } catch (error) {
        ElMessage.error(error?.response?.data?.status_msg || '无法连接到服务器')
      } finally {
        loading.value = false
        stopTimer()
      }
    }

    onUnmounted(() => {
      stopTimer()
    })

    return {
      description,
      loading,
      currentAdvice,
      elapsedSeconds,
      submitAdvice
    }
  }
}
</script>

<style scoped>
.travel-planning-page {
  min-height: 100vh;
  display: flex;
  justify-content: center;
  padding: 48px 6vw 64px;
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 16px;
}

.panel-header h1 {
  margin: 6px 0 8px;
}

.input-panel {
  padding: 36px;
  display: flex;
  flex-direction: column;
  gap: 18px;
  width: min(920px, 100%);
}

.back-btn {
  align-self: flex-start;
  border: none;
  background: rgba(112, 100, 255, 0.12);
  color: var(--primary-dark);
  padding: 10px 16px;
  border-radius: 12px;
  cursor: pointer;
}

.back-btn:hover {
  background: rgba(112, 100, 255, 0.2);
}

.field-label {
  font-weight: 600;
}

textarea {
  width: 100%;
  border-radius: var(--radius-md);
  border: 1px solid rgba(99, 102, 241, 0.24);
  padding: 14px;
  font-size: 15px;
  resize: vertical;
  min-height: 140px;
  background: rgba(255, 255, 255, 0.9);
  line-height: 1.6;
}

textarea:focus {
  outline: 2px solid rgba(112, 100, 255, 0.35);
}

.actions {
  display: flex;
  gap: 14px;
  flex-wrap: wrap;
  align-items: center;
}

.primary-btn,
.secondary-btn {
  border: none;
  border-radius: 14px;
  padding: 12px 20px;
  font-weight: 600;
  cursor: pointer;
}

.primary-btn {
  background: linear-gradient(120deg, var(--primary), var(--secondary));
  color: #fff;
}

.primary-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.secondary-btn {
  background: rgba(15, 23, 42, 0.08);
  color: var(--text-color);
}

.result-card {
  border: 1px solid rgba(99, 102, 241, 0.2);
  border-radius: var(--radius-lg);
  padding: 20px;
  background: rgba(255, 255, 255, 0.92);
}

.empty-state {
  text-align: center;
  padding: 32px;
  border: 1px dashed rgba(15, 23, 42, 0.2);
  border-radius: var(--radius-lg);
  color: var(--text-muted);
}

.timer {
  font-size: 13px;
  color: var(--text-muted);
  background: rgba(15, 23, 42, 0.06);
  padding: 8px 12px;
  border-radius: 999px;
}
</style>
