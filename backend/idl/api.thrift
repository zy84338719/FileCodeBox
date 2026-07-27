// =====================================================================
// api.thrift — hz HTTP 注解约定（thrift 版本）
// =====================================================================
//
// 这个文件不是 IDL 服务定义文件，而是 **约定文档**，用于规范其他
// .thrift IDL 文件如何写 hz 路由注解。
//
// 与 proto 时代不同，thrift 没有单独的 annotation proto 文件，所有
// hz 注解直接内联在 service method 上：
//
//   service FooService {
//       // 注解在 method 末尾，hz 工具会自动识别
//       BarResponse Bar(1: BarRequest req) (api.get = "/foo/bar")
//       //                                ^^^^^^^^^^^^^^^^^^^^^^
//       //                                注解格式: (api.METHOD = "path")
//   }
//
// 路由注解（method 末尾）:
//   (api.get    = "/path")
//   (api.post   = "/path")
//   (api.put    = "/path")
//   (api.delete = "/path")
//   (api.patch  = "/path")
//
// 路径参数用 :param 表示（hz 自动转 {param}）：
//   (api.get = "/admin/files/:id")
//
// 字段不需要 annotation。hz 自动根据 thrift 字段名生成
// json tag。例：i64 file_size → JSON "file_size"。
//
// 如果需要不同 JSON 命名，可以用 Go field tag（hz 不支持），
// 或在 Go 端处理。最稳妥是保持 thrift 字段名 = JSON 名。
//
// =====================================================================
// 命名规范
// =====================================================================
//
// namespace go <service>.<sub>     例：namespace go filecodebox.user
//                                       → Go package: filecodebox/user
//                                       → import path: backend/gen/http/model/user
//
// type 命名：XXXReq / XXXResp（保留 proto 时期的命名）
// service 命名：XxxService（与 proto 1:1）
//
// field 编号必须从 1 开始且保持原 proto 编号（API 兼容要求）。
//
// =====================================================================
