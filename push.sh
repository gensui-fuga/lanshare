#!/bin/bash
# 一鍵推送 LanShare 到 GitHub
# 用法: ./push.sh <your_github_token>

set -e

if [ -z "$1" ]; then
    echo "用法: $0 <github_token>"
    echo ""
    echo "1. 去 https://github.com/settings/tokens 創建一個新 token"
    echo "   勾選 public_repo 權限"
    echo "2. 然後執行:"
    echo "   $0 ghp_xxxxxxxxxxxxxxxxxxxx"
    exit 1
fi

TOKEN="$1"
cd "$(dirname "$0")/.."

# 設定 remote 並推送
git remote set-url origin "https://$TOKEN@github.com/gensui-fuga/lanshare.git"
git push -u origin main

# 打 tag 觸發 GitHub Actions
git tag v1.0.0
git push origin v1.0.0

echo ""
echo "✅ 推送完成！"
echo "📦 GitHub Actions 會自動打包全平台 Release"
echo "🔗 https://github.com/gensui-fuga/lanshare"
