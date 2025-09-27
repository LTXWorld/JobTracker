# 后端测试标签与执行指南

> 说明当前后端在解耦重构后的测试分层方式，以及 `run_tests.sh` 后续需补充的脚本参数规划，便于提交前校验。

## 测试标签划分

- `integration`：
  - 覆盖 handler、数据库监控、API 等需要真实依赖的测试。
  - 默认 `go test ./...` 不会执行；需显式运行 `go test -tags integration ./...`。
- `loadtest`：
  - 大量并发与性能验证（`tests/service/job_application_load_test.go`、`job_application_regression_test.go`）。
  - 推荐在资源充足的环境中执行：`go test -tags loadtest ./tests/service -run LoadTest`。

## 占位测试说明

- 对于被标签排除的目录，新增了占位测试（`tests/*/placeholder_test.go`）。
- 目的是保持 `go test ./...` 通过，同时提示执行者需要启用对应标签。

## 脚本优化规划

- `run_tests.sh` 后续可增加参数：
  1. `--integration`：串行调用 `go test -tags integration ./...`。
  2. `--loadtest`：触发性能/压测标签，并允许配置运行时间、并发等环境变量。
- 脚本输出中建议新增对占位提示的说明，避免误以为未执行测试。

## 验证建议

- 提交前默认流程：`GOCACHE=$(pwd)/.gocache go test ./...`。
- 重要需求上线前，请额外运行：
  - `go test -tags integration ./tests/auth ./tests/handler ./tests/api`。
  - `go test -tags loadtest ./tests/service -run BenchmarkLoadTest_Quick -bench .`。

## 待办摘要

1. 把监控相关逻辑抽象成 `MonitoringService`，供 handler 依赖。
2. 更新 `run_tests.sh` 以支持可选标签参数。
3. 在 CI 流程中补充 `integration` 标签的一轮冒烟测试。
