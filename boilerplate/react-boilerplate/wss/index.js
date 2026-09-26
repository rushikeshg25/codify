const http = require("http");
const express = require("express");
const fs = require("fs/promises");
const { Server: SocketServer } = require("socket.io");
const path = require("path");
const cors = require("cors");
const chokidar = require("chokidar");
const DIR = path.resolve(__dirname, "../codeground");
const pty = require("node-pty");
const {
  verifyToken,
  tokenFromHandshake,
  tokenFromRequest,
} = require("./auth");

// The browser origin allowed to reach this pod. Defaults to the local dev
// frontend rather than "*", so a hostile page cannot drive someone's terminal.
const ALLOWED_ORIGIN = process.env.CORS_ALLOW_ORIGIN || "http://localhost:3000";

const ptyProcess = pty.spawn("bash", [], {
  name: "xterm-color",
  cols: 80,
  rows: 30,
  cwd: DIR,
  env: process.env,
});
const fileMap = new Map();
const app = express();
const server = http.createServer(app);
const io = new SocketServer({
  cors: { origin: ALLOWED_ORIGIN, credentials: true },
});

app.use(cors({ origin: ALLOWED_ORIGIN, credentials: true }));

/*
 * Nothing below this line is safe to expose unauthenticated: the socket hands
 * out a pty running as the container user, and the file routes read and write
 * the project tree. Reject the connection before any of that is wired up.
 */
/** Express guard mirroring the socket check above. */
function requireToken(req, res, next) {
  try {
    req.claims = verifyToken(tokenFromRequest(req));
    next();
  } catch (err) {
    console.warn("Rejected request to", req.path, "-", err.message);
    res.status(401).json({ error: "unauthorized" });
  }
}

io.use((socket, next) => {
  try {
    socket.data.claims = verifyToken(tokenFromHandshake(socket));
    next();
  } catch (err) {
    console.warn("Rejected socket connection:", err.message);
    next(new Error("unauthorized"));
  }
});

io.attach(server);

chokidar.watch(DIR).on("all", (event, path) => {
  io.emit("file:refresh", path);
});

ptyProcess.onData((data) => {
  io.emit("terminal:data", data);
});

io.on("connection", (socket) => {
  console.log(`Socket connected`, socket.id, "user", socket.data.claims.userId);
  console.log(`Total connected clients: ${io.engine.clientsCount}`);
  socket.emit("file:refresh");

  socket.on("file:change", async ({ file, content }) => {
    await fs.writeFile(fileMap.get(file), content);
  });

  socket.on("terminal:write", (data) => {
    ptyProcess.write(data);
  });

  socket.on("terminal:init", () => {
    socket.emit("terminal:data", "\r$ ");
  });
});

app.get("/files", requireToken, async (req, res) => {
  try {
    const fileTree = await generateFileTree(DIR, fileMap);
    return res.json({ tree: fileTree });
  } catch (error) {
    console.error("Error generating file tree:", error);
    return res.status(500).json({
      error: "Failed to generate file tree",
      details: error.message,
    });
  }
});

app.get("/files/content", requireToken, async (req, res) => {
  const file = req.query.file;
  const resolved = fileMap.get(file);
  if (!resolved) {
    return res.status(404).json({ error: "File not found" });
  }
  try {
    const content = await fs.readFile(resolved, "utf-8");
    return res.json({ content });
  } catch (error) {
    console.error("Failed to read file:", error);
    return res.status(500).json({ error: "Failed to read file" });
  }
});

server.listen(9000, () => console.log(`server running on port 9000`));

async function generateFileTree(directory, fileMap) {
  const tree = {};

  async function buildTree(currentDir, currentTree) {
    const files = await fs.readdir(currentDir);

    for (const file of files) {
      const filePath = path.join(currentDir, file);
      const stat = await fs.stat(filePath);

      if (stat.isDirectory()) {
        currentTree[file] = {};
        await buildTree(filePath, currentTree[file]);
      } else {
        if (filePath.includes("node_modules")) {
          continue;
        } else {
          currentTree[file] = null;
        }
        fileMap.set(file, filePath);
      }
    }
  }

  await buildTree(directory, tree);
  return tree;
}
