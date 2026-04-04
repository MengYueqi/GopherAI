<template>
  <div class="travel-planning-page">
    <section class="input-panel glass-panel">
      <div class="panel-header">
        <div>
          <span class="pill">Travel Planner</span>
          <h1>AI 旅游规划助手</h1>
          <p v-if="showInput">输入出发地、目的地、天数与偏好，我们会为你生成可执行的行程建议。</p>
          <p v-else>为您生成的专属旅游行程方案</p>
        </div>
        <button type="button" class="back-btn" @click="$router.push('/menu')">← 返回</button>
      </div>

      <template v-if="showInput">
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

        <div class="empty-state">
          <p>提交需求后，这里会展示 AI 返回的行程规划。</p>
        </div>
      </template>

      <template v-else>
        <div class="loading-state" v-if="loading">
          <div class="loading-spinner"></div>
          <h3>正在为您生成行程方案...</h3>
          <p>已用时 {{ elapsedSeconds }}s，请稍候</p>
        </div>

        <div class="result-card" v-else>
          <div class="result-header">
            <h3>行程规划结果</h3>
            <button type="button" class="secondary-btn" @click="resetToInput">
              重新生成
            </button>
          </div>

          <template v-if="hasStructuredPlan">
            <div class="overview-grid">
              <section class="overview-card accent">
                <p class="section-eyebrow">总体概括</p>
                <p class="overview-text">{{ planData.overall_summary || '暂无总体概括' }}</p>
                <p v-if="planData.notice" class="notice-text">{{ planData.notice }}</p>
              </section>

              <section class="overview-card">
                <p class="section-eyebrow">机票价格</p>
                <div class="flight-meta">
                  <span class="meta-chip">{{ planData.flight_price?.currency || '未提供币种' }}</span>
                  <span class="meta-chip">{{ planData.flight_price?.price_range || '暂无价格区间' }}</span>
                </div>
                <p class="flight-summary">{{ planData.flight_price?.summary || '暂无机票总结' }}</p>
                <p v-if="planData.flight_price?.raw_text" class="flight-raw">{{ planData.flight_price.raw_text }}</p>
                <ul v-if="planData.flight_price?.booking_tips?.length" class="tips-list">
                  <li v-for="tip in planData.flight_price.booking_tips" :key="tip">{{ tip }}</li>
                </ul>
              </section>
            </div>

            <section class="daily-section">
              <div class="section-title-row">
                <p class="section-eyebrow">每日计划</p>
                <span class="day-count">{{ planData.daily_plans.length }} 天</span>
              </div>

              <div class="day-stack">
                <article v-for="day in planData.daily_plans" :key="`${day.day}-${day.title}`" class="day-card">
                  <div class="day-heading">
                    <div>
                      <span class="day-badge">Day {{ day.day || '-' }}</span>
                      <h4>{{ day.title || '当日安排' }}</h4>
                    </div>
                    <p class="route-text">{{ day.route || '暂无路线信息' }}</p>
                  </div>

                  <div class="day-meta">
                    <span class="meta-chip">交通：{{ day.transport || '待定' }}</span>
                  </div>

                  <p class="day-summary">{{ day.summary || '暂无当日概述' }}</p>

                  <div v-if="day.attractions?.length" class="attraction-list">
                    <section
                      v-for="attraction in day.attractions"
                      :key="`${day.day}-${attraction.name}`"
                      class="attraction-card"
                    >
                      <div class="attraction-copy">
                        <h5>{{ attraction.name || '未命名景点' }}</h5>
                        <p>{{ attraction.description || '暂无景点介绍' }}</p>
                        <ul v-if="attraction.highlights?.length" class="highlight-list">
                          <li v-for="item in attraction.highlights" :key="item">{{ item }}</li>
                        </ul>
                      </div>

                      <div class="image-grid" v-if="attraction.images?.length">
                        <a
                          v-for="image in attraction.images"
                          :key="`${attraction.name}-${image.url}`"
                          class="image-card"
                          :href="image.url"
                          target="_blank"
                          rel="noreferrer"
                        >
                          <img :src="image.url" :alt="image.title || attraction.name" />
                          <div class="image-caption">
                            <strong>{{ image.title || attraction.name }}</strong>
                            <span>{{ image.source || '图片来源' }}</span>
                            <span v-if="image.source_url" class="source-link">{{ image.source_url }}</span>
                          </div>
                        </a>
                      </div>

                      <p v-else class="empty-image">该景点暂未提供图片。</p>
                    </section>
                  </div>

                  <ul v-if="day.tips?.length" class="tips-list">
                    <li v-for="tip in day.tips" :key="tip">{{ tip }}</li>
                  </ul>
                </article>
              </div>
            </section>

            <section v-if="planData.sources?.length" class="source-section">
              <p class="section-eyebrow">来源信息</p>
              <ul class="source-list">
                <li v-for="source in planData.sources" :key="source">{{ source }}</li>
              </ul>
            </section>
          </template>

          <template v-else>
            <MdPreview
              :modelValue="rawAdvice || '暂无建议'"
              previewTheme="github"
              :showCodeRowNumber="false"
            />
          </template>
        </div>
      </template>
    </section>
  </div>
</template>

<script>
import { computed, onUnmounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { MdPreview } from 'md-editor-v3'
import api from '../utils/api'

const createEmptyPlan = () => ({
  mode: '',
  overall_summary: '',
  flight_price: {
    summary: '',
    currency: '',
    price_range: '',
    booking_tips: [],
    raw_text: ''
  },
  daily_plans: [],
  sources: [],
  notice: '',
  raw_text: ''
})

const normalizePlan = (payload) => {
  const base = createEmptyPlan()
  if (!payload || typeof payload !== 'object' || Array.isArray(payload)) {
    return base
  }

  return {
    mode: payload.mode || '',
    overall_summary: payload.overall_summary || '',
    flight_price: {
      summary: payload.flight_price?.summary || '',
      currency: payload.flight_price?.currency || '',
      price_range: payload.flight_price?.price_range || '',
      booking_tips: Array.isArray(payload.flight_price?.booking_tips) ? payload.flight_price.booking_tips : [],
      raw_text: payload.flight_price?.raw_text || ''
    },
    daily_plans: Array.isArray(payload.daily_plans) ? payload.daily_plans.map((day, index) => ({
      day: Number.isFinite(day?.day) ? day.day : index + 1,
      title: day?.title || '',
      route: day?.route || '',
      transport: day?.transport || '',
      summary: day?.summary || '',
      attractions: Array.isArray(day?.attractions) ? day.attractions.map((attraction) => ({
        name: attraction?.name || '',
        description: attraction?.description || '',
        highlights: Array.isArray(attraction?.highlights) ? attraction.highlights : [],
        images: Array.isArray(attraction?.images) ? attraction.images.filter(image => image?.url).map((image) => ({
          title: image?.title || '',
          url: image?.url || '',
          source: image?.source || '',
          source_url: image?.source_url || ''
        })) : []
      })) : [],
      tips: Array.isArray(day?.tips) ? day.tips : []
    })) : [],
    sources: Array.isArray(payload.sources) ? payload.sources : [],
    notice: payload.notice || '',
    raw_text: payload.raw_text || ''
  }
}

export default {
  name: 'TravelPlanning',
  components: {
    MdPreview
  },
  setup() {
    const description = ref('')
    const loading = ref(false)
    const planData = ref(createEmptyPlan())
    const rawAdvice = ref('')
    const elapsedSeconds = ref(0)
    const showInput = ref(true)
    let timerId = null

    const hasStructuredPlan = computed(() => {
      return planData.value.mode === 'plan' && Array.isArray(planData.value.daily_plans) && planData.value.daily_plans.length > 0
    })

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
      showInput.value = false
      loading.value = true
      startTimer()
      try {
        const payload = { description: description.value.trim() }
        const response = await api.post('/AI/agent/medical_advice', payload)

        if (response.data && response.data.status_code === 1000) {
          const advice = response.data.advice
          if (advice && typeof advice === 'object') {
            planData.value = normalizePlan(advice)
            rawAdvice.value = advice.raw_text || ''
          } else {
            planData.value = createEmptyPlan()
            rawAdvice.value = typeof advice === 'string' ? advice : '暂无建议'
          }
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

    const resetToInput = () => {
      showInput.value = true
      planData.value = createEmptyPlan()
      rawAdvice.value = ''
      description.value = ''
    }

    onUnmounted(() => {
      stopTimer()
    })

    return {
      description,
      loading,
      planData,
      rawAdvice,
      elapsedSeconds,
      showInput,
      hasStructuredPlan,
      submitAdvice,
      resetToInput
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
  width: min(1120px, 100%);
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
  padding: 24px;
  background: rgba(255, 255, 255, 0.92);
}

.result-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  gap: 16px;
}

.overview-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 18px;
  margin-bottom: 24px;
}

.overview-card {
  border-radius: 20px;
  padding: 20px;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(244, 247, 255, 0.98));
  border: 1px solid rgba(99, 102, 241, 0.14);
}

.overview-card.accent {
  background: linear-gradient(135deg, rgba(98, 116, 255, 0.12), rgba(89, 201, 169, 0.12));
}

.section-eyebrow {
  margin: 0 0 10px;
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--text-muted);
}

.overview-text,
.flight-summary,
.flight-raw,
.day-summary,
.route-text,
.notice-text {
  margin: 0;
  line-height: 1.7;
}

.notice-text {
  margin-top: 12px;
  color: #b45309;
}

.flight-meta,
.day-meta {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
  margin-bottom: 12px;
}

.meta-chip,
.day-badge,
.day-count {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  padding: 6px 12px;
  font-size: 12px;
  font-weight: 700;
  background: rgba(99, 102, 241, 0.1);
  color: var(--primary-dark);
}

.daily-section,
.source-section {
  margin-top: 24px;
}

.section-title-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  gap: 12px;
}

.day-stack {
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.day-card {
  border: 1px solid rgba(15, 23, 42, 0.08);
  border-radius: 24px;
  padding: 22px;
  background: #fff;
  box-shadow: 0 12px 32px rgba(15, 23, 42, 0.05);
}

.day-heading {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 16px;
  margin-bottom: 12px;
}

.day-heading h4,
.attraction-copy h5 {
  margin: 10px 0 0;
}

.attraction-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
  margin-top: 18px;
}

.attraction-card {
  border-radius: 20px;
  border: 1px solid rgba(99, 102, 241, 0.12);
  padding: 18px;
  background: rgba(248, 250, 255, 0.9);
}

.attraction-copy p {
  line-height: 1.7;
}

.highlight-list,
.tips-list,
.source-list {
  margin: 12px 0 0;
  padding-left: 18px;
  line-height: 1.7;
}

.image-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 14px;
  margin-top: 16px;
}

.image-card {
  display: block;
  overflow: hidden;
  border-radius: 18px;
  text-decoration: none;
  background: #fff;
  border: 1px solid rgba(15, 23, 42, 0.08);
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.image-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 14px 28px rgba(15, 23, 42, 0.12);
}

.image-card img {
  display: block;
  width: 100%;
  height: 180px;
  object-fit: cover;
  background: rgba(148, 163, 184, 0.16);
}

.image-caption {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 12px;
  color: var(--text-color);
  font-size: 13px;
}

.source-link {
  color: var(--text-muted);
  word-break: break-all;
}

.empty-image {
  margin: 16px 0 0;
  color: var(--text-muted);
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
}

.loading-state {
  display: grid;
  place-items: center;
  min-height: 320px;
  text-align: center;
  gap: 14px;
}

.loading-spinner {
  width: 44px;
  height: 44px;
  border-radius: 50%;
  border: 4px solid rgba(99, 102, 241, 0.14);
  border-top-color: var(--primary);
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

@media (max-width: 900px) {
  .travel-planning-page {
    padding: 28px 16px 40px;
  }

  .input-panel {
    padding: 22px;
  }

  .panel-header,
  .result-header,
  .day-heading,
  .section-title-row {
    flex-direction: column;
    align-items: flex-start;
  }

  .overview-grid {
    grid-template-columns: 1fr;
  }
}
</style>
