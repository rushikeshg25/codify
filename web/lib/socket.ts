import { io, Socket } from "socket.io-client";
import { useEffect, useState } from "react";

/**
 * Connects to a codeground's websocket.
 *
 * The pod rejects unauthenticated handshakes, so this stays idle until both
 * the url and a codeground-scoped access token are available.
 */
const useSocket = (url: string, token?: string) => {
  const [socket, setSocket] = useState<Socket | null>(null);

  useEffect(() => {
    if (!url || !token) return;

    const newSocket = io(url, {
      transports: ["websocket"],
      auth: { token },
      reconnection: true,
      reconnectionAttempts: 5,
      reconnectionDelay: 1000,
    });

    newSocket.on("connect", () => {
      console.log("Socket connected successfully");
      setSocket(newSocket);
    });

    newSocket.on("connect_error", (error) => {
      console.error("Socket connection error:", error);
    });

    return () => {
      if (newSocket) {
        newSocket.disconnect();
        setSocket(null);
      }
    };
  }, [url, token]);

  return socket;
};

export default useSocket;
