const jwt = require("jsonwebtoken");

/*
 * Every codeground pod is handed its own derived signing key and its own id.
 * The key is derived by the control plane as HMAC(master, "codeground:<id>"),
 * so this pod can only ever verify tokens minted for itself — compromising one
 * codeground does not yield access to any other.
 *
 * Both values are mandatory. A pod that starts without them would serve an
 * unauthenticated root shell, so refuse to start instead.
 */
const TOKEN_KEY = process.env.CODEGROUND_TOKEN_KEY;
const CODEGROUND_ID = process.env.CODEGROUND_ID;

if (!TOKEN_KEY || !CODEGROUND_ID) {
  console.error(
    "Refusing to start: CODEGROUND_TOKEN_KEY and CODEGROUND_ID must both be set",
  );
  process.exit(1);
}

/** Verifies a codeground access token, throwing if it is not valid for us. */
function verifyToken(token) {
  if (!token) {
    throw new Error("missing token");
  }

  const claims = jwt.verify(token, TOKEN_KEY, {
    algorithms: ["HS256"],
    audience: "codeground",
    issuer: "codify",
  });

  // Defence in depth: the derived key already pins the token to this pod.
  if (claims.codegroundId !== CODEGROUND_ID) {
    throw new Error("token is scoped to a different codeground");
  }

  return claims;
}

/** Pulls the token out of a socket.io handshake (auth payload or query). */
function tokenFromHandshake(socket) {
  return socket.handshake?.auth?.token || socket.handshake?.query?.token;
}

/** Pulls the token out of an Express request (bearer header or query). */
function tokenFromRequest(req) {
  const header = req.get("authorization") || "";
  if (header.toLowerCase().startsWith("bearer ")) {
    return header.slice(7).trim();
  }
  return req.query.token;
}

module.exports = { verifyToken, tokenFromHandshake, tokenFromRequest };
