#!/bin/bash

# FileCodeBox 测试套件运行器
# 自动运行所有测试脚本并生成报告

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BASE_URL="http://localhost:12345"
TIMESTAMP=$(date '+%Y%m%d_%H%M%S')
REPORT_FILE="test_report_${TIMESTAMP}.txt"

echo "=== FileCodeBox 测试套件运行器 ===" | tee "$REPORT_FILE"
echo "开始时间: $(date)" | tee -a "$REPORT_FILE"
echo "测试目录: $SCRIPT_DIR" | tee -a "$REPORT_FILE"
echo | tee -a "$REPORT_FILE"

# 检查服务器状态
echo "检查服务器状态..." | tee -a "$REPORT_FILE"
if ! curl -s --connect-timeout 3 "$BASE_URL" > /dev/null 2>&1; then
    echo "❌ 服务器未运行在 $BASE_URL" | tee -a "$REPORT_FILE"
    echo "请先启动 FileCodeBox 服务器，然后重新运行测试" | tee -a "$REPORT_FILE"
    exit 1
fi
echo "✅ 服务器运行正常" | tee -a "$REPORT_FILE"
echo | tee -a "$REPORT_FILE"

# 测试脚本分类 - 使用兼容性更好的方法
get_test_scripts() {
    local category=$1
    case "$category" in
        "core")
            echo "test_api.sh test_admin.sh test_chunk.sh"
            ;;
        "storage")
            echo "test_storage_management.sh test_webdav_config.sh"
            ;;
        "database")
            echo "test_database_config.sh test_date_grouping.sh"
            ;;
        "frontend")
            echo "test_web.sh test_ui_features.sh test_javascript.sh test_progress.sh"
            ;;
        "issues")
            echo "test_upload_limit.sh"
            ;;
        "performance")
            echo "benchmark.sh"
            ;;
        "resume")
            echo "test_resume_upload.sh"
            ;;
        "basic")
            echo "simple_test.sh"
            ;;
        *)
            echo ""
            ;;
    esac
}

# 运行测试分类
run_category_tests() {
    local category=$1
    local scripts=$2
    local category_upper=$(echo "$category" | tr '[:lower:]' '[:upper:]')
    
    echo "=== 📂 ${category_upper} 测试 ===" | tee -a "$REPORT_FILE"
    echo "测试脚本: $scripts" | tee -a "$REPORT_FILE"
    echo | tee -a "$REPORT_FILE"
    
    local passed=0
    local failed=0
    
    for script in $scripts; do
        if [[ -f "$SCRIPT_DIR/$script" ]]; then
            echo "🔄 运行: $script" | tee -a "$REPORT_FILE"
            echo "开始时间: $(date)" | tee -a "$REPORT_FILE"
            
            # 运行测试脚本
            if timeout 60 "$SCRIPT_DIR/$script" >> "$REPORT_FILE" 2>&1; then
                echo "✅ $script - 通过" | tee -a "$REPORT_FILE"
                ((passed++))
            else
                echo "❌ $script - 失败" | tee -a "$REPORT_FILE"
                ((failed++))
            fi
            echo "结束时间: $(date)" | tee -a "$REPORT_FILE"
            echo "----------------------------------------" | tee -a "$REPORT_FILE"
        else
            echo "⚠️  脚本不存在: $script" | tee -a "$REPORT_FILE"
            ((failed++))
        fi
    done
    
    echo "📊 ${category_upper} 测试结果: 通过 $passed, 失败 $failed" | tee -a "$REPORT_FILE"
    echo | tee -a "$REPORT_FILE"
    
    return $failed
}

# 主测试流程
main() {
    local total_passed=0
    local total_failed=0
    
    echo "开始运行测试套件..." | tee -a "$REPORT_FILE"
    echo | tee -a "$REPORT_FILE"
    
    # 按分类运行测试
    for category in core storage database frontend issues performance resume basic; do
        local scripts=$(get_test_scripts "$category")
        if run_category_tests "$category" "$scripts"; then
            echo "✅ $category 测试分类通过" | tee -a "$REPORT_FILE"
        else
            echo "❌ $category 测试分类有失败项" | tee -a "$REPORT_FILE"
            ((total_failed++))
        fi
        echo | tee -a "$REPORT_FILE"
    done
    
    # 生成总结报告
    echo "=== 📋 测试总结报告 ===" | tee -a "$REPORT_FILE"
    echo "总测试时间: $(date)" | tee -a "$REPORT_FILE"
    echo "报告文件: $REPORT_FILE" | tee -a "$REPORT_FILE"
    
    # 统计所有脚本执行情况
    local total_scripts=$(find "$SCRIPT_DIR" -name "*.sh" ! -name "run_all_tests.sh" | wc -l)
    echo "总测试脚本数: $total_scripts" | tee -a "$REPORT_FILE"
    
    if [[ $total_failed -eq 0 ]]; then
        echo "🎉 所有测试分类都通过了！" | tee -a "$REPORT_FILE"
        exit 0
    else
        echo "⚠️  有 $total_failed 个测试分类包含失败项" | tee -a "$REPORT_FILE"
        echo "详细信息请查看报告文件: $REPORT_FILE" | tee -a "$REPORT_FILE"
        exit 1
    fi
}

# 处理命令行参数
case "${1:-all}" in
    "core")
        scripts=$(get_test_scripts "core")
        run_category_tests "core" "$scripts"
        ;;
    "storage")
        scripts=$(get_test_scripts "storage")
        run_category_tests "storage" "$scripts"
        ;;
    "frontend")
        scripts=$(get_test_scripts "frontend")
        run_category_tests "frontend" "$scripts"
        ;;
    "performance")
        scripts=$(get_test_scripts "performance")
        run_category_tests "performance" "$scripts"
        ;;
    "resume")
        scripts=$(get_test_scripts "resume")
        run_category_tests "resume" "$scripts"
        ;;
    "all"|*)
        main
        ;;
esac
