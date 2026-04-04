<template>
  <div class="travel-planning-page">
    <section class="input-panel chinese-panel">
      <div class="panel-header">
        <div>
          <span class="chinese-tag">知行阁</span>
          <h1>知行 · 游天下</h1>
          <p v-if="showInput">输入出发地、目的地、天数与偏好，为您生成完备的出行方案。</p>
          <p v-else>为您定制的专属旅行方案</p>
        </div>
        <button type="button" class="back-btn" @click="$router.push('/menu')">← 返</button>
      </div>

      <template v-if="showInput">
        <label for="travel-input" class="field-label">行旅需求</label>
        <textarea
          id="travel-input"
          v-model="description"
          placeholder="例如：自沪上出发，七日游滇，喜自然风光，预算六千左右"
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
            {{ loading ? '运筹中…' : '策划行程' }}
          </button>
          <span v-if="loading" class="timer">已用时 {{ elapsedSeconds }}秒</span>
        </div>

        <div class="empty-state">
          <p>提交需求后，此处将展示生成的行程方案。</p>
        </div>
      </template>

      <template v-else>
        <div class="loading-state" v-if="loading">
          <div class="loading-spinner"></div>
          <h3>正在为您运筹行程方案...</h3>
          <p>已用时 {{ elapsedSeconds }}秒，请稍候</p>
        </div>

        <div class="result-card" v-else>
          <div class="result-header">
            <h3>行程方案</h3>
            <button type="button" class="secondary-btn" @click="resetToInput">
              重绘
            </button>
          </div>

          <template v-if="hasStructuredPlan">
            <div class="overview-grid">
              <section class="overview-card accent">
                <p class="section-eyebrow">总览</p>
                <p class="overview-text">{{ planData.overall_summary || '暂无总体概括' }}</p>
                <p v-if="planData.notice" class="notice-text">{{ planData.notice }}</p>
              </section>

              <section class="overview-card">
                <p class="section-eyebrow">票务信息</p>
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
                <p class="section-eyebrow">每日行程</p>
                <span class="day-count">{{ planData.daily_plans.length }} 日</span>
              </div>

              <div class="day-stack">
                <article v-for="day in planData.daily_plans" :key="`${day.day}-${day.title}`" class="day-card">
                  <div class="day-heading">
                    <div>
                      <span class="day-badge">第 {{ day.day || '-' }} 日</span>
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
              <p class="section-eyebrow">参考资料</p>
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
  background: var(--bg-color);
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 16px;
  border-bottom: 1px solid rgba(28, 28, 28, 0.08);
  padding-bottom: 20px;
}

.panel-header h1 {
  margin: 8px 0 12px;
  font-family: 'Noto Serif SC', serif;
  color: var(--secondary);
  font-weight: 600;
  letter-spacing: 0.05em;
  font-size: 32px;
}

.panel-header p {
  line-height: 1.7;
}

.input-panel {
  padding: 40px;
  display: flex;
  flex-direction: column;
  gap: 24px;
  width: min(1120px, 100%);
}

.back-btn {
  align-self: flex-start;
  border: 1px solid var(--card-border);
  background: transparent;
  color: var(--text-color);
  padding: 8px 16px;
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: all 0.3s ease;
  font-family: 'Noto Serif SC', serif;
}

.back-btn:hover {
  border-color: var(--primary);
  color: var(--primary);
  background: rgba(157, 41, 51, 0.05);
}

.field-label {
  font-weight: 500;
  font-family: 'Noto Serif SC', serif;
  color: var(--secondary);
  font-size: 16px;
}

textarea {
  width: 100%;
  border-radius: var(--radius-sm);
  border: 1px solid rgba(28, 28, 28, 0.15);
  padding: 16px;
  font-size: 15px;
  resize: vertical;
  min-height: 140px;
  background: rgba(255, 255, 255, 0.8);
  line-height: 1.8;
  transition: all 0.3s ease;
  font-family: inherit;
}

textarea:focus {
  outline: none;
  border-color: var(--primary);
  box-shadow: 0 0 0 3px rgba(157, 41, 51, 0.08);
}

.actions {
  display: flex;
  gap: 14px;
  flex-wrap: wrap;
  align-items: center;
}

.primary-btn,
.secondary-btn {
  border-radius: var(--radius-sm);
  padding: 10px 20px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
  font-family: 'Noto Serif SC', serif;
  letter-spacing: 0.05em;
}

.primary-btn {
  border: 1px solid var(--primary);
  background: var(--primary);
  color: #fff;
  box-shadow: var(--shadow-sm);
  position: relative;
  overflow: hidden;
}

.primary-btn::before {
  content: '';
  position: absolute;
  top: 0;
  left: -100%;
  width: 100%;
  height: 100%;
  background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.2), transparent);
  transition: left 0.5s ease;
}

.primary-btn:hover::before {
  left: 100%;
}

.primary-btn:hover:not(:disabled) {
  background: var(--primary-dark);
  border-color: var(--primary-dark);
  transform: translateY(-2px);
  box-shadow: var(--shadow-md);
}

.primary-btn:disabled {
  background: rgba(157, 41, 51, 0.3);
  border-color: rgba(157, 41, 51, 0.3);
  box-shadow: none;
  cursor: not-allowed;
}

.secondary-btn {
  border: 1px solid var(--card-border);
  background: transparent;
  color: var(--text-color);
}

.secondary-btn:hover {
  border-color: var(--primary);
  color: var(--primary);
  background: rgba(157, 41, 51, 0.05);
  transform: translateY(-1px);
}

.result-card {
  border: 1px solid var(--card-border);
  border-radius: var(--radius-lg);
  padding: 28px;
  background: rgba(255, 255, 255, 0.95);
}

.result-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  gap: 16px;
  padding-bottom: 16px;
  border-bottom: 1px solid rgba(28, 28, 28, 0.06);
}

.result-header h3 {
  font-family: 'Noto Serif SC', serif;
  color: var(--secondary);
  font-weight: 600;
  font-size: 24px;
}

.overview-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 20px;
  margin-bottom: 32px;
}

.overview-card {
  border-radius: var(--radius-md);
  padding: 24px;
  background: var(--card-bg);
  border: 1px solid var(--card-border);
  transition: all 0.3s ease;
}

.overview-card:hover {
  transform: translateY(-4px);
  box-shadow: var(--shadow-md);
}

.overview-card.accent {
  background: linear-gradient(135deg, rgba(157, 41, 51, 0.08), rgba(212, 175, 55, 0.06));
  border-left: 3px solid var(--primary);
}

.section-eyebrow {
  margin: 0 0 12px;
  font-size: 12px;
  font-weight: 600;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--primary);
  font-family: 'Noto Serif SC', serif;
  border-left: 2px solid var(--primary);
  padding-left: 8px;
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
  border-radius: var(--radius-sm);
  padding: 6px 12px;
  font-size: 12px;
  font-weight: 500;
  background: rgba(157, 41, 51, 0.1);
  color: var(--primary-dark);
  font-family: 'Noto Serif SC', serif;
  border-left: 2px solid var(--primary);
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
  border: 1px solid var(--card-border);
  border-radius: var(--radius-lg);
  padding: 24px;
  background: var(--card-bg);
  box-shadow: var(--shadow-sm);
  transition: all 0.3s ease;
  position: relative;
  overflow: hidden;
}

.day-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  width: 4px;
  height: 0;
  background: var(--secondary);
  transition: height 0.3s ease;
}

.day-card:hover {
  transform: translateY(-4px);
  box-shadow: var(--shadow-lg);
}

.day-card:hover::before {
  height: 100%;
}

.day-heading {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 16px;
  margin-bottom: 16px;
}

.day-heading h4,
.attraction-copy h5 {
  margin: 10px 0 0;
  font-family: 'Noto Serif SC', serif;
  color: var(--secondary);
  font-weight: 600;
}

.attraction-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
  margin-top: 20px;
}

.attraction-card {
  border-radius: var(--radius-md);
  border: 1px solid rgba(28, 28, 28, 0.08);
  padding: 20px;
  background: rgba(250, 246, 237, 0.3);
  transition: all 0.3s ease;
}

.attraction-card:hover {
  transform: translateX(4px);
  border-color: var(--jade);
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
  border-radius: var(--radius-md);
  text-decoration: none;
  background: #fff;
  border: 1px solid rgba(28, 28, 28, 0.08);
  transition: all 0.3s ease;
}

.image-card:hover {
  transform: translateY(-4px);
  box-shadow: var(--shadow-md);
}

.image-card img {
  display: block;
  width: 100%;
  height: 180px;
  object-fit: cover;
  background: rgba(28, 28, 28, 0.05);
  transition: transform 0.3s ease;
}

.image-card:hover img {
  transform: scale(1.05);
}

.image-caption {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 12px;
  color: var(--text-color);
  font-size: 13px;
  border-top: 1px solid rgba(28, 28, 28, 0.06);
}

.image-caption strong {
  font-family: 'Noto Serif SC', serif;
  color: var(--secondary);
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
