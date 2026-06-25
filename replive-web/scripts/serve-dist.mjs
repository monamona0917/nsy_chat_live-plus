#!/usr/bin/env node
import fs from "node:fs";
import http from "node:http";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const projectRoot = path.resolve(__dirname, "..");
const distRoot = path.join(projectRoot, "dist");

const args = parseArgs(process.argv.slice(2));
const host = args.host || process.env.HOST || "0.0.0.0";
const port = Number(args.port || process.env.PORT || 5173);
const apiTarget =
  args.api ||
  process.env.API_PROXY_TARGET ||
  process.env.VITE_API_PROXY_TARGET ||
  "http://127.0.0.1:8888";

if (!fs.existsSync(path.join(distRoot, "index.html"))) {
  console.error("未找到 dist/index.html，请先运行 npm run compile。");
  process.exit(1);
}

const server = http.createServer((req, res) => {
  if (!req.url) {
    sendText(res, 400, "bad request");
    return;
  }

  if (req.url.startsWith("/api/") || req.url.startsWith("/media/")) {
    proxyRequest(req, res);
    return;
  }

  serveStatic(req, res);
});

server.listen(port, host, () => {
  console.log(`replive_web: http://${host === "0.0.0.0" ? "localhost" : host}:${port}/`);
  console.log(`API proxy: ${apiTarget}`);
});

function parseArgs(rawArgs) {
  const parsed = {};
  for (let index = 0; index < rawArgs.length; index++) {
    const arg = rawArgs[index];
    if (arg === "--host") parsed.host = rawArgs[++index];
    if (arg === "--port") parsed.port = rawArgs[++index];
    if (arg === "--api") parsed.api = rawArgs[++index];
  }
  return parsed;
}

function serveStatic(req, res) {
  const requestURL = new URL(req.url, "http://localhost");
  const decodedPath = decodeURIComponent(requestURL.pathname);
  const candidatePath = path.normalize(path.join(distRoot, decodedPath));

  if (!candidatePath.startsWith(distRoot)) {
    sendText(res, 403, "forbidden");
    return;
  }

  let filePath = candidatePath;
  if (fs.existsSync(filePath) && fs.statSync(filePath).isDirectory()) {
    filePath = path.join(filePath, "index.html");
  }
  if (!fs.existsSync(filePath)) {
    filePath = path.join(distRoot, "index.html");
  }

  fs.readFile(filePath, (err, data) => {
    if (err) {
      sendText(res, 500, "read file failed");
      return;
    }

    res.writeHead(200, {
      "Content-Type": contentType(filePath),
      "Cache-Control": isAsset(filePath)
        ? "public, max-age=31536000, immutable"
        : "no-cache",
    });
    res.end(data);
  });
}

function proxyRequest(req, res) {
  const targetURL = new URL(req.url || "/", apiTarget);
  const proxyReq = http.request(
    targetURL,
    {
      method: req.method,
      headers: {
        ...req.headers,
        host: targetURL.host,
      },
    },
    (proxyRes) => {
      res.writeHead(proxyRes.statusCode || 502, proxyRes.headers);
      proxyRes.pipe(res);
    },
  );

  proxyReq.on("error", (err) => {
    sendText(res, 502, `proxy failed: ${err.message}`);
  });

  req.pipe(proxyReq);
}

function sendText(res, statusCode, text) {
  res.writeHead(statusCode, { "Content-Type": "text/plain; charset=utf-8" });
  res.end(text);
}

function contentType(filePath) {
  const ext = path.extname(filePath).toLowerCase();
  const types = {
    ".css": "text/css; charset=utf-8",
    ".gif": "image/gif",
    ".html": "text/html; charset=utf-8",
    ".ico": "image/x-icon",
    ".jpeg": "image/jpeg",
    ".jpg": "image/jpeg",
    ".js": "text/javascript; charset=utf-8",
    ".json": "application/json; charset=utf-8",
    ".png": "image/png",
    ".svg": "image/svg+xml",
    ".webp": "image/webp",
  };
  return types[ext] || "application/octet-stream";
}

function isAsset(filePath) {
  return filePath.includes(`${path.sep}assets${path.sep}`);
}
