---
layout: home
title: FileCodeBox 文档中心
lang: zh-CN
hero:
  name: FileCodeBox
  text: Go Edition 文档中心
  tagline: 像拿快递一样安全、高效地分享文件与文本
  actions:
    - text: 快速开始
      link: /guide/getting-started
    - text: 功能概览
      link: /guide/introduction
    - text: GitHub
      link: https://github.com/zy84338719/FileCodeBox
      theme: alt
features:
  - icon: 🚀
    title: 极速部署
    details: Docker、二进制、Compose 一键跑通，内置 Setup 向导
  - icon: 🔐
    title: 安全可靠
    details: 提取码、有效期、限次、JWT 后台与可选用户体系
  - icon: 💻
    title: 现代体验
    details: 2025 新主题，Dashboard/Admin 全面无障碍与响应式
  - icon: 🔌
    title: 可插拔存储
    details: 本地、S3、WebDAV、OneDrive 多后端无缝切换
---

## 当前版本

> **v1.9.9** — 焕然一新的 Dashboard/Admin UI、统一设计系统、升级的忘记密码流程，以及全新文档体系。

- 完整版本记录请查看仓库 `docs/` 中的发布笔记与重构报告
- 通过 `git tag v1.9.9` 获取最新稳定版本

## 文档地图

### 入门
- [什么是 FileCodeBox](/guide/introduction)
- [快速开始](/guide/getting-started)
- [文件上传体验](/guide/upload)
- [分享与领取](/guide/share)

### 运维与配置
- [管理后台总览](/guide/management)
- [存储适配指南](/guide/storage)
- [系统配置详解](/guide/configuration)
- [安全加固最佳实践](/guide/security)
- [常见问题排查](/guide/troubleshooting)

### 开发者专区
- [API 使用指南](/api/)
- [Upload/Chunk API 示例](/api/upload)
- [管理类 API 示例](/api/admin)

## 适用读者

- 需要搭建团队级“文件快递柜”的运维 / DevOps
- 希望基于 FileCodeBox 二次开发、接入自定义权限策略的研发
- 想快速掌握 2025 版前后端架构与配置体系的产品 / 项目负责人

## 撰写原则

- **贴近实践**：所有命令、配置片段均在真实环境验证
- **模块化组织**：入门、配置、API 分区，帮助迅速定位问题
- **版本提示**：涉及特定版本（如 v1.9.9 UI）时都会明确标注
- **持续更新**：欢迎在 GitHub Issues 提交改进建议或补充案例

> 遇到未覆盖的场景，请首先查阅仓库 `docs/` 目录下的专题文章，或通过 Issue / Discussions 反馈需求。
