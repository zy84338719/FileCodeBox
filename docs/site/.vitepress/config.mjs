import { defineConfig } from 'vitepress'

export default defineConfig({
  lang: 'zh-CN',
  title: 'FileCodeBox',
  description: '文件快传服务的部署、使用与 API 指南',
  themeConfig: {
    nav: [
      { text: '指南', link: '/guide/introduction' },
      { text: '快速上手', link: '/guide/getting-started' },
      { text: 'API 参考', link: '/api/' }
    ],
    sidebar: {
      '/guide/': [
        {
          text: '指南',
          items: [
            { text: '概览', link: '/guide/introduction' },
            { text: '快速上手', link: '/guide/getting-started' },
            { text: '上传与分片', link: '/guide/upload' },
            { text: '分享能力', link: '/guide/share' },
            { text: '后台管理', link: '/guide/management' },
            { text: '存储后端', link: '/guide/storage' },
            { text: '系统配置', link: '/guide/configuration' },
            { text: '安全加固', link: '/guide/security' },
            { text: '故障排查', link: '/guide/troubleshooting' }
          ]
        }
      ],
      '/api/': [
        {
          text: 'API 参考',
          items: [
            { text: '总览', link: '/api/' },
            { text: '上传与分片', link: '/api/upload' },
            { text: '后台管理', link: '/api/admin' }
          ]
        }
      ]
    },
    socialLinks: [
      { icon: 'github', link: 'https://github.com/zy84338719/FileCodeBox' }
    ]
  }
})
