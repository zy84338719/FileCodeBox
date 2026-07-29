import DOMPurify from 'dompurify'

/**
 * 消毒 HTML，移除 script/事件处理器等危险内容。
 * 用于 v-html 渲染用户内容（markdown/代码高亮输出）前，防止存储型 XSS。
 */
export function sanitizeHtml(html: string): string {
  return DOMPurify.sanitize(html, {
    FORBID_TAGS: ['script', 'style', 'iframe', 'object', 'embed', 'form'],
    FORBID_ATTR: ['onerror', 'onload', 'onclick', 'onmouseover', 'onfocus', 'onblur'],
  })
}
