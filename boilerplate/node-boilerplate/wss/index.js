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

  socket.on("file:change", async ({ path, content }) => {
    await fs.writeFile(DIR, content);
  });

  socket.on("terminal:write", (data) => {
    ptyProcess.write(data);
  });

  socket.on("terminal:init", () => {
    socket.emit("terminal:data", "\r$ ");
  });
});

app.get("/files", async (req, res) => {
  const fileTree = await generateFileTree(DIR);
  return res.json({ tree: fileTree });
});

app.get("/files/content", async (req, res) => {
  const path = req.query.path;
  const content = await fs.readFile(`${DIR}/${path}`, "utf-8");
  return res.json({ content });
});

server.listen(9000, () => console.log(`server running on port 9000`));

let idCounter = 0;

async function generateFileTree(directory) {
  const tree = {
    name: path.basename(directory),
    _id: idCounter++,
    children: [],
  };

  async function buildTree(currentDir, currentNode) {
    const files = await fs.readdir(currentDir);

    for (const file of files) {
      const filePath = path.join(currentDir, file);
      const stat = await fs.stat(filePath);

      if (stat.isDirectory()) {
        const childNode = { name: file, _id: idCounter++, children: [] };
        currentNode.children.push(childNode);
        await buildTree(filePath, childNode);
      } else {
        currentNode.children.push({ name: file, _id: idCounter++ });
      }
    }
  }

  await buildTree(directory, tree);
  return tree;
}
