const http = require("http");
const express = require("express");
const fs = require("fs/promises");
const { Server: SocketServer } = require("socket.io");
const path = require("path");
const cors = require("cors");
const chokidar = require("chokidar");
const DIR = path.resolve(__dirname, "../codeground");
const pty = require("node-pty");

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
  cors: "*",
});

app.use(cors());

io.attach(server);

chokidar.watch(DIR).on("all", (event, path) => {
  io.emit("file:refresh", path);
});

ptyProcess.onData((data) => {
  io.emit("terminal:data", data);
});

io.on("connection", (socket) => {
  console.log(`Socket connected`, socket.id);
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

app.get("/files", async (req, res) => {
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

app.get("/files/content", async (req, res) => {
  const file = req.query.file;
  console.log(fileMap.get(file));
  const content = await fs.readFile(fileMap.get(file), "utf-8");
  return res.json({ content });
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
