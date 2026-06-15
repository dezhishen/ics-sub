# ICS 订阅聚合

这是一个可直接使用的日历订阅聚合站点：

- 网页中可按分组查看可订阅日历
- 每个日历提供标准 `.ics` 链接，可用于 Apple Calendar、Google Calendar、Outlook 等客户端
- 数据由 GitHub Actions 自动生成并发布到 GitHub Pages

## 使用说明

1. 打开你的 GitHub Pages 站点（示例：`https://<你的用户名>.github.io/<仓库名>/`）。
2. 在页面中找到要订阅的日历。
3. 复制对应的 ICS 链接（通常形如 `.../ics/<calendar-id>.ics`）。
4. 在你的日历客户端中添加“通过 URL 订阅”。

## 常见客户端添加方式

- Apple Calendar（macOS/iOS）：
  - 选择“新建订阅日历”并粘贴 ICS URL。
- Google Calendar：
  - 左侧“其他日历” -> “通过 URL 添加”。
- Outlook：
  - 选择“从 Internet 订阅日历”并粘贴 ICS URL。

## 自动更新

- 工作流文件：`.github/workflows/deploy-pages.yml`
- 触发方式：
  - 推送到 `main` 分支时自动更新
  - 手动触发（`workflow_dispatch`）
  - 定时触发：每小时整点（UTC）自动生成并发布一次

## 自建/二次部署

如果你想维护自己的订阅内容：

1. Fork 本仓库。
2. 按你的需求调整 `plugins/` 下的数据来源。
3. 确保仓库已启用 GitHub Pages（发布分支为 `gh-pages`）。
4. 推送到 `main` 后等待 Action 自动发布。

## 目录说明

- `web/public/data/subscriptions.json`：前端展示用的数据索引
- `web/public/ics/`：实际订阅的 ICS 文件目录

