#!/usr/bin/env python3
"""
FileCodeBox MCP Server 测试客户端

这个脚本演示了如何连接和使用 FileCodeBox MCP 服务器的基本功能。
"""

import json
import socket
import sys
import time
from typing import Any, Dict, Optional


class MCPClient:
    """简单的 MCP 客户端实现"""
    
    def __init__(self, host: str = "localhost", port: int = 8081):
        self.host = host
        self.port = port
        self.sock: Optional[socket.socket] = None
        self.request_id = 0
    
    def connect(self) -> bool:
        """连接到 MCP 服务器"""
        try:
            self.sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            self.sock.connect((self.host, self.port))
            print(f"✅ 已连接到 MCP 服务器 {self.host}:{self.port}")
            return True
        except Exception as e:
            print(f"❌ 连接失败: {e}")
            return False
    
    def disconnect(self):
        """断开连接"""
        if self.sock:
            self.sock.close()
            self.sock = None
            print("🔌 已断开连接")
    
    def send_request(self, method: str, params: Dict[str, Any] = None) -> Dict[str, Any]:
        """发送 JSON-RPC 请求"""
        if not self.sock:
            raise Exception("未连接到服务器")
        
        self.request_id += 1
        request = {
            "jsonrpc": "2.0",
            "id": self.request_id,
            "method": method,
            "params": params or {}
        }
        
        # 发送请求
        message = json.dumps(request) + '\n'
        self.sock.send(message.encode('utf-8'))
        
        # 接收响应
        response_data = self.sock.recv(4096).decode('utf-8')
        
        if not response_data:
            raise Exception("服务器关闭了连接")
        
        # 解析响应（可能包含多行）
        lines = response_data.strip().split('\n')
        for line in lines:
            if line.strip():
                try:
                    response = json.loads(line)
                    if response.get('id') == self.request_id:
                        return response
                except json.JSONDecodeError:
                    continue
        
        raise Exception("未收到有效响应")
    
    def initialize(self) -> bool:
        """初始化 MCP 连接"""
        try:
            response = self.send_request("initialize", {
                "protocolVersion": "2024-11-05",
                "capabilities": {
                    "roots": {"listChanged": True},
                    "sampling": {}
                },
                "clientInfo": {
                    "name": "filecodebox-test-client",
                    "version": "1.0.0"
                }
            })
            
            if "result" in response:
                print("✅ MCP 初始化成功")
                return True
            else:
                print(f"❌ 初始化失败: {response.get('error', '未知错误')}")
                return False
        except Exception as e:
            print(f"❌ 初始化异常: {e}")
            return False
    
    def call_tool(self, tool_name: str, arguments: Dict[str, Any] = None) -> Dict[str, Any]:
        """调用 MCP 工具"""
        try:
            response = self.send_request("tools/call", {
                "name": tool_name,
                "arguments": arguments or {}
            })
            return response
        except Exception as e:
            print(f"❌ 工具调用异常: {e}")
            return {"error": str(e)}
    
    def list_tools(self) -> Dict[str, Any]:
        """列出可用工具"""
        try:
            response = self.send_request("tools/list")
            return response
        except Exception as e:
            print(f"❌ 列出工具异常: {e}")
            return {"error": str(e)}


def test_basic_functionality(client: MCPClient):
    """测试基本功能"""
    print("\n🧪 测试基本功能...")
    
    # 1. 列出工具
    print("\n📋 列出可用工具...")
    tools_response = client.list_tools()
    if "result" in tools_response:
        tools = tools_response["result"].get("tools", [])
        print(f"可用工具数量: {len(tools)}")
        for tool in tools[:3]:  # 只显示前3个
            print(f"  - {tool.get('name', '未知')}: {tool.get('description', '无描述')}")
    else:
        print(f"获取工具列表失败: {tools_response.get('error', '未知错误')}")
    
    # 2. 获取系统状态
    print("\n📊 获取系统状态...")
    status_response = client.call_tool("get_system_status")
    if "result" in status_response:
        print("✅ 系统状态获取成功")
        content = status_response["result"].get("content", [])
        if content and len(content) > 0:
            text_content = content[0].get("text", "")
            # 只显示前几行
            lines = text_content.split('\n')[:5]
            for line in lines:
                if line.strip():
                    print(f"  {line}")
    else:
        print(f"获取系统状态失败: {status_response.get('error', '未知错误')}")
    
    # 3. 创建文本分享
    print("\n📝 创建文本分享...")
    test_text = f"MCP 测试分享 - {time.strftime('%Y-%m-%d %H:%M:%S')}"
    share_response = client.call_tool("share_text", {
        "text": test_text,
        "expire_value": 1,
        "expire_style": "day"
    })
    
    share_code = None
    if "result" in share_response:
        print("✅ 文本分享创建成功")
        content = share_response["result"].get("content", [])
        if content and len(content) > 0:
            text_content = content[0].get("text", "")
            # 从响应中提取分享代码
            for line in text_content.split('\n'):
                if "分享代码:" in line or "代码:" in line:
                    parts = line.split(":")
                    if len(parts) > 1:
                        share_code = parts[1].strip()
                        print(f"  分享代码: {share_code}")
                        break
    else:
        print(f"创建文本分享失败: {share_response.get('error', '未知错误')}")
    
    # 4. 如果成功创建了分享，尝试获取它
    if share_code:
        print(f"\n🔍 获取分享信息 (代码: {share_code})...")
        get_response = client.call_tool("get_share", {"code": share_code})
        if "result" in get_response:
            print("✅ 获取分享信息成功")
            is_error = get_response["result"].get("isError", False)
            if not is_error:
                print("  分享信息验证通过")
            else:
                print("  分享信息获取出错")
        else:
            print(f"获取分享信息失败: {get_response.get('error', '未知错误')}")
        
        # 5. 清理：删除测试分享
        print(f"\n🗑️  清理测试分享 (代码: {share_code})...")
        delete_response = client.call_tool("delete_share", {"code": share_code})
        if "result" in delete_response:
            is_error = delete_response["result"].get("isError", False)
            if not is_error:
                print("✅ 测试分享删除成功")
            else:
                print("⚠️  删除分享时出现错误")
        else:
            print(f"删除分享失败: {delete_response.get('error', '未知错误')}")


def test_storage_info(client: MCPClient):
    """测试存储信息功能"""
    print("\n💾 测试存储信息...")
    storage_response = client.call_tool("get_storage_info")
    if "result" in storage_response:
        print("✅ 存储信息获取成功")
        is_error = storage_response["result"].get("isError", False)
        if not is_error:
            content = storage_response["result"].get("content", [])
            if content and len(content) > 0:
                text_content = content[0].get("text", "")
                lines = text_content.split('\n')[:3]  # 只显示前几行
                for line in lines:
                    if line.strip():
                        print(f"  {line}")
        else:
            print("  存储信息获取出错")
    else:
        print(f"获取存储信息失败: {storage_response.get('error', '未知错误')}")


def main():
    """主函数"""
    print("🚀 FileCodeBox MCP Server 测试客户端")
    print("=" * 50)
    
    if len(sys.argv) > 1:
        if sys.argv[1] in ["-h", "--help", "help"]:
            print("使用方法:")
            print("  python3 test_mcp_client.py [host] [port]")
            print("")
            print("参数:")
            print("  host    MCP 服务器地址 (默认: localhost)")
            print("  port    MCP 服务器端口 (默认: 8081)")
            print("")
            print("示例:")
            print("  python3 test_mcp_client.py")
            print("  python3 test_mcp_client.py localhost 8081")
            return
    
    # 解析命令行参数
    host = sys.argv[1] if len(sys.argv) > 1 else "localhost"
    port = int(sys.argv[2]) if len(sys.argv) > 2 else 8081
    
    # 创建客户端并连接
    client = MCPClient(host, port)
    
    try:
        # 连接和初始化
        if not client.connect():
            return
        
        if not client.initialize():
            return
        
        # 运行测试
        test_basic_functionality(client)
        test_storage_info(client)
        
        print("\n✅ 所有测试完成!")
        print("\n💡 提示: 确保 FileCodeBox 应用已启动并设置了以下环境变量:")
        print("   export ENABLE_MCP_SERVER=true")
        print("   export MCP_PORT=8081")
        
    except KeyboardInterrupt:
        print("\n\n⏹️  用户中断测试")
    except Exception as e:
        print(f"\n❌ 测试过程中出现异常: {e}")
    finally:
        client.disconnect()


if __name__ == "__main__":
    main()
