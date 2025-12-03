#!/bin/bash

# 数据库表转结构体生成结果比对脚本
# 用途：比对 'go run . db2struct' 和 'gsus db2struct' 两种方式生成的内容是否一致

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置变量
TEMP_DIR="./tmp_compare_db2struct"
OUTPUT_DIR_1="${TEMP_DIR}/go_run_output"
OUTPUT_DIR_2="${TEMP_DIR}/gsus_output"
DIFF_REPORT="${TEMP_DIR}/diff_report.txt"

# 打印带颜色的消息
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 清理函数
cleanup() {
    if [ -d "$TEMP_DIR" ]; then
        print_info "清理临时目录..."
        rm -rf "$TEMP_DIR"
    fi
}

# 检查必要的命令
check_requirements() {
    print_info "检查必要的命令..."

    if ! command -v go &> /dev/null; then
        print_error "未找到 go 命令，请先安装 Go"
        exit 1
    fi

    if ! command -v diff &> /dev/null; then
        print_error "未找到 diff 命令"
        exit 1
    fi

    print_success "所有必要命令已就绪"
}

# 检查配置文件
check_config() {
    print_info "检查配置文件..."

    if [ ! -f ".gsus.yaml" ] && [ ! -f ".gsus/config.yaml" ]; then
        print_error "未找到配置文件 .gsus.yaml 或 .gsus/config.yaml"
        print_info "请先创建配置文件，可以参考 .gsus.example.yaml"
        exit 1
    fi

    print_success "配置文件存在"
}

# 准备临时目录
prepare_dirs() {
    print_info "准备临时目录..."

    # 清理旧的临时目录
    cleanup

    # 创建新的临时目录
    mkdir -p "$OUTPUT_DIR_1"
    mkdir -p "$OUTPUT_DIR_2"

    print_success "临时目录创建完成"
}

# 备份配置文件并修改输出路径
backup_and_modify_config() {
    print_info "备份并修改配置文件..."

    if [ -f ".gsus.yaml" ]; then
        CONFIG_FILE=".gsus.yaml"
    else
        CONFIG_FILE=".gsus/config.yaml"
    fi

    cp "$CONFIG_FILE" "${CONFIG_FILE}.backup"
    print_success "配置文件已备份到 ${CONFIG_FILE}.backup"
}

# 恢复配置文件
restore_config() {
    print_info "恢复配置文件..."

    if [ -f ".gsus.yaml.backup" ]; then
        mv ".gsus.yaml.backup" ".gsus.yaml"
    elif [ -f ".gsus/config.yaml.backup" ]; then
        mv ".gsus/config.yaml.backup" ".gsus/config.yaml"
    fi

    print_success "配置文件已恢复"
}

# 修改配置文件的输出路径
modify_config_path() {
    local new_path=$1

    if [ -f ".gsus.yaml" ]; then
        CONFIG_FILE=".gsus.yaml"
    else
        CONFIG_FILE=".gsus/config.yaml"
    fi

    # 使用 sed 修改 path 配置
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        sed -i '' "s|path:.*|path: $new_path|g" "$CONFIG_FILE"
    else
        # Linux
        sed -i "s|path:.*|path: $new_path|g" "$CONFIG_FILE"
    fi
}

# 执行 go run . db2struct
run_go_run() {
    print_info "执行 'go run . db2struct' 生成代码..."

    modify_config_path "$OUTPUT_DIR_1"

    if go run . db2struct "$@" 2>&1 | tee "${TEMP_DIR}/go_run.log"; then
        print_success "'go run . db2struct' 执行完成"
    else
        print_error "'go run . db2struct' 执行失败"
        cat "${TEMP_DIR}/go_run.log"
        restore_config
        exit 1
    fi
}

# 执行 gsus db2struct
run_gsus() {
    print_info "执行 'gsus db2struct' 生成代码..."

    # 检查 gsus 命令是否存在
    if ! command -v gsus &> /dev/null; then
        print_warning "未找到 gsus 命令，尝试使用 'go install' 安装..."
        go install .

        if ! command -v gsus &> /dev/null; then
            print_error "gsus 命令安装失败，请手动安装"
            restore_config
            exit 1
        fi
    fi

    modify_config_path "$OUTPUT_DIR_2"

    if gsus db2struct "$@" 2>&1 | tee "${TEMP_DIR}/gsus.log"; then
        print_success "'gsus db2struct' 执行完成"
    else
        print_error "'gsus db2struct' 执行失败"
        cat "${TEMP_DIR}/gsus.log"
        restore_config
        exit 1
    fi
}

# 比对生成的文件
compare_outputs() {
    print_info "开始比对生成的文件..."

    # 初始化差异报告
    echo "数据库表转结构体生成结果比对报告" > "$DIFF_REPORT"
    echo "生成时间: $(date)" >> "$DIFF_REPORT"
    echo "========================================" >> "$DIFF_REPORT"
    echo "" >> "$DIFF_REPORT"

    local has_diff=0

    # 检查文件数量
    local count1=$(find "$OUTPUT_DIR_1" -type f -name "*.go" | wc -l)
    local count2=$(find "$OUTPUT_DIR_2" -type f -name "*.go" | wc -l)

    echo "文件数量比对:" >> "$DIFF_REPORT"
    echo "  go run . db2struct: $count1 个文件" >> "$DIFF_REPORT"
    echo "  gsus db2struct: $count2 个文件" >> "$DIFF_REPORT"
    echo "" >> "$DIFF_REPORT"

    if [ "$count1" -ne "$count2" ]; then
        print_warning "生成的文件数量不一致！"
        echo "⚠️  文件数量不一致" >> "$DIFF_REPORT"
        has_diff=1
    else
        print_success "生成的文件数量一致"
        echo "✓ 文件数量一致" >> "$DIFF_REPORT"
    fi
    echo "" >> "$DIFF_REPORT"

    # 比对每个文件
    echo "文件内容比对:" >> "$DIFF_REPORT"
    echo "----------------------------------------" >> "$DIFF_REPORT"

    for file1 in "$OUTPUT_DIR_1"/*.go; do
        if [ ! -f "$file1" ]; then
            continue
        fi

        filename=$(basename "$file1")
        file2="$OUTPUT_DIR_2/$filename"

        if [ ! -f "$file2" ]; then
            print_warning "文件 $filename 仅在 go run 输出中存在"
            echo "⚠️  文件 $filename 仅在 go run 输出中存在" >> "$DIFF_REPORT"
            has_diff=1
            continue
        fi

        # 使用 diff 比对文件
        if diff -u "$file1" "$file2" > "${TEMP_DIR}/${filename}.diff" 2>&1; then
            print_success "文件 $filename 内容一致"
            echo "✓ $filename - 内容一致" >> "$DIFF_REPORT"
        else
            print_error "文件 $filename 内容不一致"
            echo "" >> "$DIFF_REPORT"
            echo "✗ $filename - 内容不一致" >> "$DIFF_REPORT"
            echo "差异详情:" >> "$DIFF_REPORT"
            cat "${TEMP_DIR}/${filename}.diff" >> "$DIFF_REPORT"
            echo "" >> "$DIFF_REPORT"
            has_diff=1
        fi
    done

    # 检查仅在 gsus 输出中存在的文件
    for file2 in "$OUTPUT_DIR_2"/*.go; do
        if [ ! -f "$file2" ]; then
            continue
        fi

        filename=$(basename "$file2")
        file1="$OUTPUT_DIR_1/$filename"

        if [ ! -f "$file1" ]; then
            print_warning "文件 $filename 仅在 gsus 输出中存在"
            echo "⚠️  文件 $filename 仅在 gsus 输出中存在" >> "$DIFF_REPORT"
            has_diff=1
        fi
    done

    echo "" >> "$DIFF_REPORT"
    echo "========================================" >> "$DIFF_REPORT"

    if [ $has_diff -eq 0 ]; then
        echo "结论: ✓ 两种方式生成的内容完全一致" >> "$DIFF_REPORT"
        print_success "两种方式生成的内容完全一致！"
        return 0
    else
        echo "结论: ✗ 两种方式生成的内容存在差异" >> "$DIFF_REPORT"
        print_error "两种方式生成的内容存在差异！"
        return 1
    fi
}

# 显示差异报告
show_report() {
    print_info "差异报告已保存到: $DIFF_REPORT"
    echo ""
    cat "$DIFF_REPORT"
    echo ""

    print_info "详细的生成文件保存在:"
    print_info "  go run 输出: $OUTPUT_DIR_1"
    print_info "  gsus 输出: $OUTPUT_DIR_2"
}

# 主函数
main() {
    echo ""
    print_info "=========================================="
    print_info "数据库表转结构体生成结果比对工具"
    print_info "=========================================="
    echo ""

    # 检查环境
    check_requirements
    check_config

    # 准备环境
    prepare_dirs
    backup_and_modify_config

    # 捕获退出信号，确保恢复配置
    trap restore_config EXIT

    # 执行生成
    run_go_run "$@"
    run_gsus "$@"

    # 恢复配置
    restore_config

    # 比对结果
    if compare_outputs; then
        show_report
        print_success "比对完成！"
        exit 0
    else
        show_report
        print_error "比对完成，发现差异！"
        exit 1
    fi
}

# 显示帮助信息
show_help() {
    cat << EOF
用法: $0 [表名...]

数据库表转结构体生成结果比对工具

参数:
  [表名...]    可选，指定要生成的表名，不指定则生成所有表

示例:
  $0              # 生成所有表并比对
  $0 users        # 仅生成 users 表并比对
  $0 users posts  # 生成 users 和 posts 表并比对

选项:
  -h, --help     显示此帮助信息

注意:
  1. 需要先配置 .gsus.yaml 或 .gsus/config.yaml 文件
  2. 确保数据库连接信息正确
  3. 脚本会自动备份和恢复配置文件
  4. 比对结果保存在 ./tmp_compare_db2struct/ 目录

EOF
}

# 解析命令行参数
if [ "$1" = "-h" ] || [ "$1" = "--help" ]; then
    show_help
    exit 0
fi

# 执行主函数
main "$@"
