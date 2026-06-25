## 编译与运行

普通 Windows 用户不需要安装 Node/npm/Bun。发布时用仓库根目录的脚本打包：

```powershell
.\scripts\build-windows.ps1
```

脚本会生成：

- `dist/replive.exe`：后端程序
- `dist/replive_web.exe`：前端程序，内置已编译页面
- `dist/使用说明.txt`

用户双击 `replive.exe` 后，再双击 `replive_web.exe` 即可。

开发环境也可以只运行前端静态产物：

```bash
npm install --package-lock=false
npm run compile
npm run run -- --port 5173 --api http://127.0.0.1:8888
```

`npm run run` 只使用 Node 标准库托管 `dist/`，不依赖 Vite dev server。`/api/*` 和 `/media/*` 会代理到 `--api` 指定的后端地址。
