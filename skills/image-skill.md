---
name: claw-image
description: "图片处理服务 CLI，支持格式转换、压缩、缩放、OCR、AI识别与生成"
metadata:
  {
    "openclaw": {
      "emoji": "🖼️",
      "requires": { "bins": ["claw-pliers"] }
    }
  }
---

# Image Service Skill

图片处理服务，提供格式转换、压缩、缩放、OCR、AI识别与生成功能。

## 安装

```bash
go build -o claw-pliers ./cli/
```

## 配置

配置 API 密钥 (可选):
```bash
claw-pliers-cli image config --ocr-key <key> --vision-key <key>
```

## 命令

### 格式转换
```bash
claw-pliers-cli image convert input.jpg output.webp --quality 80
```

### 图片压缩
```bash
claw-pliers-cli image compress input.jpg --quality 75
claw-pliers-cli image compress input.jpg --max-size 200KB
```

### 缩放
```bash
claw-pliers-cli image resize input.jpg output.jpg --width 800
```

### 旋转
```bash
claw-pliers-cli image rotate input.jpg output.jpg --degrees 90
```

### OCR 文字识别
```bash
claw-pliers-cli image ocr document.jpg
claw-pliers-cli image ocr document.jpg --output result.txt
```

### AI 图片识别
```bash
claw-pliers-cli image recognize photo.jpg
claw-pliers-cli image recognize chart.png --prompt "分析数据趋势"
```

### AI 图片生成
```bash
claw-pliers-cli image generate "一只可爱的小猫咪"
```

## API 端点

| 方法 | 路径 | 描述 |
|------|------|------|
| POST | /api/v1/image/convert | 格式转换 |
| POST | /api/v1/image/compress | 图片压缩 |
| POST | /api/v1/image/resize | 图片缩放 |
| POST | /api/v1/image/rotate | 旋转翻转 |
| POST | /api/v1/image/watermark | 添加水印 |
| POST | /api/v1/image/ocr | OCR 识别 |
| POST | /api/v1/image/recognize | AI 识别 |
| POST | /api/v1/image/generate | AI 生成 |

## 认证

所有 API 请求需要通过 `X-Local-Key` 头进行认证。
