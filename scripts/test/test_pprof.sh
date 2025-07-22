#!/bin/bash

# Gateway pprof性能分析服务测试脚本
# 注意：pprof服务已集成到主应用中，请先启动主应用

set -e

echo "🚀 Gateway pprof测试脚本"
echo "======================="
echo "⚠️  注意：请确保Gateway主应用已启动"
echo ""

# 配置
PPROF_HOST="localhost:6060"
OUTPUT_DIR="./pprof_test_output"
SAMPLE_TIME="10"

# 检查go工具是否可用
if ! command -v go &> /dev/null; then
    echo "❌ Go工具未找到，请先安装Go"
    exit 1
fi

# 创建输出目录
mkdir -p "$OUTPUT_DIR"

echo "📋 测试配置:"
echo "  - 目标服务: $PPROF_HOST"
echo "  - 输出目录: $OUTPUT_DIR"
echo "  - 采样时间: ${SAMPLE_TIME}s"
echo ""

# 测试函数
test_service_health() {
    echo "🔍 测试1: 健康检查"
    if curl -s -f "http://$PPROF_HOST/health" > /dev/null; then
        echo "  ✅ 健康检查通过"
    else
        echo "  ❌ 健康检查失败"
        return 1
    fi
}

test_service_info() {
    echo "🔍 测试2: 服务信息"
    if curl -s -f "http://$PPROF_HOST/info" > /dev/null; then
        echo "  ✅ 服务信息获取成功"
        echo "  📊 服务信息:"
        curl -s "http://$PPROF_HOST/info" | jq '.' 2>/dev/null || curl -s "http://$PPROF_HOST/info"
    else
        echo "  ❌ 服务信息获取失败"
        return 1
    fi
}

test_pprof_endpoints() {
    echo "🔍 测试3: pprof端点"
    
    local endpoints=(
        "profile?seconds=$SAMPLE_TIME"
        "heap"
        "goroutine"
        "allocs"
        "block"
        "mutex"
        "threadcreate"
    )
    
    for endpoint in "${endpoints[@]}"; do
        echo "  📊 测试端点: $endpoint"
        if curl -s -f "http://$PPROF_HOST/debug/pprof/$endpoint" -o "$OUTPUT_DIR/${endpoint//[?=]/}.prof"; then
            echo "    ✅ $endpoint 数据收集成功"
        else
            echo "    ❌ $endpoint 数据收集失败"
        fi
    done
}

test_pprof_analysis() {
    echo "🔍 测试4: pprof分析"
    
    local profiles=(
        "profile?seconds=$SAMPLE_TIME.prof"
        "heap.prof"
        "goroutine.prof"
    )
    
    for profile in "${profiles[@]}"; do
        local profile_file="$OUTPUT_DIR/${profile//[?=]/}"
        if [[ -f "$profile_file" ]]; then
            echo "  📊 分析文件: $profile"
            local output_file="$OUTPUT_DIR/${profile//[?=.]/}_analysis.txt"
            
            if go tool pprof -top "$profile_file" > "$output_file" 2>/dev/null; then
                echo "    ✅ ${profile} 分析完成"
                echo "    📄 报告: $output_file"
            else
                echo "    ❌ ${profile} 分析失败"
            fi
        fi
    done
}

test_web_interface() {
    echo "🔍 测试5: Web界面"
    if curl -s -f "http://$PPROF_HOST/debug/pprof/" > /dev/null; then
        echo "  ✅ Web界面可访问"
        echo "  🌐 访问地址: http://$PPROF_HOST/debug/pprof/"
    else
        echo "  ❌ Web界面访问失败"
        return 1
    fi
}

test_manual_analysis() {
    echo "🔍 测试6: 手动分析触发"
    if curl -s -X POST "http://$PPROF_HOST/analyze" > /dev/null; then
        echo "  ✅ 手动分析触发成功"
    else
        echo "  ❌ 手动分析触发失败"
        return 1
    fi
}

show_usage_examples() {
    echo ""
    echo "📚 使用示例:"
    echo "============"
    echo ""
    echo "1. 启动Gateway主应用:"
    echo "   go run cmd/app/main.go"
    echo ""
    echo "2. CPU分析:"
    echo "   go tool pprof http://$PPROF_HOST/debug/pprof/profile?seconds=30"
    echo ""
    echo "3. 内存分析:"
    echo "   go tool pprof http://$PPROF_HOST/debug/pprof/heap"
    echo ""
    echo "4. 协程分析:"
    echo "   go tool pprof http://$PPROF_HOST/debug/pprof/goroutine"
    echo ""
    echo "5. 生成火焰图:"
    echo "   go tool pprof -http=:8080 http://$PPROF_HOST/debug/pprof/profile?seconds=30"
    echo ""
    echo "6. Web界面:"
    echo "   浏览器打开 http://$PPROF_HOST/debug/pprof/"
    echo ""
    echo "7. 健康检查:"
    echo "   curl http://$PPROF_HOST/health"
    echo ""
    echo "8. 服务信息:"
    echo "   curl http://$PPROF_HOST/info"
    echo ""
    echo "9. 配置pprof (在configs/app.yaml中):"
    echo "   app:"
    echo "     pprof:"
    echo "       enabled: true"
    echo "       listen: \":6060\""
    echo "       auto_analysis:"
    echo "         enabled: true"
    echo ""
}

cleanup() {
    echo ""
    echo "🧹 清理测试数据..."
    if [[ -d "$OUTPUT_DIR" ]]; then
        rm -rf "$OUTPUT_DIR"
        echo "  ✅ 清理完成"
    fi
}

# 主测试流程
main() {
    local failed_tests=0
    
    # 检查服务是否运行
    echo "🔍 检查pprof服务状态..."
    if ! curl -s -f "http://$PPROF_HOST/health" > /dev/null; then
        echo "❌ pprof服务未运行或无法访问"
        echo "请确保Gateway主应用正在运行"
        echo "启动命令: go run cmd/app/main.go"
        echo "确认配置: configs/app.yaml 中 app.pprof.enabled: true"
        exit 1
    fi
    
    echo "✅ pprof服务正在运行"
    echo ""
    
    # 运行测试
    test_service_health || ((failed_tests++))
    echo ""
    
    test_service_info || ((failed_tests++))
    echo ""
    
    test_pprof_endpoints || ((failed_tests++))
    echo ""
    
    test_pprof_analysis || ((failed_tests++))
    echo ""
    
    test_web_interface || ((failed_tests++))
    echo ""
    
    test_manual_analysis || ((failed_tests++))
    echo ""
    
    # 显示测试结果
    echo "📊 测试结果:"
    echo "============"
    if [[ $failed_tests -eq 0 ]]; then
        echo "✅ 所有测试通过! pprof服务工作正常"
    else
        echo "❌ $failed_tests 个测试失败"
    fi
    
    # 显示使用示例
    show_usage_examples
    
    # 询问是否清理
    echo ""
    read -p "是否删除测试输出文件? (y/n): " -n 1 -r
    echo ""
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        cleanup
    else
        echo "📁 测试文件保留在: $OUTPUT_DIR"
    fi
    
    return $failed_tests
}

# 处理中断信号
trap cleanup EXIT

# 运行主函数
main "$@" 