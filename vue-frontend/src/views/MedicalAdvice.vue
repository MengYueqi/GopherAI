<template>
  <div class="medical-advice-page">
    <section class="input-panel glass-panel">
      <div class="panel-header">
        <div>
          <span class="pill">Medical Agent</span>
          <h1>AI 医疗建议助手</h1>
          <p>描述症状或检查结果，我们会为你生成结构化的建议，仅供学习参考。</p>
        </div>
        <button type="button" class="back-btn" @click="$router.push('/menu')">← 返回</button>
      </div>

      <label for="symptom-input" class="field-label">症状或场景描述</label>
      <textarea
        id="symptom-input"
        v-model="description"
        placeholder="例如：最近持续咳嗽伴随低烧，需要了解可能的原因与建议"
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
          {{ loading ? '生成中…' : '生成建议' }}
        </button>
        <button type="button" class="secondary-btn" :disabled="!history.length" @click="clearHistory">
          清空历史
        </button>
      </div>
      <p class="disclaimer">⚠️ 本页面仅供学习与研究，无法替代专业医生诊断。</p>

      <div class="result-card" v-if="currentAdvice">
        <h3>最新建议</h3>
        <MdPreview
          :modelValue="currentAdvice"
          previewTheme="github"
          :showCodeRowNumber="false"
        />
      </div>
      <div class="empty-state" v-else>
        <p>提交问题后，这里会展示 AI 返回的建议摘要。</p>
      </div>
    </section>

    <section class="history-panel glass-panel">
      <div class="panel-header">
        <div>
          <h2>历史记录</h2>
          <p>最近 {{ history.length }} 条查询</p>
        </div>
      </div>
      <ul class="history-list" v-if="history.length">
        <li v-for="(item, index) in history" :key="`${item.time}-${index}`">
          <div class="history-meta">
            <strong>{{ item.time }}</strong>
            <span>{{ item.description }}</span>
          </div>
          <MdPreview
            :modelValue="item.advice"
            previewTheme="github"
            :showCodeRowNumber="false"
          />
        </li>
      </ul>
      <div class="empty-state" v-else>
        <p>历史记录为空，生成成功的建议会自动保存在这里。</p>
      </div>
    </section>
  </div>
</template>

<script>
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { MdPreview } from 'md-editor-v3'
import api from '../utils/api'

export default {
  name: 'MedicalAdvice',
  components: {
    MdPreview
  },
  setup() {
    const description = ref('')
    const loading = ref(false)
    const currentAdvice = ref('')
    const history = ref([])

    const submitAdvice = async () => {
      if (!description.value.trim() || loading.value) return
      loading.value = true
      try {
        const payload = { description: description.value.trim() }
        const response = await api.post('/AI/agent/medical_advice', payload)

        if (response.data && response.data.status_code === 1000) {
          const advice = response.data.advice || '暂无建议'
          currentAdvice.value = advice
          history.value.unshift({
            description: payload.description,
            advice,
            time: new Date().toLocaleString()
          })
          description.value = ''
        } else {
          const message = response.data?.status_msg || '生成失败，请稍后再试'
          ElMessage.error(message)
        }
      } catch (error) {
        ElMessage.error(error?.response?.data?.status_msg || '无法连接到服务器')
      } finally {
        loading.value = false
      }
    }

    const clearHistory = () => {
      history.value = []
    }

    return {
      description,
      loading,
      currentAdvice,
      history,
      submitAdvice,
      clearHistory
    }
  }
}
</script>

<style scoped>
.medical-advice-page {
  min-height: 100vh;
  display: grid;
  grid-template-columns: minmax(320px, 2fr) minmax(260px, 1fr);
  gap: 28px;
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

.input-panel,
.history-panel {
  padding: 36px;
  display: flex;
  flex-direction: column;
  gap: 18px;
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
  gap: 12px;
  flex-wrap: wrap;
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

.disclaimer {
  font-size: 13px;
  color: #b45309;
  background: rgba(251, 191, 36, 0.2);
  padding: 10px 12px;
  border-radius: 12px;
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

.history-list {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.history-list li {
  border: 1px solid rgba(15, 23, 42, 0.08);
  border-radius: var(--radius-md);
  padding: 14px 16px;
  background: rgba(255, 255, 255, 0.9);
}

.history-meta {
  display: flex;
  flex-direction: column;
  gap: 4px;
  margin-bottom: 8px;
}

.history-meta strong {
  font-size: 14px;
  color: var(--primary-dark);
}

.history-meta span {
  font-size: 15px;
  color: var(--text-color);
}

@media (max-width: 900px) {
  .medical-advice-page {
    grid-template-columns: 1fr;
  }
}
</style>
