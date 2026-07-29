<template>
  <div class="trend-chart">
    <svg
      :viewBox="`0 0 ${width} ${height}`"
      preserveAspectRatio="none"
      class="chart-svg"
    >
      <!-- 网格线 -->
      <g class="grid">
        <line
          v-for="(y, i) in gridLines"
          :key="`h-${i}`"
          :x1="padding.left"
          :y1="y"
          :x2="width - padding.right"
          :y2="y"
          stroke="var(--color-border)"
          stroke-dasharray="3,3"
        />
      </g>

      <!-- Y 轴标签 -->
      <g class="y-axis-labels">
        <text
          v-for="(y, i) in gridLines"
          :key="`yl-${i}`"
          :x="padding.left - 6"
          :y="y + 4"
          text-anchor="end"
          font-size="10"
          fill="var(--color-text-secondary)"
        >
          {{ Math.round(maxValue * (1 - i / 4)) }}
        </text>
      </g>

      <!-- X 轴标签 -->
      <g class="x-axis-labels">
        <text
          v-for="(point, i) in points"
          :key="`xl-${i}`"
          :x="point.x"
          :y="height - 4"
          text-anchor="middle"
          font-size="10"
          fill="var(--color-text-secondary)"
        >
          {{ point.label }}
        </text>
      </g>

      <!-- 上传面积 -->
      <path
        v-if="uploadPath"
        :d="uploadAreaPath"
        fill="url(#uploadGradient)"
        opacity="0.3"
      />
      <path
        v-if="uploadPath"
        :d="uploadPath"
        fill="none"
        stroke="#5e6ad2"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
      />

      <!-- 下载折线 -->
      <path
        v-if="downloadPath"
        :d="downloadPath"
        fill="none"
        stroke="#f56c6c"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
        stroke-dasharray="4,2"
      />

      <!-- 数据点 -->
      <g class="points">
        <circle
          v-for="(p, i) in points"
          :key="`u-${i}`"
          :cx="p.x"
          :cy="p.uploadY"
          r="3"
          fill="#5e6ad2"
        />
      </g>

      <!-- 定义渐变 -->
      <defs>
        <linearGradient id="uploadGradient" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stop-color="#5e6ad2" stop-opacity="0.6" />
          <stop offset="100%" stop-color="#5e6ad2" stop-opacity="0" />
        </linearGradient>
      </defs>
    </svg>

    <div class="chart-legend">
      <div class="legend-item">
        <span class="legend-dot upload"></span>
        <span>{{ uploadLabel }}</span>
      </div>
      <div class="legend-item" v-if="data && data.length > 0 && data[0]?.downloads !== undefined">
        <span class="legend-dot download"></span>
        <span>{{ downloadLabel }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

export interface TrendPoint {
  date: string // YYYY-MM-DD or label
  uploads: number
  downloads?: number
}

interface Props {
  data: TrendPoint[]
  width?: number
  height?: number
  uploadLabel?: string
  downloadLabel?: string
}

const props = withDefaults(defineProps<Props>(), {
  width: 600,
  height: 240,
  uploadLabel: 'Uploads',
  downloadLabel: 'Downloads',
})

const padding = { top: 16, right: 16, bottom: 28, left: 36 }

const maxValue = computed(() => {
  if (!props.data || props.data.length === 0) return 100
  let max = 0
  for (const p of props.data) {
    if (p.uploads > max) max = p.uploads
    if (p.downloads !== undefined && p.downloads > max) max = p.downloads
  }
  return Math.max(max, 10) * 1.1
})

const points = computed(() => {
  if (!props.data || props.data.length === 0) return []
  const chartW = props.width - padding.left - padding.right
  const chartH = props.height - padding.top - padding.bottom
  const n = props.data.length
  return props.data.map((d, i) => {
    const x = padding.left + (i / Math.max(n - 1, 1)) * chartW
    const uploadY = padding.top + chartH - (d.uploads / maxValue.value) * chartH
    const downloadY =
      d.downloads !== undefined
        ? padding.top + chartH - (d.downloads / maxValue.value) * chartH
        : uploadY
    // 截取 label 后几位（MM-DD）
    const label = d.date.length >= 10 ? d.date.slice(5) : d.date
    return { x, uploadY, downloadY, label }
  })
})

const uploadPath = computed(() => {
  if (points.value.length === 0) return ''
  return points.value
    .map((p, i) => `${i === 0 ? 'M' : 'L'} ${p.x} ${p.uploadY}`)
    .join(' ')
})

const uploadAreaPath = computed(() => {
  if (points.value.length === 0) return ''
  const baseY = props.height - padding.bottom
  const first = points.value[0]!
  const last = points.value[points.value.length - 1]!
  const line = points.value
    .map((p, i) => `${i === 0 ? 'M' : 'L'} ${p.x} ${p.uploadY}`)
    .join(' ')
  return `${line} L ${last.x} ${baseY} L ${first.x} ${baseY} Z`
})

const downloadPath = computed(() => {
  if (points.value.length === 0) return ''
  const hasDownload = props.data.some((d) => d.downloads !== undefined)
  if (!hasDownload) return ''
  return points.value
    .map((p, i) => `${i === 0 ? 'M' : 'L'} ${p.x} ${p.downloadY}`)
    .join(' ')
})

const gridLines = computed(() => {
  const chartH = props.height - padding.top - padding.bottom
  return [0, 0.25, 0.5, 0.75, 1].map(
    (p) => padding.top + chartH * p
  )
})
</script>

<style scoped>
.trend-chart {
  width: 100%;
}

.chart-svg {
  width: 100%;
  height: 240px;
  display: block;
}

.chart-legend {
  display: flex;
  justify-content: center;
  gap: 24px;
  margin-top: 8px;
  font-size: 13px;
  color: var(--color-text-secondary);
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 6px;
}

.legend-dot {
  width: 12px;
  height: 12px;
  border-radius: 50%;
}

.legend-dot.upload {
  background: #5e6ad2;
}

.legend-dot.download {
  background: #f56c6c;
}
</style>
