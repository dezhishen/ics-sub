# ics-sub

基于 Go 插件生成日历数据与 ICS 文件，并使用 Vue 展示分组订阅链接，最终通过 GitHub Pages 发布。

## 架构

- `plugins/*`：数据插件，输出统一的日历数据模型（含 `group`、`name`、`events`）
- `cmd/generate`：读取插件数据，产出：
  - `web/public/data/subscriptions.json`（前端分组索引）
  - `web/public/ics/<calendar-id>.ics`（每个日历一个文件）
- `web`：Vue + Vite 前端，按分组页签展示并支持过滤
- `.github/workflows/deploy-pages.yml`：CI 先跑 Go 生成，再构建前端并推送到 `gh-pages`

## 本地使用

1. 生成数据

```bash
go run ./cmd/generate
```

2. 启动前端

```bash
cd web
npm install
npm run dev
```

3. 构建产物

```bash
npm run build
```

## 数据约定

- 一份源数据同时驱动 JSON 与 ICS，避免重复维护
- 每个日历按 `id` 输出一个 `.ics` 文件（示例：`cn-holiday.ics`）
- JSON 以 `groups` 组织，前端天然支持分组页签和过滤
