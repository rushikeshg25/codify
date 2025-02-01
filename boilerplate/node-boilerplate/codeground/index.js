const express = require("express");

const app = express();

app.get("/", (req, res) => {
  res.send("Hello World!");
});

app.get("/hello", (req, res) => {
  res.send("Hello there");
});

app.listen(8090, () => {
  console.log("Server is running on port 8090");
});
